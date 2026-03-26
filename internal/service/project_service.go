package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrEnvNotFound     = errors.New("environment not found")
)

type ProjectService struct {
	projectRepo *repository.ProjectRepository
}

func NewProjectService(projectRepo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}

func (s *ProjectService) CreateProject(ctx context.Context, orgID uuid.UUID, name string, description *string, emoji string) (*models.Project, error) {
	if emoji == "" {
		emoji = "🚀"
	}

	project := &models.Project{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		Description:    description,
		Emoji:          emoji,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("CreateProject: %w", err)
	}

	// Create default environments
	for _, env := range []struct {
		name    string
		envType string
	}{
		{"Production", models.EnvTypeProduction},
		{"Staging", models.EnvTypeStaging},
	} {
		e := &models.Environment{
			ID:        uuid.New(),
			ProjectID: project.ID,
			Name:      env.name,
			Type:      env.envType,
		}
		if err := s.projectRepo.CreateEnvironment(ctx, e); err != nil {
			return nil, fmt.Errorf("CreateProject: create env %s: %w", env.name, err)
		}
	}

	// Reload with environments
	project, _ = s.projectRepo.FindByID(ctx, project.ID, orgID)
	return project, nil
}

func (s *ProjectService) GetProject(ctx context.Context, id, orgID uuid.UUID) (*models.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("GetProject: %w", err)
	}
	return project, nil
}

func (s *ProjectService) ListProjects(ctx context.Context, orgID uuid.UUID) ([]models.Project, error) {
	return s.projectRepo.ListByOrgID(ctx, orgID)
}

func (s *ProjectService) UpdateProject(ctx context.Context, id, orgID uuid.UUID, name *string, description *string, emoji *string) (*models.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	if name != nil {
		project.Name = *name
	}
	if description != nil {
		project.Description = description
	}
	if emoji != nil {
		project.Emoji = *emoji
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("UpdateProject: %w", err)
	}
	return project, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, id, orgID uuid.UUID) error {
	_, err := s.projectRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return ErrProjectNotFound
	}
	return s.projectRepo.Delete(ctx, id, orgID)
}

// Environments

func (s *ProjectService) CreateEnvironment(ctx context.Context, projectID, orgID uuid.UUID, name, envType string) (*models.Environment, error) {
	// Verify project belongs to org
	_, err := s.projectRepo.FindByID(ctx, projectID, orgID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	if envType == "" {
		envType = models.EnvTypeCustom
	}

	env := &models.Environment{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      name,
		Type:      envType,
	}

	if err := s.projectRepo.CreateEnvironment(ctx, env); err != nil {
		return nil, fmt.Errorf("CreateEnvironment: %w", err)
	}
	return env, nil
}

func (s *ProjectService) ListEnvironments(ctx context.Context, projectID, orgID uuid.UUID) ([]models.Environment, error) {
	// Verify project belongs to org
	_, err := s.projectRepo.FindByID(ctx, projectID, orgID)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	return s.projectRepo.ListEnvironments(ctx, projectID)
}

func (s *ProjectService) UpdateEnvironment(ctx context.Context, envID, projectID, orgID uuid.UUID, name *string) (*models.Environment, error) {
	// Verify project belongs to org
	_, err := s.projectRepo.FindByID(ctx, projectID, orgID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	env, err := s.projectRepo.FindEnvironmentByID(ctx, envID)
	if err != nil {
		return nil, ErrEnvNotFound
	}

	if env.ProjectID != projectID {
		return nil, ErrEnvNotFound
	}

	if name != nil {
		env.Name = *name
	}

	if err := s.projectRepo.UpdateEnvironment(ctx, env); err != nil {
		return nil, fmt.Errorf("UpdateEnvironment: %w", err)
	}
	return env, nil
}

func (s *ProjectService) DeleteEnvironment(ctx context.Context, envID, projectID, orgID uuid.UUID) error {
	// Verify project belongs to org
	_, err := s.projectRepo.FindByID(ctx, projectID, orgID)
	if err != nil {
		return ErrProjectNotFound
	}

	env, err := s.projectRepo.FindEnvironmentByID(ctx, envID)
	if err != nil {
		return ErrEnvNotFound
	}

	if env.ProjectID != projectID {
		return ErrEnvNotFound
	}

	return s.projectRepo.DeleteEnvironment(ctx, envID)
}
