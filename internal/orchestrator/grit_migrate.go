package orchestrator

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/docker"
)

// GritMigrateSpec configures the Grit migration hook.
type GritMigrateSpec struct {
	OrgSlug         string            // org network to attach to
	RepoURL         string            // clone URL (a token is embedded below if needed)
	Branch          string            // branch to migrate from
	APIContext      string            // API build context, e.g. "apps/api" ("" = repo root for single)
	DatabaseURL     string            // target DB connection string (reachable in-network)
	Env             map[string]string // full production env for the migrator
	GitConnectionID string            // org git connection, for private-repo clone auth
	OrgID           string            // owning org (uuid) — needed to resolve the token
}

// RunGritMigrations runs a Grit app's migrator once, before cutover, holding a
// Postgres advisory lock for the whole run so overlapping deploys can't race
// (grit-knowledge/06). It runs a one-off golang container on the org network
// that clones the repo, acquires the lock in a single psql session, and invokes
// `go run ./cmd/migrate` under it — the lock is released when that session ends.
// Returns the migrator's combined logs. A non-zero exit is an error, and the
// caller must NOT cut over.
func (o *Orchestrator) RunGritMigrations(ctx context.Context, spec GritMigrateSpec) (string, error) {
	if spec.DatabaseURL == "" {
		return "", fmt.Errorf("RunGritMigrations: DATABASE_URL is required")
	}

	apiDir := "/src"
	if spec.APIContext != "" {
		apiDir = "/src/" + strings.TrimLeft(spec.APIContext, "/")
	}

	// Embed the git token so the migrator can clone a PRIVATE repo, exactly as
	// buildFromGit does for the image build. Without this the clone on the next
	// line fails with "could not read Username for 'https://github.com'" (git
	// exit 128) and the deploy aborts before cutover — the build would have
	// succeeded but migrations could not fetch the same source.
	cloneURL := spec.RepoURL
	if o.tokenFetcher != nil && spec.GitConnectionID != "" && spec.OrgID != "" {
		if _, token, _, err := o.tokenFetcher.ResolveGitToken(ctx, spec.GitConnectionID, spec.OrgID); err != nil {
			log.Warn().Err(err).Msg("RunGritMigrations: token resolve failed; attempting public clone")
		} else if token != "" {
			if parsed, perr := url.Parse(cloneURL); perr == nil {
				parsed.User = url.UserPassword("x-access-token", token)
				cloneURL = parsed.String()
			}
		}
	}

	// One psql session holds the advisory lock while `\!` runs the migrator, so
	// exactly one migrate runs at a time per DB. A sentinel file surfaces the
	// migrator's exit code (psql's \! status isn't propagated reliably).
	script := strings.Join([]string{
		"set -e",
		"apk add --no-cache git postgresql-client >/dev/null 2>&1",
		fmt.Sprintf("git clone --depth 1 --branch %q %q /src >/dev/null 2>&1", spec.Branch, cloneURL),
		fmt.Sprintf("cd %q", apiDir),
		"rm -f /tmp/migrate.ok",
		// -q quiet; hold session-level advisory lock, then run migrator via \!.
		`psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -q ` +
			`-c "SELECT pg_advisory_lock(hashtext('grit_migrate'))" ` +
			`-c "\! (go run ./cmd/migrate && touch /tmp/migrate.ok)"`,
		// Fail the container if the migrator didn't complete.
		"test -f /tmp/migrate.ok",
	}, "\n")

	env := map[string]string{}
	for k, v := range spec.Env {
		env[k] = v
	}
	env["DATABASE_URL"] = spec.DatabaseURL

	oneOff := docker.OneOffSpec{
		// 1.25 covers Grit apps whose deps require a newer toolchain than the
		// 1.24 the shipped Dockerfiles pin; go's GOTOOLCHAIN would otherwise
		// try (and fail offline) to upgrade.
		Image:       "golang:1.25-alpine",
		Command:     []string{"sh", "-c", script},
		EnvVars:     env,
		NetworkName: docker.GetOrgNetworkName(spec.OrgSlug),
		Labels: map[string]string{
			"orbita.grit.migrate": "true",
			"orbita.org":          spec.OrgSlug,
			"orbita.managed":      "true",
		},
		MaxLogBytes: 200 * 1024,
	}

	exitCode, logs, err := o.dockerClient.RunOneOffContainer(ctx, oneOff)
	if err != nil {
		return logs, fmt.Errorf("RunGritMigrations: %w", err)
	}
	if exitCode != 0 {
		return logs, fmt.Errorf("RunGritMigrations: migrator exited %d (deploy aborted, not cutting over): %s", exitCode, tailLog(logs, 600))
	}
	return logs, nil
}

func tailLog(s string, n int) string {
	if len(s) > n {
		return "..." + s[len(s)-n:]
	}
	return s
}
