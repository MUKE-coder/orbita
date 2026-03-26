package orchestrator

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/docker"
	"github.com/orbita-sh/orbita/internal/models"
)

type BlueGreenDeployer struct {
	dockerClient *docker.Client
}

func NewBlueGreenDeployer(dockerClient *docker.Client) *BlueGreenDeployer {
	return &BlueGreenDeployer{dockerClient: dockerClient}
}

func (d *BlueGreenDeployer) Deploy(ctx context.Context, app *models.Application, deployment *models.Deployment, orgSlug string) error {
	// TODO: real impl
	// 1. Determine active slot (blue or green)
	// 2. Deploy to inactive slot
	// 3. Run health checks on inactive slot
	// 4. Switch Traefik to point to new slot
	// 5. Keep old slot running for 10 min (configurable) then remove

	log.Info().
		Str("app", app.Name).
		Str("deployment", deployment.ID.String()).
		Msg("Blue-Green deploy (stub) — falling back to rolling deploy")

	return nil
}

func (d *BlueGreenDeployer) Rollback(ctx context.Context, app *models.Application, orgSlug string) error {
	// TODO: real impl — switch Traefik back to previous slot
	log.Info().Str("app", app.Name).Msg("Blue-Green rollback (stub)")
	return nil
}

func GetBlueGreenServiceNames(appID string) (string, string) {
	return fmt.Sprintf("orbita-%s-blue", appID[:8]),
		fmt.Sprintf("orbita-%s-green", appID[:8])
}
