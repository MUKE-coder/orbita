package orchestrator

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/docker"
	"github.com/orbita-sh/orbita/internal/models"
)

// Pinned so builds are reproducible; bump deliberately. musl build so it runs
// in the alpine-based docker:cli builder.
const nixpacksVersion = "1.41.0"

// buildWithNixpacks builds a git repo that has no Dockerfile into an image by
// running Nixpacks in a one-off builder container. Nixpacks auto-detects the
// language (Node, Python, Go, Ruby, PHP, static, …) and produces a runnable
// image — the "deploy any app" path.
//
// The builder is a docker:cli container with the host Docker socket bound in,
// so `nixpacks build` (which shells out to `docker build`) produces the image
// on the host daemon, exactly where DeployApplication expects to find it. This
// mirrors buildFromGit's token handling and tag naming so the rest of the
// deploy pipeline is unchanged.
func (o *Orchestrator) buildWithNixpacks(ctx context.Context, app *models.Application, deployment *models.Deployment, srcCfg *SourceConfig, orgSlug string) (string, error) {
	if srcCfg.RepoFullName == "" {
		return "", fmt.Errorf("buildWithNixpacks: repo_full_name missing")
	}
	if srcCfg.Branch == "" {
		return "", fmt.Errorf("buildWithNixpacks: branch missing")
	}

	// Same private-repo auth as the Dockerfile path.
	token := ""
	if o.tokenFetcher != nil && srcCfg.GitConnectionID != "" {
		if _, t, _, err := o.tokenFetcher.ResolveGitToken(ctx, srcCfg.GitConnectionID, app.OrganizationID.String()); err != nil {
			log.Warn().Err(err).Msg("buildWithNixpacks: token resolve failed; attempting public clone")
		} else {
			token = t
		}
	}
	cloneURL := srcCfg.RepoURL
	if cloneURL == "" {
		cloneURL = fmt.Sprintf("https://github.com/%s.git", srcCfg.RepoFullName)
	}
	if token != "" {
		parsed, err := url.Parse(cloneURL)
		if err != nil {
			return "", fmt.Errorf("buildWithNixpacks: parse url: %w", err)
		}
		parsed.User = url.UserPassword("x-access-token", token)
		cloneURL = parsed.String()
	}

	tag := fmt.Sprintf("orbita-%s-%s:v%d", orgSlug, app.ID.String()[:8], deployment.Version)

	// Build directory inside the builder: repo root, or a sub-path if a build
	// context was set.
	buildDir := "/src"
	if srcCfg.BuildContext != "" {
		buildDir = "/src/" + strings.TrimLeft(srcCfg.BuildContext, "/")
	}

	nixURL := fmt.Sprintf(
		"https://github.com/railwayapp/nixpacks/releases/download/v%s/nixpacks-v%s-x86_64-unknown-linux-musl.tar.gz",
		nixpacksVersion, nixpacksVersion)

	// The build args become --env flags so things like NEXT_PUBLIC_* are baked in.
	envFlags := ""
	for k, v := range srcCfg.BuildArgs {
		envFlags += fmt.Sprintf(" --env %s=%s", shellQuote(k), shellQuote(v))
	}

	script := strings.Join([]string{
		"set -e",
		"apk add --no-cache curl git tar >/dev/null 2>&1",
		fmt.Sprintf("curl -fsSL %q -o /tmp/nixpacks.tgz", nixURL),
		"tar xzf /tmp/nixpacks.tgz -C /usr/local/bin/",
		"chmod +x /usr/local/bin/nixpacks",
		fmt.Sprintf("git clone --depth 1 --branch %q %q /src >/dev/null 2>&1", srcCfg.Branch, cloneURL),
		fmt.Sprintf("nixpacks build %q --name %q%s", buildDir, tag, envFlags),
	}, "\n")

	log.Info().
		Str("repo", srcCfg.RepoFullName).
		Str("branch", srcCfg.Branch).
		Str("tag", tag).
		Str("nixpacks", nixpacksVersion).
		Msg("Building image with Nixpacks")

	exitCode, logs, err := o.dockerClient.RunOneOffContainer(ctx, docker.OneOffSpec{
		Image:       "docker:27-cli",
		Command:     []string{"sh", "-c", script},
		Binds:       []string{"/var/run/docker.sock:/var/run/docker.sock"},
		Labels:      map[string]string{"orbita.builder": "nixpacks", "orbita.org": orgSlug, "orbita.managed": "true"},
		MaxLogBytes: 256 * 1024,
	})
	if err != nil {
		return "", fmt.Errorf("buildWithNixpacks: %w", err)
	}
	if exitCode != 0 {
		return "", fmt.Errorf("buildWithNixpacks: build failed (exit %d): %s", exitCode, truncate(tailLog(logs, 800), 800))
	}
	return tag, nil
}
