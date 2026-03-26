package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/docker"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

var (
	ErrTemplateNotFound = errors.New("template not found")
	ErrServiceNotFound  = errors.New("service not found")
)

type TemplateService struct {
	serviceRepo  *repository.ServiceRepository
	dockerClient *docker.Client
}

func NewTemplateService(serviceRepo *repository.ServiceRepository, dockerClient *docker.Client) *TemplateService {
	return &TemplateService{
		serviceRepo:  serviceRepo,
		dockerClient: dockerClient,
	}
}

func (s *TemplateService) ListTemplates(ctx context.Context) ([]models.Template, error) {
	return s.serviceRepo.ListTemplates(ctx)
}

func (s *TemplateService) GetTemplate(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	t, err := s.serviceRepo.FindTemplateByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("GetTemplate: %w", err)
	}
	return t, nil
}

type DeployServiceInput struct {
	TemplateID    uuid.UUID         `json:"template_id"`
	Name          string            `json:"name"`
	EnvironmentID uuid.UUID         `json:"environment_id"`
	Params        map[string]string `json:"params"`
}

func (s *TemplateService) DeployService(ctx context.Context, orgID uuid.UUID, orgSlug string, input DeployServiceInput) (*models.DeployedService, error) {
	tmpl, err := s.serviceRepo.FindTemplateByID(ctx, input.TemplateID)
	if err != nil {
		return nil, ErrTemplateNotFound
	}

	// Render compose template with params
	rendered, err := renderTemplate(tmpl.ComposeTemplate, input.Params)
	if err != nil {
		return nil, fmt.Errorf("DeployService: render template: %w", err)
	}

	config, _ := json.Marshal(map[string]interface{}{
		"params":   input.Params,
		"rendered": rendered,
	})

	svc := &models.DeployedService{
		ID:             uuid.New(),
		EnvironmentID:  input.EnvironmentID,
		OrganizationID: orgID,
		TemplateID:     &input.TemplateID,
		Name:           input.Name,
		Config:         config,
		Status:         models.ServiceStatusCreating,
	}

	if err := s.serviceRepo.Create(ctx, svc); err != nil {
		return nil, fmt.Errorf("DeployService: create: %w", err)
	}

	// Deploy via Docker (stub — creates a single service from the first image in the compose)
	serviceName := fmt.Sprintf("orbita-svc-%s", svc.ID.String()[:8])
	spec := docker.ServiceSpec{
		Name:         serviceName,
		Image:        "nginx:latest", // Placeholder — real impl would parse compose YAML
		Replicas:     1,
		NetworkName:  docker.GetOrgNetworkName(orgSlug),
		CgroupParent: fmt.Sprintf("orbita-org-%s", orgSlug),
		Labels: map[string]string{
			"orbita.service.id": svc.ID.String(),
			"orbita.org":        orgSlug,
			"orbita.managed":    "true",
		},
	}

	serviceID, err := s.dockerClient.CreateService(ctx, spec)
	if err != nil {
		svc.Status = models.ServiceStatusFailed
		_ = s.serviceRepo.Update(ctx, svc)
		return svc, fmt.Errorf("DeployService: docker: %w", err)
	}

	svc.DockerServiceIDs = []string{serviceID}
	svc.Status = models.ServiceStatusRunning
	_ = s.serviceRepo.Update(ctx, svc)

	log.Info().Str("service", svc.Name).Str("template", tmpl.Name).Msg("Service deployed")

	// Reload with template
	svc, _ = s.serviceRepo.FindByID(ctx, svc.ID, orgID)
	return svc, nil
}

func (s *TemplateService) GetService(ctx context.Context, id, orgID uuid.UUID) (*models.DeployedService, error) {
	svc, err := s.serviceRepo.FindByID(ctx, id, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, fmt.Errorf("GetService: %w", err)
	}
	return svc, nil
}

func (s *TemplateService) ListServices(ctx context.Context, orgID uuid.UUID) ([]models.DeployedService, error) {
	return s.serviceRepo.ListByOrgID(ctx, orgID)
}

func (s *TemplateService) DeleteService(ctx context.Context, id, orgID uuid.UUID) error {
	svc, err := s.serviceRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return ErrServiceNotFound
	}

	for _, dockerID := range svc.DockerServiceIDs {
		_ = s.dockerClient.RemoveService(ctx, dockerID)
	}

	return s.serviceRepo.Delete(ctx, id, orgID)
}

func renderTemplate(tmplStr string, params map[string]string) (string, error) {
	t, err := template.New("compose").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("renderTemplate: parse: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("renderTemplate: execute: %w", err)
	}

	return buf.String(), nil
}
