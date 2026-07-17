package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/docker"
	"github.com/orbita-sh/orbita/internal/models"
)

type Orchestrator struct {
	dockerClient *docker.Client
	tokenFetcher TokenFetcher
}

// TokenFetcher resolves a git connection ID to a decrypted access token +
// provider metadata. Wired by the service layer at initialization. Accepts
// UUIDs as strings to avoid a circular import between orchestrator and service.
type TokenFetcher interface {
	ResolveGitToken(ctx context.Context, connIDStr, orgIDStr string) (provider, token, baseURL string, err error)
}

func New(dockerClient *docker.Client) *Orchestrator {
	return &Orchestrator{dockerClient: dockerClient}
}

// SetTokenFetcher wires a git token resolver. Used by git-source deploys.
func (o *Orchestrator) SetTokenFetcher(f TokenFetcher) {
	o.tokenFetcher = f
}

type SourceConfig struct {
	// docker-image
	Image string `json:"image"`

	// git / grit
	GitConnectionID string `json:"git_connection_id,omitempty"`
	RepoFullName    string `json:"repo_full_name,omitempty"`
	RepoURL         string `json:"repo_url,omitempty"`
	Branch          string `json:"branch,omitempty"`
	DockerfilePath  string `json:"dockerfile_path,omitempty"`
	BuildContext    string `json:"build_context,omitempty"`

	// grit only: build-time args baked into the image (e.g. NEXT_PUBLIC_API_URL
	// for the Next.js services) and the service role for labeling.
	BuildArgs map[string]string `json:"build_args,omitempty"`
	GritRole  string            `json:"grit_role,omitempty"`
}

type DeployConfig struct {
	MemoryMB  int `json:"memory_mb,omitempty"`
	CPUShares int `json:"cpu_shares,omitempty"` // 1000 = 1 core
}

func (o *Orchestrator) DeployApplication(ctx context.Context, app *models.Application, deployment *models.Deployment, orgSlug string, envVars map[string]string) error {
	log.Info().
		Str("app", app.Name).
		Str("source_type", app.SourceType).
		Str("deployment_id", deployment.ID.String()).
		Msg("Starting deployment")

	var srcCfg SourceConfig
	if err := json.Unmarshal(app.SourceConfig, &srcCfg); err != nil {
		return fmt.Errorf("DeployApplication: parse source config: %w", err)
	}

	var deployCfg DeployConfig
	_ = json.Unmarshal(app.DeployConfig, &deployCfg)

	// Resolve the image reference we'll run, building it from git if needed.
	imageRef, err := o.resolveImage(ctx, app, deployment, &srcCfg, orgSlug)
	if err != nil {
		return err
	}
	deployment.ImageRef = imageRef

	// Build the swarm service spec
	port := 0
	if app.Port != nil {
		port = *app.Port
	}

	spec := docker.ServiceSpec{
		Name:        fmt.Sprintf("orbita-%s", app.ID.String()[:8]),
		Image:       imageRef,
		Replicas:    app.Replicas,
		Port:        port,
		PublishPort: true, // ingress fallback; Traefik routes via the org network when domains exist
		EnvVars:     envVars,
		Labels: map[string]string{
			"orbita.app.id":  app.ID.String(),
			"orbita.org":     orgSlug,
			"orbita.managed": "true",
		},
		NetworkName:  docker.GetOrgNetworkName(orgSlug),
		CgroupParent: fmt.Sprintf("orbita-org-%s", orgSlug),
	}

	// Keep Traefik attached to this org's network on every deploy. Traefik drops
	// overlay attachments when its container is recreated (updates, reboots), and
	// only org-creation re-attached it — so without this, routing into the org
	// 502s after any Traefik restart. Idempotent + best-effort.
	_ = docker.EnsureTraefikOnOrgNetwork(orgSlug)

	// Apply per-container resource limits
	if deployCfg.MemoryMB > 0 {
		spec.MemoryLimit = int64(deployCfg.MemoryMB) * 1024 * 1024
	}
	if deployCfg.CPUShares > 0 {
		// NanoCPUs: 1e9 per core. 1000 shares == 1 core == 1e9 nanoCPUs
		spec.CPULimit = int64(deployCfg.CPUShares) * 1_000_000
	}

	if app.DockerServiceID != nil && *app.DockerServiceID != "" {
		if err := o.dockerClient.UpdateService(ctx, *app.DockerServiceID, spec); err != nil {
			return fmt.Errorf("DeployApplication: update service: %w", err)
		}
	} else {
		serviceID, err := o.dockerClient.CreateService(ctx, spec)
		if err != nil {
			return fmt.Errorf("DeployApplication: create service: %w", err)
		}
		app.DockerServiceID = &serviceID
	}

	// Gate success on the service actually converging. On a failed update Swarm
	// rolls back to the previous task (start-first order), so the running
	// version keeps serving and the deployment is recorded as failed.
	if err := o.dockerClient.WaitForServiceConverged(ctx, *app.DockerServiceID, 3*time.Minute); err != nil {
		return fmt.Errorf("DeployApplication: %w", err)
	}

	app.Status = models.AppStatusRunning
	log.Info().Str("app", app.Name).Str("image", imageRef).Msg("Deployment completed")
	return nil
}

// resolveImage returns a usable Docker image reference for deployment.
// - docker-image source: pulls srcCfg.Image and uses it directly
// - git source: builds via Docker's remote git context and tags as orbita-<slug>-<app>:<deploy>
func (o *Orchestrator) resolveImage(ctx context.Context, app *models.Application, deployment *models.Deployment, srcCfg *SourceConfig, orgSlug string) (string, error) {
	switch app.SourceType {
	case models.SourceTypeDockerImage:
		if srcCfg.Image == "" {
			return "", fmt.Errorf("resolveImage: empty image for docker-image source")
		}
		reader, err := o.dockerClient.PullImage(ctx, srcCfg.Image, "")
		if err != nil {
			return "", fmt.Errorf("resolveImage: pull: %w", err)
		}
		// Drain to ensure pull completes before we try to run it.
		_, _ = io.Copy(io.Discard, reader)
		reader.Close()
		return srcCfg.Image, nil

	case models.SourceTypeGit, models.SourceTypeGrit:
		// Grit services build exactly like a git source — a remote git context
		// with a specific Dockerfile, build context, and (for Next.js apps)
		// build args. The recipe is derived from grit.json (see internal/grit).
		return o.buildFromGit(ctx, app, deployment, srcCfg, orgSlug)

	default:
		return "", fmt.Errorf("resolveImage: unsupported source_type %q", app.SourceType)
	}
}

// buildFromGit uses Docker's built-in remote git context to clone and build an
// image. For private repos the PAT is embedded in the URL. Returns the local
// image tag to deploy.
func (o *Orchestrator) buildFromGit(ctx context.Context, app *models.Application, deployment *models.Deployment, srcCfg *SourceConfig, orgSlug string) (string, error) {
	if srcCfg.RepoFullName == "" {
		return "", fmt.Errorf("buildFromGit: repo_full_name missing")
	}
	if srcCfg.Branch == "" {
		return "", fmt.Errorf("buildFromGit: branch missing")
	}

	// Resolve token
	token := ""
	if o.tokenFetcher != nil && srcCfg.GitConnectionID != "" {
		_, t, _, err := o.tokenFetcher.ResolveGitToken(ctx, srcCfg.GitConnectionID, app.OrganizationID.String())
		if err != nil {
			log.Warn().Err(err).Msg("buildFromGit: token resolve failed; attempting public clone")
		} else {
			token = t
		}
	}

	// Compose git URL. We support GitHub-style clone URLs; self-hosted providers
	// should have supplied the clone URL at app-create time.
	cloneURL := srcCfg.RepoURL
	if cloneURL == "" {
		cloneURL = fmt.Sprintf("https://github.com/%s.git", srcCfg.RepoFullName)
	}

	// Embed the token for private-repo access: https://<token>@host/owner/repo.git
	if token != "" {
		parsed, err := url.Parse(cloneURL)
		if err != nil {
			return "", fmt.Errorf("buildFromGit: parse url: %w", err)
		}
		parsed.User = url.UserPassword("x-access-token", token)
		cloneURL = parsed.String()
	}

	// Docker's remote context syntax: URL#ref[:dir]
	remote := cloneURL + "#" + srcCfg.Branch
	if srcCfg.BuildContext != "" {
		remote = remote + ":" + strings.TrimLeft(srcCfg.BuildContext, "/")
	}

	dockerfile := srcCfg.DockerfilePath
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	// Tag: orbita-<orgSlug>-<app-short-id>:v<deploy-version>
	tag := fmt.Sprintf("orbita-%s-%s:v%d", orgSlug, app.ID.String()[:8], deployment.Version)

	log.Info().
		Str("repo", srcCfg.RepoFullName).
		Str("branch", srcCfg.Branch).
		Str("dockerfile", dockerfile).
		Str("tag", tag).
		Msg("Building image from git")

	reader, err := o.dockerClient.BuildImage(ctx, remote, tag, dockerfile, srcCfg.BuildArgs, nil)
	if err != nil {
		return "", fmt.Errorf("buildFromGit: %w", err)
	}
	defer reader.Close()

	// Drain build output. If the build fails, Docker returns an error message
	// as a JSON line in the stream — capture last 4KB for diagnostics.
	tail := drainLastBytes(reader, 4096)
	if strings.Contains(tail, "errorDetail") {
		return "", fmt.Errorf("buildFromGit: build failed: %s", truncate(tail, 500))
	}

	return tag, nil
}

// drainLastBytes reads a stream to EOF keeping only the last N bytes.
func drainLastBytes(r io.Reader, n int) string {
	buf := make([]byte, 0, n*2)
	chunk := make([]byte, 4096)
	for {
		m, err := r.Read(chunk)
		if m > 0 {
			buf = append(buf, chunk[:m]...)
			if len(buf) > n*2 {
				buf = buf[len(buf)-n:]
			}
		}
		if err != nil {
			break
		}
	}
	if len(buf) > n {
		buf = buf[len(buf)-n:]
	}
	return string(buf)
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

func (o *Orchestrator) StopApplication(ctx context.Context, app *models.Application) error {
	if app.DockerServiceID == nil {
		return nil
	}
	if err := o.dockerClient.ScaleService(ctx, *app.DockerServiceID, 0); err != nil {
		return fmt.Errorf("StopApplication: %w", err)
	}
	app.Status = models.AppStatusStopped
	return nil
}

func (o *Orchestrator) StartApplication(ctx context.Context, app *models.Application) error {
	if app.DockerServiceID == nil {
		return fmt.Errorf("StartApplication: no service ID")
	}
	if err := o.dockerClient.ScaleService(ctx, *app.DockerServiceID, app.Replicas); err != nil {
		return fmt.Errorf("StartApplication: %w", err)
	}
	app.Status = models.AppStatusRunning
	return nil
}

func (o *Orchestrator) RestartApplication(ctx context.Context, app *models.Application) error {
	if app.DockerServiceID == nil {
		return fmt.Errorf("RestartApplication: no service ID")
	}
	if err := o.dockerClient.ScaleService(ctx, *app.DockerServiceID, 0); err != nil {
		return fmt.Errorf("RestartApplication: stop: %w", err)
	}
	time.Sleep(500 * time.Millisecond)
	if err := o.dockerClient.ScaleService(ctx, *app.DockerServiceID, app.Replicas); err != nil {
		return fmt.Errorf("RestartApplication: start: %w", err)
	}
	app.Status = models.AppStatusRunning
	return nil
}

func (o *Orchestrator) RemoveApplication(ctx context.Context, app *models.Application) error {
	if app.DockerServiceID != nil {
		if err := o.dockerClient.RemoveService(ctx, *app.DockerServiceID); err != nil {
			return fmt.Errorf("RemoveApplication: %w", err)
		}
	}
	return nil
}

func (o *Orchestrator) GetApplicationStatus(ctx context.Context, app *models.Application) (string, error) {
	if app.DockerServiceID == nil {
		return app.Status, nil
	}
	info, err := o.dockerClient.InspectService(ctx, *app.DockerServiceID)
	if err != nil {
		return app.Status, nil
	}
	return info.Status, nil
}

// ExecInApplication runs a one-off command in the app's running container and
// returns combined output + exit code.
func (o *Orchestrator) ExecInApplication(ctx context.Context, app *models.Application, cmd []string) (*docker.ExecResult, error) {
	if app.DockerServiceID == nil || *app.DockerServiceID == "" {
		return nil, fmt.Errorf("ExecInApplication: app has no running service")
	}
	containerID, err := o.dockerClient.FindContainerIDForService(ctx, *app.DockerServiceID)
	if err != nil {
		return nil, fmt.Errorf("ExecInApplication: %w", err)
	}
	res, err := o.dockerClient.ExecInContainer(ctx, containerID, cmd, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("ExecInApplication: %w", err)
	}
	return res, nil
}

// GetApplicationMetrics returns live container stats for the app.
func (o *Orchestrator) GetApplicationMetrics(ctx context.Context, app *models.Application) (map[string]interface{}, error) {
	if app.DockerServiceID == nil || *app.DockerServiceID == "" {
		return nil, fmt.Errorf("GetApplicationMetrics: app has no running service")
	}
	containerID, err := o.dockerClient.FindContainerIDForService(ctx, *app.DockerServiceID)
	if err != nil {
		return nil, fmt.Errorf("GetApplicationMetrics: %w", err)
	}
	stats, err := o.dockerClient.GetContainerStats(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("GetApplicationMetrics: %w", err)
	}
	return stats, nil
}

func (o *Orchestrator) GetApplicationLogs(ctx context.Context, app *models.Application, tail int) (string, error) {
	if app.DockerServiceID == nil {
		return "No service running.\n", nil
	}
	reader, err := o.dockerClient.GetServiceLogs(ctx, *app.DockerServiceID, tail)
	if err != nil {
		return "", fmt.Errorf("GetApplicationLogs: %w", err)
	}
	defer reader.Close()

	buf := make([]byte, 64*1024)
	n, _ := reader.Read(buf)
	return string(buf[:n]), nil
}
