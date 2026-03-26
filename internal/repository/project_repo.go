package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Projects

func (r *ProjectRepository) Create(ctx context.Context, project *models.Project) error {
	if err := r.db.WithContext(ctx).Create(project).Error; err != nil {
		return fmt.Errorf("ProjectRepo.Create: %w", err)
	}
	return nil
}

func (r *ProjectRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*models.Project, error) {
	var project models.Project
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Preload("Environments").Where("id = ?", id).First(&project).Error; err != nil {
		return nil, fmt.Errorf("ProjectRepo.FindByID: %w", err)
	}
	return &project, nil
}

func (r *ProjectRepository) ListByOrgID(ctx context.Context, orgID uuid.UUID) ([]models.Project, error) {
	var projects []models.Project
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Preload("Environments").Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("ProjectRepo.ListByOrgID: %w", err)
	}
	return projects, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *models.Project) error {
	if err := r.db.WithContext(ctx).Save(project).Error; err != nil {
		return fmt.Errorf("ProjectRepo.Update: %w", err)
	}
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("id = ?", id).Delete(&models.Project{}).Error; err != nil {
		return fmt.Errorf("ProjectRepo.Delete: %w", err)
	}
	return nil
}

// Environments

func (r *ProjectRepository) CreateEnvironment(ctx context.Context, env *models.Environment) error {
	if err := r.db.WithContext(ctx).Create(env).Error; err != nil {
		return fmt.Errorf("ProjectRepo.CreateEnvironment: %w", err)
	}
	return nil
}

func (r *ProjectRepository) FindEnvironmentByID(ctx context.Context, id uuid.UUID) (*models.Environment, error) {
	var env models.Environment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&env).Error; err != nil {
		return nil, fmt.Errorf("ProjectRepo.FindEnvironmentByID: %w", err)
	}
	return &env, nil
}

func (r *ProjectRepository) ListEnvironments(ctx context.Context, projectID uuid.UUID) ([]models.Environment, error) {
	var envs []models.Environment
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).
		Order("created_at ASC").Find(&envs).Error; err != nil {
		return nil, fmt.Errorf("ProjectRepo.ListEnvironments: %w", err)
	}
	return envs, nil
}

func (r *ProjectRepository) UpdateEnvironment(ctx context.Context, env *models.Environment) error {
	if err := r.db.WithContext(ctx).Save(env).Error; err != nil {
		return fmt.Errorf("ProjectRepo.UpdateEnvironment: %w", err)
	}
	return nil
}

func (r *ProjectRepository) DeleteEnvironment(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Environment{}).Error; err != nil {
		return fmt.Errorf("ProjectRepo.DeleteEnvironment: %w", err)
	}
	return nil
}
