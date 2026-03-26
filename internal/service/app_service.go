package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	appRepo      *repository.AppRepository
	orchestrator *orchestrator.Orchestrator
}

func NewAppService(appRepo *repository.AppRepository, orch *orchestrator.Orchestrator) *AppService {
	return &AppService{
		appRepo:      appRepo,
		orchestrator: orch,
	}
}

type CreateAppInput struct {
	Name          string `json:"name"`
	EnvironmentID uuid.UUID `json:"environment_id"`
	SourceType    string `json:"source_type"`
	Image         string `json:"image"`
	Port          *int   `json:"port"`
	Replicas      int    `json:"replicas"`
}

func (s *AppService) CreateApp(ctx context.Context, orgID uuid.UUID, input CreateAppInput) (*models.Application, error) {
	sourceConfig, _ := json.Marshal(map[string]string{"image": input.Image})

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
		DeployConfig:   json.RawMessage("{}"),
		Status:         models.AppStatusCreated,
		Replicas:       replicas,
		Port:           input.Port,
	}

	if err := s.appRepo.Create(ctx, app); err != nil {
		return nil, fmt.Errorf("CreateApp: %w", err)
	}
	return app, nil
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

	return s.appRepo.Delete(ctx, id, orgID)
}

func (s *AppService) Deploy(ctx context.Context, appID, orgID uuid.UUID, orgSlug string, userID *uuid.UUID) (*models.Deployment, error) {
	app, err := s.appRepo.FindByID(ctx, appID, orgID)
	if err != nil {
		return nil, ErrAppNotFound
	}

	// Parse image from source config
	var srcCfg struct {
		Image string `json:"image"`
	}
	_ = json.Unmarshal(app.SourceConfig, &srcCfg)

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
		TriggerType:  models.TriggerManual,
	}

	if err := s.appRepo.CreateDeployment(ctx, deployment); err != nil {
		return nil, fmt.Errorf("Deploy: create deployment: %w", err)
	}

	// Update app status
	app.Status = models.AppStatusDeploying
	_ = s.appRepo.Update(ctx, app)

	// Run deployment
	if err := s.orchestrator.DeployApplication(ctx, app, deployment, orgSlug); err != nil {
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

	return deployment, nil
}

func (s *AppService) Rollback(ctx context.Context, appID, deploymentID, orgID uuid.UUID, orgSlug string) (*models.Deployment, error) {
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
		TriggerType:  "rollback",
	}

	if err := s.appRepo.CreateDeployment(ctx, rollbackDeploy); err != nil {
		return nil, fmt.Errorf("Rollback: create deployment: %w", err)
	}

	if err := s.orchestrator.DeployApplication(ctx, app, rollbackDeploy, orgSlug); err != nil {
		rollbackDeploy.Status = models.DeployStatusFailed
		errMsg := err.Error()
		rollbackDeploy.ErrorMessage = &errMsg
		_ = s.appRepo.UpdateDeployment(ctx, rollbackDeploy)
		return rollbackDeploy, err
	}

	rollbackDeploy.Status = models.DeployStatusSuccess
	finishedAt := time.Now()
	rollbackDeploy.FinishedAt = &finishedAt
	_ = s.appRepo.UpdateDeployment(ctx, rollbackDeploy)
	_ = s.appRepo.Update(ctx, app)

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
