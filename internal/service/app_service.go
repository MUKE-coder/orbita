package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/orchestrator"
	"github.com/orbita-sh/orbita/internal/repository"
)

var (
	ErrAppNotFound        = errors.New("application not found")
	ErrDeploymentNotFound = errors.New("deployment not found")
	ErrAppAlreadyRunning  = errors.New("application is already running")
	ErrAppNotRunning      = errors.New("application is not running")
)

type AppService struct {
	appRepo        *repository.AppRepository
	orchestrator   *orchestrator.Orchestrator
	envResolver    EnvResolver
	routeRefresher RouteRefresher
}

// RouteRefresher keeps Traefik routes in sync with app lifecycle events.
// Satisfied by *DomainService. Optional: no-op when unset.
type RouteRefresher interface {
	RefreshAppRoutes(ctx context.Context, appID uuid.UUID, port int) error
	RemoveAppRoutes(ctx context.Context, appID uuid.UUID) error
}

// EnvResolver supplies the full env-var map (secrets decrypted) for a resource.
// Satisfied by *EnvService. Optional: deploys proceed with no env when unset.
type EnvResolver interface {
	GetEnvVarMap(ctx context.Context, resourceID uuid.UUID, resourceType string, orgID uuid.UUID) (map[string]string, error)
}

func NewAppService(appRepo *repository.AppRepository, orch *orchestrator.Orchestrator) *AppService {
	return &AppService{
		appRepo:      appRepo,
		orchestrator: orch,
	}
}

// SetEnvResolver wires env-var resolution into deploys.
func (s *AppService) SetEnvResolver(r EnvResolver) {
	s.envResolver = r
}

// SetRouteRefresher wires Traefik route maintenance into deploys and deletes.
func (s *AppService) SetRouteRefresher(r RouteRefresher) {
	s.routeRefresher = r
}

// refreshRoutes best-effort refreshes the app's Traefik routes after a deploy.
func (s *AppService) refreshRoutes(ctx context.Context, app *models.Application) {
	if s.routeRefresher == nil {
		return
	}
	port := 0
	if app.Port != nil {
		port = *app.Port
	}
	if err := s.routeRefresher.RefreshAppRoutes(ctx, app.ID, port); err != nil {
		log.Warn().Err(err).Str("app", app.Name).Msg("Failed to refresh Traefik routes after deploy")
	}
}

// resolveEnvForDeploy returns the env map for an app, or an empty map when no
// resolver is wired. Resolution failure aborts the deploy rather than silently
// starting the app without its configuration.
func (s *AppService) resolveEnvForDeploy(ctx context.Context, app *models.Application) (map[string]string, error) {
	if s.envResolver == nil {
		return map[string]string{}, nil
	}
	envVars, err := s.envResolver.GetEnvVarMap(ctx, app.ID, models.ResourceTypeApp, app.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("resolve env vars: %w", err)
	}
	return envVars, nil
}

type CreateAppInput struct {
	Name          string    `json:"name"`
	EnvironmentID uuid.UUID `json:"environment_id"`
	SourceType    string    `json:"source_type"` // "docker-image" | "git"
	Port          *int      `json:"port"`
	Replicas      int       `json:"replicas"`

	// Per-container resource limits (0 = unlimited)
	MemoryMB  int `json:"memory_mb"`
	CPUShares int `json:"cpu_shares"` // 1000 = 1 core

	// source_type = docker-image
	Image string `json:"image"`

	// source_type = git
	GitConnectionID *uuid.UUID `json:"git_connection_id"` // optional — public repos build without one
	RepoFullName    string     `json:"repo_full_name"`    // "owner/repo"
	RepoURL         string     `json:"repo_url"`          // public clone URL; token is added at build time
	Branch          string     `json:"branch"`
	DockerfilePath  string     `json:"dockerfile_path"`
	BuildContext    string     `json:"build_context"`
	AutoDeploy      *bool      `json:"auto_deploy"` // git apps default to true
}

func (s *AppService) CreateApp(ctx context.Context, orgID uuid.UUID, input CreateAppInput) (*models.Application, error) {
	// Build source_config JSON depending on source type
	src := map[string]interface{}{
		"image": input.Image, // kept for docker-image and as a build target tag for git
	}
	if input.SourceType == "git" {
		dockerfile := input.DockerfilePath
		if dockerfile == "" {
			dockerfile = "Dockerfile"
		}
		// Persist a canonical clone URL: webhook payloads match on it, and the
		// build path uses it. Derived from repo_full_name when not supplied.
		repoURL := input.RepoURL
		if repoURL == "" && input.RepoFullName != "" {
			repoURL = fmt.Sprintf("https://github.com/%s.git", input.RepoFullName)
		}
		src["git_connection_id"] = input.GitConnectionID
		src["repo_full_name"] = input.RepoFullName
		src["repo_url"] = repoURL
		src["branch"] = input.Branch
		src["dockerfile_path"] = dockerfile
		src["build_context"] = input.BuildContext
	}
	sourceConfig, _ := json.Marshal(src)

	// Persist per-container resource + port config into deploy_config
	deploy := map[string]interface{}{}
	if input.MemoryMB > 0 {
		deploy["memory_mb"] = input.MemoryMB
	}
	if input.CPUShares > 0 {
		deploy["cpu_shares"] = input.CPUShares
	}
	deployConfig, _ := json.Marshal(deploy)

	replicas := input.Replicas
	if replicas < 1 {
		replicas = 1
	}

	app := &models.Application{
		ID:             uuid.New(),
		EnvironmentID:  input.EnvironmentID,
		OrganizationID: orgID,
		Name:           input.Name,
		SourceType:     input.SourceType,
		SourceConfig:   sourceConfig,
		BuildConfig:    json.RawMessage("{}"),
		DeployConfig:   deployConfig,
		Status:         models.AppStatusCreated,
		Replicas:       replicas,
		Port:           input.Port,
	}

	if input.SourceType == "git" {
		// Auto-deploy on push is the default for git apps; a webhook secret is
		// always generated so unsigned webhook deliveries are never accepted.
		autoDeploy := true
		if input.AutoDeploy != nil {
			autoDeploy = *input.AutoDeploy
		}
		app.AutoDeploy = autoDeploy

		secret, err := generateWebhookSecret()
		if err != nil {
			return nil, fmt.Errorf("CreateApp: webhook secret: %w", err)
		}
		app.WebhookSecret = &secret
	}

	if err := s.appRepo.Create(ctx, app); err != nil {
		return nil, fmt.Errorf("CreateApp: %w", err)
	}
	return app, nil
}

// generateWebhookSecret returns a 32-byte hex secret for HMAC-SHA256 webhook
// signature verification.
func generateWebhookSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// RegenerateWebhookSecret rotates the app's webhook secret and returns the new
// value (the only time it is exposed).
func (s *AppService) RegenerateWebhookSecret(ctx context.Context, appID, orgID uuid.UUID) (string, error) {
	app, err := s.appRepo.FindByID(ctx, appID, orgID)
	if err != nil {
		return "", ErrAppNotFound
	}
	secret, err := generateWebhookSecret()
	if err != nil {
		return "", fmt.Errorf("RegenerateWebhookSecret: %w", err)
	}
	app.WebhookSecret = &secret
	if err := s.appRepo.Update(ctx, app); err != nil {
		return "", fmt.Errorf("RegenerateWebhookSecret: %w", err)
	}
	return secret, nil
}

func (s *AppService) GetApp(ctx context.Context, id, orgID uuid.UUID) (*models.Application, error) {
	app, err := s.appRepo.FindByID(ctx, id, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAppNotFound
		}
		return nil, fmt.Errorf("GetApp: %w", err)
	}
	return app, nil
}

func (s *AppService) ListApps(ctx context.Context, orgID uuid.UUID) ([]models.Application, error) {
	return s.appRepo.ListByOrgID(ctx, orgID)
}

func (s *AppService) UpdateApp(ctx context.Context, id, orgID uuid.UUID, updates map[string]interface{}) (*models.Application, error) {
	app, err := s.appRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, ErrAppNotFound
	}

	if name, ok := updates["name"].(string); ok && name != "" {
		app.Name = name
	}
	if port, ok := updates["port"].(float64); ok {
		p := int(port)
		app.Port = &p
	}
	if replicas, ok := updates["replicas"].(float64); ok {
		app.Replicas = int(replicas)
	}
	if autoDeploy, ok := updates["auto_deploy"].(bool); ok {
		app.AutoDeploy = autoDeploy
		// Enabling auto-deploy on an app that predates always-on webhook
		// secrets must not open an unsigned-webhook hole.
		if autoDeploy && (app.WebhookSecret == nil || *app.WebhookSecret == "") {
			secret, err := generateWebhookSecret()
			if err != nil {
				return nil, fmt.Errorf("UpdateApp: webhook secret: %w", err)
			}
			app.WebhookSecret = &secret
		}
	}

	if err := s.appRepo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("UpdateApp: %w", err)
	}
	return app, nil
}

func (s *AppService) DeleteApp(ctx context.Context, id, orgID uuid.UUID, orgSlug string) error {
	app, err := s.appRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return ErrAppNotFound
	}

	if err := s.orchestrator.RemoveApplication(ctx, app); err != nil {
		return fmt.Errorf("DeleteApp: remove: %w", err)
	}

	if s.routeRefresher != nil {
		if err := s.routeRefresher.RemoveAppRoutes(ctx, app.ID); err != nil {
			log.Warn().Err(err).Str("app", app.Name).Msg("Failed to remove Traefik routes on app delete")
		}
	}

	return s.appRepo.Delete(ctx, id, orgID)
}

func (s *AppService) Deploy(ctx context.Context, appID, orgID uuid.UUID, orgSlug string, userID *uuid.UUID) (*models.Deployment, error) {
	return s.DeployWithTrigger(ctx, appID, orgID, orgSlug, userID, models.TriggerManual)
}

// DeployWithTrigger deploys an app recording how the deploy was initiated
// (manual, webhook, ...).
func (s *AppService) DeployWithTrigger(ctx context.Context, appID, orgID uuid.UUID, orgSlug string, userID *uuid.UUID, triggerType string) (*models.Deployment, error) {
	app, err := s.appRepo.FindByID(ctx, appID, orgID)
	if err != nil {
		return nil, ErrAppNotFound
	}

	// Parse image from source config
	var srcCfg struct {
		Image string `json:"image"`
	}
	_ = json.Unmarshal(app.SourceConfig, &srcCfg)

	envVars, err := s.resolveEnvForDeploy(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("Deploy: %w", err)
	}

	version, _ := s.appRepo.GetNextDeployVersion(ctx, appID)
	now := time.Now()

	deployment := &models.Deployment{
		ID:           uuid.New(),
		AppID:        appID,
		Version:      version,
		ImageRef:     srcCfg.Image,
		DeployConfig: app.DeployConfig,
		Status:       models.DeployStatusRunning,
		StartedAt:    &now,
		TriggeredBy:  userID,
		TriggerType:  triggerType,
	}

	if err := s.appRepo.CreateDeployment(ctx, deployment); err != nil {
		return nil, fmt.Errorf("Deploy: create deployment: %w", err)
	}

	// Update app status
	app.Status = models.AppStatusDeploying
	_ = s.appRepo.Update(ctx, app)

	// Run deployment
	if err := s.orchestrator.DeployApplication(ctx, app, deployment, orgSlug, envVars); err != nil {
		deployment.Status = models.DeployStatusFailed
		errMsg := err.Error()
		deployment.ErrorMessage = &errMsg
		finishedAt := time.Now()
		deployment.FinishedAt = &finishedAt
		_ = s.appRepo.UpdateDeployment(ctx, deployment)

		app.Status = models.AppStatusFailed
		_ = s.appRepo.Update(ctx, app)

		return deployment, fmt.Errorf("Deploy: %w", err)
	}

	// Success
	deployment.Status = models.DeployStatusSuccess
	finishedAt := time.Now()
	deployment.FinishedAt = &finishedAt
	_ = s.appRepo.UpdateDeployment(ctx, deployment)
	_ = s.appRepo.Update(ctx, app)

	s.refreshRoutes(ctx, app)

	return deployment, nil
}

func (s *AppService) Rollback(ctx context.Context, appID, deploymentID, orgID uuid.UUID, orgSlug string, userID *uuid.UUID) (*models.Deployment, error) {
	app, err := s.appRepo.FindByID(ctx, appID, orgID)
	if err != nil {
		return nil, ErrAppNotFound
	}

	targetDeploy, err := s.appRepo.FindDeploymentByID(ctx, deploymentID)
	if err != nil {
		return nil, ErrDeploymentNotFound
	}

	if targetDeploy.AppID != appID {
		return nil, ErrDeploymentNotFound
	}

	envVars, err := s.resolveEnvForDeploy(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("Rollback: %w", err)
	}

	version, _ := s.appRepo.GetNextDeployVersion(ctx, appID)
	now := time.Now()

	rollbackDeploy := &models.Deployment{
		ID:           uuid.New(),
		AppID:        appID,
		Version:      version,
		ImageRef:     targetDeploy.ImageRef,
		DeployConfig: targetDeploy.DeployConfig,
		Status:       models.DeployStatusRunning,
		StartedAt:    &now,
		TriggeredBy:  userID,
		TriggerType:  models.TriggerRollback,
	}

	if err := s.appRepo.CreateDeployment(ctx, rollbackDeploy); err != nil {
		return nil, fmt.Errorf("Rollback: create deployment: %w", err)
	}

	if err := s.orchestrator.DeployApplication(ctx, app, rollbackDeploy, orgSlug, envVars); err != nil {
		rollbackDeploy.Status = models.DeployStatusFailed
		errMsg := err.Error()
		rollbackDeploy.ErrorMessage = &errMsg
		finishedAt := time.Now()
		rollbackDeploy.FinishedAt = &finishedAt
		_ = s.appRepo.UpdateDeployment(ctx, rollbackDeploy)

		app.Status = models.AppStatusFailed
		_ = s.appRepo.Update(ctx, app)
		return rollbackDeploy, err
	}

	rollbackDeploy.Status = models.DeployStatusSuccess
	finishedAt := time.Now()
	rollbackDeploy.FinishedAt = &finishedAt
	_ = s.appRepo.UpdateDeployment(ctx, rollbackDeploy)
	_ = s.appRepo.Update(ctx, app)

	s.refreshRoutes(ctx, app)

	return rollbackDeploy, nil
}

func (s *AppService) Stop(ctx context.Context, appID, orgID uuid.UUID) error {
	app, err := s.appRepo.FindByID(ctx, appID, orgID)
	if err != nil {
		return ErrAppNotFound
	}
	if err := s.orchestrator.StopApplication(ctx, app); err != nil {
		return fmt.Errorf("Stop: %w", err)
	}
	return s.appRepo.Update(ctx, app)
}

func (s *AppService) Start(ctx context.Context, appID, orgID uuid.UUID) error {
	app, err := s.appRepo.FindByID(ctx, appID, orgID)
	if err != nil {
		return ErrAppNotFound
	}
	if err := s.orchestrator.StartApplication(ctx, app); err != nil {
		return fmt.Errorf("Start: %w", err)
	}
	return s.appRepo.Update(ctx, app)
}

func (s *AppService) Restart(ctx context.Context, appID, orgID uuid.UUID) error {
	app, err := s.appRepo.FindByID(ctx, appID, orgID)
	if err != nil {
		return ErrAppNotFound
	}
	if err := s.orchestrator.RestartApplication(ctx, app); err != nil {
		return fmt.Errorf("Restart: %w", err)
	}
	return s.appRepo.Update(ctx, app)
}

func (s *AppService) GetLogs(ctx context.Context, appID, orgID uuid.UUID, tail int) (string, error) {
	app, err := s.appRepo.FindByID(ctx, appID, orgID)
	if err != nil {
		return "", ErrAppNotFound
	}
	return s.orchestrator.GetApplicationLogs(ctx, app, tail)
}

func (s *AppService) GetStatus(ctx context.Context, appID, orgID uuid.UUID) (string, error) {
	app, err := s.appRepo.FindByID(ctx, appID, orgID)
	if err != nil {
		return "", ErrAppNotFound
	}
	return s.orchestrator.GetApplicationStatus(ctx, app)
}

func (s *AppService) ListDeployments(ctx context.Context, appID uuid.UUID, limit int) ([]models.Deployment, error) {
	return s.appRepo.ListDeployments(ctx, appID, limit)
}

func (s *AppService) GetMetrics(ctx context.Context, appID, orgID uuid.UUID) (map[string]interface{}, error) {
	app, err := s.appRepo.FindByID(ctx, appID, orgID)
	if err != nil {
		return nil, ErrAppNotFound
	}

	if app.DockerServiceID == nil {
		return map[string]interface{}{
			"cpu_percent":  0,
			"memory_usage": 0,
			"memory_limit": 0,
			"status":       app.Status,
		}, nil
	}

	// TODO: real metrics from Docker stats API
	return map[string]interface{}{
		"cpu_percent":  2.5,
		"memory_usage": 67108864,
		"memory_limit": 134217728,
		"network_rx":   1024000,
		"network_tx":   512000,
		"status":       app.Status,
	}, nil
}
