package orchestrator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
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
func (o *Orchestrator) deployCompose(ctx context.Context, app *models.Application, deployment *models.Deployment, srcCfg *SourceConfig, orgSlug string, envVars map[string]string) error {
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
		Override:    buildComposeOverride(web, orgNet, alias),
		DotEnv:      renderDotEnv(envVars),
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

	if err := o.dockerClient.WaitForServiceConverged(ctx, svcID, 5*time.Minute); err != nil {
		return fmt.Errorf("deployCompose: %s: %w", target, err)
	}

	deployment.ImageRef = "compose:" + stack
	app.Status = models.AppStatusRunning
	log.Info().Str("app", app.Name).Str("stack", stack).Msg("Compose stack deployed")
	return nil
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
	Override    string // the network/alias merge layer
}

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

	// Build from source only when the compose file actually declares a build —
	// installing the Compose plugin costs a download, so skip it otherwise.
	//
	// `docker stack deploy` ignores build: entirely and rejects any service
	// without an image ("no image specified"), which is how most compose files
	// in the wild are written. So after building we emit a third compose file
	// pinning each built service to the tag Compose produced
	// (<project>-<service>) and merge it in. Services that declare their own
	// image: aren't tagged that way, so the inspect skips them and their own
	// declaration wins.
	pluginURL := fmt.Sprintf("https://github.com/docker/compose/releases/download/v%s/docker-compose-linux-x86_64", composePluginVersion)
	steps = append(steps,
		"IMAGES_ARG=''",
		fmt.Sprintf("if grep -qE '^[[:space:]]+build:' %s; then", shellQuote(base)),
		"  echo '[orbita] compose file declares build: building images from source'",
		"  mkdir -p /usr/local/lib/docker/cli-plugins",
		fmt.Sprintf("  curl -fsSL %s -o /usr/local/lib/docker/cli-plugins/docker-compose", shellQuote(pluginURL)),
		"  chmod +x /usr/local/lib/docker/cli-plugins/docker-compose",
		fmt.Sprintf("  cd %s && docker compose -p %s -f %s build", projectDir, shellQuote(p.Stack), shellQuote(base)),
		"  echo 'services:' > /orbita-images.yml",
		fmt.Sprintf("  for s in $(docker compose -p %s -f %s config --services); do", shellQuote(p.Stack), shellQuote(base)),
		fmt.Sprintf("    if docker image inspect \"%s-$s:latest\" >/dev/null 2>&1; then", p.Stack),
		// Two echos rather than one printf: no backslash escaping to get wrong
		// between Go, the shell, and the container.
		"      echo \"  $s:\" >> /orbita-images.yml",
		fmt.Sprintf("      echo \"    image: %s-$s:latest\" >> /orbita-images.yml", p.Stack),
		"    fi",
		"  done",
		"  IMAGES_ARG='-c /orbita-images.yml'",
		"fi",
		fmt.Sprintf("cd %s && docker stack deploy --detach=true --prune --with-registry-auth -c %s $IMAGES_ARG -c /orbita-override.yml %s",
			projectDir, shellQuote(base), shellQuote(p.Stack)),
	)

	return strings.Join(steps, "\n")
}

// buildComposeOverride returns the merge layer that makes a compose stack
// routable by Traefik. Kept as a plain string so we never have to round-trip
// (and risk mangling) the user's own compose file.
func buildComposeOverride(webService, orgNet, alias string) string {
	return fmt.Sprintf(`services:
  %s:
    networks:
      default: {}
      %s:
        aliases:
          - %s
networks:
  %s:
    external: true
`, webService, orgNet, alias, orgNet)
}

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
