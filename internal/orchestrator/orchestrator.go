package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/docker"
	"github.com/orbita-sh/orbita/internal/models"
)

type Orchestrator struct {
	dockerClient *docker.Client
}

func New(dockerClient *docker.Client) *Orchestrator {
	return &Orchestrator{dockerClient: dockerClient}
}

type SourceConfig struct {
	Image    string `json:"image"`
	RepoURL  string `json:"repo_url,omitempty"`
	Branch   string `json:"branch,omitempty"`
	RootDir  string `json:"root_dir,omitempty"`
}

func (o *Orchestrator) DeployApplication(ctx context.Context, app *models.Application, deployment *models.Deployment, orgSlug string) error {
	log.Info().Str("app", app.Name).Str("deployment_id", deployment.ID.String()).Msg("Starting deployment")

	// Parse source config
	var srcCfg SourceConfig
	if err := json.Unmarshal(app.SourceConfig, &srcCfg); err != nil {
		return fmt.Errorf("DeployApplication: parse source config: %w", err)
	}

	imageRef := deployment.ImageRef
	if imageRef == "" {
		imageRef = srcCfg.Image
	}

	// Pull image
	reader, err := o.dockerClient.PullImage(ctx, imageRef, "")
	if err != nil {
		return fmt.Errorf("DeployApplication: pull image: %w", err)
	}
	defer reader.Close()

	// Build service spec
	port := 0
	if app.Port != nil {
		port = *app.Port
	}

	spec := docker.ServiceSpec{
		Name:         fmt.Sprintf("orbita-%s", app.ID.String()[:8]),
		Image:        imageRef,
		Replicas:     app.Replicas,
		Port:         port,
		Labels: map[string]string{
			"orbita.app.id":  app.ID.String(),
			"orbita.org":     orgSlug,
			"orbita.managed": "true",
		},
		NetworkName:  docker.GetOrgNetworkName(orgSlug),
		CgroupParent: fmt.Sprintf("orbita-org-%s", orgSlug),
	}

	if app.DockerServiceID != nil && *app.DockerServiceID != "" {
		// Update existing service
		if err := o.dockerClient.UpdateService(ctx, *app.DockerServiceID, spec); err != nil {
			return fmt.Errorf("DeployApplication: update service: %w", err)
		}
	} else {
		// Create new service
		serviceID, err := o.dockerClient.CreateService(ctx, spec)
		if err != nil {
			return fmt.Errorf("DeployApplication: create service: %w", err)
		}
		app.DockerServiceID = &serviceID
	}

	app.Status = models.AppStatusRunning
	log.Info().Str("app", app.Name).Msg("Deployment completed")

	return nil
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
	// Stop then start
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
