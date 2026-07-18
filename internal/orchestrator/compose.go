package orchestrator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/docker"
	"github.com/orbita-sh/orbita/internal/models"
)

// Pinned Compose plugin, installed into the docker:cli builder so `docker
// compose build` is available for services that build from source.
const composePluginVersion = "2.29.7"

// stackLabel is the label Swarm stamps on every service created by
// `docker stack deploy`. It's our only handle on "all services of this app".
const stackLabel = "com.docker.stack.namespace"

// stackName is the Swarm stack for an app. Services inside it are named
// <stack>_<service>.
func stackName(app *models.Application) string {
	return fmt.Sprintf("orbita-%s", app.ID.String()[:8])
}

// deployCompose brings up a multi-service Docker Compose stack.
//
// Compose is a different shape from every other source: one app maps to N Swarm
// services, not one. Rather than reimplement `docker stack deploy` against the
// SDK, we shell out to it from a one-off docker:cli container with the host
// socket bound in — the same proven trick as the Nixpacks builder.
//
// Two details make this fit Orbita without touching the routing layer:
//
//  1. We never parse the user's compose file. Instead we generate a tiny
//     override file and let `docker stack deploy -c base -c override` merge
//     them. Whatever they wrote keeps working.
//  2. The override attaches the web service to the org's overlay network with
//     the DNS alias `orbita-<app8>` — exactly the name Traefik already routes
//     to for single-service apps. So domains, TLS and routing need no changes.
func (o *Orchestrator) deployCompose(ctx context.Context, app *models.Application, deployment *models.Deployment, srcCfg *SourceConfig, deployCfg DeployConfig, orgSlug string, envVars map[string]string) error {
	web := strings.TrimSpace(srcCfg.ComposeService)
	if web == "" {
		return fmt.Errorf("deployCompose: no web service selected (source_config.compose_service)")
	}
	if srcCfg.ComposeContent == "" && srcCfg.RepoFullName == "" {
		return fmt.Errorf("deployCompose: needs either compose_content or a git repo")
	}

	stack := stackName(app)
	orgNet := docker.GetOrgNetworkName(orgSlug)
	alias := fmt.Sprintf("orbita-%s", app.ID.String()[:8])

	// Traefik must be on the org network or the alias resolves to nothing.
	_ = docker.EnsureTraefikOnOrgNetwork(orgSlug)

	params := composeScriptParams{
		Stack:       stack,
		ComposePath: strings.TrimSpace(srcCfg.ComposePath),
		Override:    buildComposeOverride(web, orgNet, alias, deployCfg),
		DotEnv:      renderDotEnv(envVars),
		AppEnv:      renderEnvFile(envVars),
	}
	if srcCfg.ComposeContent != "" {
		// Inline compose pasted into the dashboard — no clone needed.
		params.InlineYAML = srcCfg.ComposeContent
		params.ComposePath = "docker-compose.yml"
	} else {
		// From git: clone so relative build contexts resolve.
		cloneURL, err := o.composeCloneURL(ctx, app, srcCfg)
		if err != nil {
			return err
		}
		params.CloneURL = cloneURL
		params.Branch = srcCfg.Branch
		if params.Branch == "" {
			params.Branch = "main"
		}
	}

	composePath := params.ComposePath
	script := composeScript(params)

	log.Info().
		Str("app", app.Name).
		Str("stack", stack).
		Str("web_service", web).
		Str("compose", composePath).
		Msg("Deploying Docker Compose stack")

	exitCode, logs, err := o.dockerClient.RunOneOffContainer(ctx, docker.OneOffSpec{
		Image:       "docker:27-cli",
		Command:     []string{"sh", "-c", script},
		Binds:       []string{"/var/run/docker.sock:/var/run/docker.sock"},
		EnvVars:     envVars,
		Labels:      map[string]string{"orbita.builder": "compose", "orbita.org": orgSlug, "orbita.managed": "true"},
		MaxLogBytes: 256 * 1024,
	})
	if err != nil {
		return fmt.Errorf("deployCompose: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("deployCompose: stack deploy failed (exit %d): %s", exitCode, truncate(tailLog(logs, 1200), 1200))
	}

	// Find the web service Swarm created so the rest of Orbita (logs, metrics,
	// terminal, status) keeps working against a single service ID.
	target := stack + "_" + web
	svcID, err := o.findStackService(ctx, stack, target)
	if err != nil {
		return err
	}
	app.DockerServiceID = &svcID

	// Wait on every service, not just the web one. `stack deploy --detach`
	// returns as soon as Swarm accepts the specs, so without this a crashlooping
	// worker or database would be reported as a successful deploy.
	if err := o.waitForStackConverged(ctx, app, 5*time.Minute); err != nil {
		return fmt.Errorf("deployCompose: %w", err)
	}

	deployment.ImageRef = "compose:" + stack
	app.Status = models.AppStatusRunning
	log.Info().Str("app", app.Name).Str("stack", stack).Msg("Compose stack deployed")
	return nil
}

// waitForStackConverged blocks until every service in the stack has converged,
// reporting the specific service that failed. The deadline is shared across all
// services so a stack can't stall for timeout × N.
func (o *Orchestrator) waitForStackConverged(ctx context.Context, app *models.Application, timeout time.Duration) error {
	services, err := o.composeStackServices(ctx, app)
	if err != nil {
		return fmt.Errorf("list stack services: %w", err)
	}
	if len(services) == 0 {
		return fmt.Errorf("stack %s has no services", stackName(app))
	}

	deadline, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for _, svc := range services {
		remaining := time.Until(mustDeadline(deadline, timeout))
		if remaining <= 0 {
			return fmt.Errorf("timed out waiting for stack %s to converge", stackName(app))
		}
		if err := o.dockerClient.WaitForServiceConverged(deadline, svc.ID, remaining); err != nil {
			return fmt.Errorf("service %s did not start: %w", svc.Name, err)
		}
	}
	return nil
}

// mustDeadline returns ctx's deadline, or now+fallback if it has none.
func mustDeadline(ctx context.Context, fallback time.Duration) time.Time {
	if d, ok := ctx.Deadline(); ok {
		return d
	}
	return time.Now().Add(fallback)
}

// composeScriptParams is everything the builder script needs. Kept separate
// from the orchestrator so the generated shell can be unit-tested.
type composeScriptParams struct {
	Stack       string // swarm stack name, also the compose project name
	ComposePath string // compose file path relative to the project dir
	InlineYAML  string // pasted compose file (mutually exclusive with CloneURL)
	CloneURL    string // authenticated git URL (mutually exclusive with InlineYAML)
	Branch      string
	DotEnv      string // rendered .env for ${VAR} interpolation
	AppEnv      string // env_file injected into the running services
	Override    string // the network/alias/resources merge layer
}

// appEnvPath is where the app's env vars land inside the builder. Kept outside
// the project dir so a repo can never shadow it.
const appEnvPath = "/orbita-app.env"

// renderDotEnv turns the app's env vars into .env lines.
func renderDotEnv(envVars map[string]string) string {
	if len(envVars) == 0 {
		return ""
	}
	var b strings.Builder
	for k, v := range envVars {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(v)
		b.WriteString("\n")
	}
	return b.String()
}

// writeFileStep emits a shell line that materialises content at path. Content
// is base64'd so arbitrary YAML (quotes, newlines, $) survives the round trip.
func writeFileStep(path, content string) string {
	enc := base64.StdEncoding.EncodeToString([]byte(content))
	return fmt.Sprintf("echo %s | base64 -d > %s", shellQuote(enc), shellQuote(path))
}

// composeScript builds the shell run inside the docker:cli builder: materialise
// the compose project, build anything that needs building, then deploy the
// stack. Pure — no Docker calls — so it can be tested directly.
func composeScript(p composeScriptParams) string {
	const projectDir = "/work"

	composePath := p.ComposePath
	if composePath == "" {
		composePath = "docker-compose.yml"
	}
	base := projectDir + "/" + strings.TrimLeft(composePath, "/")

	steps := []string{
		"set -e",
		"apk add --no-cache curl git >/dev/null 2>&1",
	}

	if p.InlineYAML != "" {
		steps = append(steps,
			"mkdir -p "+projectDir,
			writeFileStep(base, p.InlineYAML),
		)
	} else {
		steps = append(steps,
			fmt.Sprintf("git clone --depth 1 --branch %q %q %s >/dev/null 2>&1", p.Branch, p.CloneURL, projectDir),
		)
	}

	// App env vars land in a .env file so ${VAR} interpolation inside the
	// compose file resolves the same way it does locally.
	if p.DotEnv != "" {
		steps = append(steps, writeFileStep(projectDir+"/.env", p.DotEnv))
	}

	steps = append(steps, writeFileStep("/orbita-override.yml", p.Override))

	// docker:cli ships the Compose plugin, but don't bet the deploy on it — fall
	// back to a pinned download if a future base image drops it.
	pluginURL := fmt.Sprintf("https://github.com/docker/compose/releases/download/v%s/docker-compose-linux-$(uname -m)", composePluginVersion)
	steps = append(steps,
		"if ! docker compose version >/dev/null 2>&1; then",
		"  echo '[orbita] installing the compose plugin'",
		"  mkdir -p /usr/local/lib/docker/cli-plugins",
		fmt.Sprintf("  curl -fsSL %s -o /usr/local/lib/docker/cli-plugins/docker-compose", shellQuote(pluginURL)),
		"  chmod +x /usr/local/lib/docker/cli-plugins/docker-compose",
		"fi",
		fmt.Sprintf("SERVICES=$(docker compose -p %s -f %s config --services)", shellQuote(p.Stack), shellQuote(base)),
	)

	// Orbita's env vars have to reach the running containers, not just
	// interpolate the compose file. env_file is applied to *every* service (a
	// worker needs DATABASE_URL as much as the web tier does); a service's own
	// `environment:` still wins, since Compose gives it precedence over env_file.
	envArg := ""
	if p.AppEnv != "" {
		steps = append(steps,
			writeFileStep(appEnvPath, p.AppEnv),
			"echo 'services:' > /orbita-env.yml",
			"for s in $SERVICES; do",
			"  echo \"  $s:\" >> /orbita-env.yml",
			"  echo '    env_file:' >> /orbita-env.yml",
			"  echo '      - "+appEnvPath+"' >> /orbita-env.yml",
			"done",
		)
		envArg = " -c /orbita-env.yml"
	}

	// `docker stack deploy` ignores build: entirely and rejects any service
	// without an image ("no image specified"), which is how most compose files
	// in the wild are written. So after building we emit another compose file
	// pinning each built service to the tag Compose produced
	// (<project>-<service>) and merge it in. Services that declare their own
	// image: aren't tagged that way, so the inspect skips them and their own
	// declaration wins.
	steps = append(steps,
		"IMAGES_ARG=''",
		fmt.Sprintf("if grep -qE '^[[:space:]]+build:' %s; then", shellQuote(base)),
		"  echo '[orbita] compose file declares build: building images from source'",
		fmt.Sprintf("  cd %s && docker compose -p %s -f %s build", projectDir, shellQuote(p.Stack), shellQuote(base)),
		"  echo 'services:' > /orbita-images.yml",
		"  for s in $SERVICES; do",
		fmt.Sprintf("    if docker image inspect \"%s-$s:latest\" >/dev/null 2>&1; then", p.Stack),
		// Two echos rather than one printf: no backslash escaping to get wrong
		// between Go, the shell, and the container.
		"      echo \"  $s:\" >> /orbita-images.yml",
		fmt.Sprintf("      echo \"    image: %s-$s:latest\" >> /orbita-images.yml", p.Stack),
		"    fi",
		"  done",
		"  IMAGES_ARG='-c /orbita-images.yml'",
		"fi",
		fmt.Sprintf("cd %s && docker stack deploy --detach=true --prune --with-registry-auth -c %s $IMAGES_ARG%s -c /orbita-override.yml %s",
			projectDir, shellQuote(base), envArg, shellQuote(p.Stack)),
	)

	return strings.Join(steps, "\n")
}

// buildComposeOverride returns the merge layer that makes a compose stack
// routable by Traefik. Kept as a plain string so we never have to round-trip
// (and risk mangling) the user's own compose file.
//
// Resource limits land on the web service only. Orbita's memory/CPU setting is
// a single per-container figure, and blanket-applying it to every service would
// both clobber any `deploy.resources` the user wrote and happily OOM a database
// that was sized for a web tier. Per-service limits belong in the compose file,
// which can express them properly.
func buildComposeOverride(webService, orgNet, alias string, deployCfg DeployConfig) string {
	var b strings.Builder
	fmt.Fprintf(&b, "services:\n  %s:\n", webService)
	fmt.Fprintf(&b, "    networks:\n      default: {}\n      %s:\n        aliases:\n          - %s\n", orgNet, alias)

	if deployCfg.MemoryMB > 0 || deployCfg.CPUShares > 0 {
		b.WriteString("    deploy:\n      resources:\n        limits:\n")
		if deployCfg.MemoryMB > 0 {
			fmt.Fprintf(&b, "          memory: %dM\n", deployCfg.MemoryMB)
		}
		if deployCfg.CPUShares > 0 {
			// 1000 shares == 1 core, matching the non-compose path.
			fmt.Fprintf(&b, "          cpus: \"%.2f\"\n", float64(deployCfg.CPUShares)/1000.0)
		}
	}

	fmt.Fprintf(&b, "networks:\n  %s:\n    external: true\n", orgNet)
	return b.String()
}

// renderEnvFile writes the app's env vars in docker-compose env_file format.
//
// Values are written raw. Verified against `docker stack deploy`: it takes each
// value literally to end of line — spaces, quotes, `#` and `$` all survive
// untouched, and `$` is NOT interpolated. Quoting would be actively wrong, as
// the surrounding quotes end up *inside* the value.
//
// The format can't represent a newline, so multi-line values (a PEM key, say)
// are skipped with a warning rather than silently corrupting every variable
// after them. Keys that aren't valid environment names are skipped too.
func renderEnvFile(envVars map[string]string) string {
	if len(envVars) == 0 {
		return ""
	}
	keys := make([]string, 0, len(envVars))
	for k := range envVars {
		if envKeyRe.MatchString(k) {
			keys = append(keys, k)
		} else {
			log.Warn().Str("key", k).Msg("compose: skipping env var with an invalid name")
		}
	}
	sort.Strings(keys) // deterministic output keeps deploys reproducible

	var b strings.Builder
	for _, k := range keys {
		v := envVars[k]
		if strings.ContainsAny(v, "\n\r\x00") {
			log.Warn().Str("key", k).Msg("compose: skipping env var — multi-line values can't be passed to a compose stack")
			continue
		}
		fmt.Fprintf(&b, "%s=%s\n", k, v)
	}
	return b.String()
}

// envKeyRe matches a POSIX-ish environment variable name.
var envKeyRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// findStackService resolves the Swarm service ID for <stack>_<service>. Swarm
// registers services a moment after `stack deploy` returns, so we retry briefly.
func (o *Orchestrator) findStackService(ctx context.Context, stack, fullName string) (string, error) {
	var names []string
	for attempt := 0; attempt < 10; attempt++ {
		services, err := o.dockerClient.ListServicesByLabel(ctx, stackLabel, stack)
		if err != nil {
			return "", fmt.Errorf("findStackService: %w", err)
		}
		names = names[:0]
		for _, s := range services {
			if s.Name == fullName {
				return s.ID, nil
			}
			names = append(names, s.Name)
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return "", fmt.Errorf("deployCompose: web service %q not found in stack (services: %s) — check the service name matches your compose file", fullName, strings.Join(names, ", "))
}

// composeCloneURL builds an authenticated clone URL, mirroring buildFromGit.
func (o *Orchestrator) composeCloneURL(ctx context.Context, app *models.Application, srcCfg *SourceConfig) (string, error) {
	token := ""
	if o.tokenFetcher != nil && srcCfg.GitConnectionID != "" {
		if _, t, _, err := o.tokenFetcher.ResolveGitToken(ctx, srcCfg.GitConnectionID, app.OrganizationID.String()); err != nil {
			log.Warn().Err(err).Msg("deployCompose: token resolve failed; attempting public clone")
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
			return "", fmt.Errorf("deployCompose: parse url: %w", err)
		}
		parsed.User = url.UserPassword("x-access-token", token)
		cloneURL = parsed.String()
	}
	return cloneURL, nil
}

// composeStackServices lists every Swarm service belonging to the app's stack.
func (o *Orchestrator) composeStackServices(ctx context.Context, app *models.Application) ([]docker.ServiceInfo, error) {
	return o.dockerClient.ListServicesByLabel(ctx, stackLabel, stackName(app))
}

// stopComposeStack scales every service in the stack to zero.
func (o *Orchestrator) stopComposeStack(ctx context.Context, app *models.Application) error {
	services, err := o.composeStackServices(ctx, app)
	if err != nil {
		return err
	}
	if len(services) == 0 {
		return fmt.Errorf("no services found for stack %s", stackName(app))
	}
	var firstErr error
	for _, svc := range services {
		if err := o.dockerClient.ScaleService(ctx, svc.ID, 0); err != nil {
			log.Warn().Err(err).Str("service", svc.Name).Msg("compose: stop failed")
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// startComposeStack brings a stopped stack back up.
//
// Swarm doesn't remember what each service was scaled to before we zeroed it,
// so we restore the web service to the app's replica count and everything else
// to a single replica — scaling a database or worker to the web tier's replica
// count would be wrong. Apps needing per-service scale should declare
// `deploy.replicas` in the compose file and redeploy, which is authoritative.
func (o *Orchestrator) startComposeStack(ctx context.Context, app *models.Application) error {
	services, err := o.composeStackServices(ctx, app)
	if err != nil {
		return err
	}
	if len(services) == 0 {
		return fmt.Errorf("no services found for stack %s", stackName(app))
	}

	webFull := ""
	if svc := composeWebService(app); svc != "" {
		webFull = stackName(app) + "_" + svc
	}
	webReplicas := app.Replicas
	if webReplicas < 1 {
		webReplicas = 1
	}

	var firstErr error
	for _, svc := range services {
		want := 1
		if svc.Name == webFull {
			want = webReplicas
		}
		if err := o.dockerClient.ScaleService(ctx, svc.ID, want); err != nil {
			log.Warn().Err(err).Str("service", svc.Name).Msg("compose: start failed")
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// composeWebService reads the routable service name back out of the app's
// stored source config. Best-effort: an empty result just means no service gets
// the app's replica count on start.
func composeWebService(app *models.Application) string {
	var cfg SourceConfig
	if err := json.Unmarshal(app.SourceConfig, &cfg); err != nil {
		return ""
	}
	return strings.TrimSpace(cfg.ComposeService)
}

// restartComposeStack restarts every service in place, preserving each one's
// replica count (unlike a stop/start cycle, which loses it).
func (o *Orchestrator) restartComposeStack(ctx context.Context, app *models.Application) error {
	services, err := o.composeStackServices(ctx, app)
	if err != nil {
		return err
	}
	if len(services) == 0 {
		return fmt.Errorf("no services found for stack %s", stackName(app))
	}
	var firstErr error
	for _, svc := range services {
		if err := o.dockerClient.ForceUpdateService(ctx, svc.ID); err != nil {
			log.Warn().Err(err).Str("service", svc.Name).Msg("compose: restart failed")
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// removeComposeStack tears down every service in the stack.
func (o *Orchestrator) removeComposeStack(ctx context.Context, app *models.Application) error {
	services, err := o.composeStackServices(ctx, app)
	if err != nil {
		return err
	}
	var firstErr error
	for _, svc := range services {
		if err := o.dockerClient.RemoveService(ctx, svc.ID); err != nil {
			log.Warn().Err(err).Str("service", svc.Name).Msg("compose: remove failed")
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// isCompose reports whether the app deploys as a Compose stack.
func isCompose(app *models.Application) bool {
	return app.SourceType == models.SourceTypeCompose
}
