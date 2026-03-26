package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type AppRepository struct {
	db *gorm.DB
}

func NewAppRepository(db *gorm.DB) *AppRepository {
	return &AppRepository{db: db}
}

func (r *AppRepository) Create(ctx context.Context, app *models.Application) error {
	if err := r.db.WithContext(ctx).Create(app).Error; err != nil {
		return fmt.Errorf("AppRepo.Create: %w", err)
	}
	return nil
}

func (r *AppRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*models.Application, error) {
	var app models.Application
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("id = ?", id).First(&app).Error; err != nil {
		return nil, fmt.Errorf("AppRepo.FindByID: %w", err)
	}
	return &app, nil
}

func (r *AppRepository) ListByOrgID(ctx context.Context, orgID uuid.UUID) ([]models.Application, error) {
	var apps []models.Application
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Order("created_at DESC").Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("AppRepo.ListByOrgID: %w", err)
	}
	return apps, nil
}

func (r *AppRepository) ListByEnvID(ctx context.Context, envID, orgID uuid.UUID) ([]models.Application, error) {
	var apps []models.Application
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("environment_id = ?", envID).Order("created_at DESC").Find(&apps).Error; err != nil {
		return nil, fmt.Errorf("AppRepo.ListByEnvID: %w", err)
	}
	return apps, nil
}

func (r *AppRepository) Update(ctx context.Context, app *models.Application) error {
	if err := r.db.WithContext(ctx).Save(app).Error; err != nil {
		return fmt.Errorf("AppRepo.Update: %w", err)
	}
	return nil
}

func (r *AppRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("id = ?", id).Delete(&models.Application{}).Error; err != nil {
		return fmt.Errorf("AppRepo.Delete: %w", err)
	}
	return nil
}

// Deployments

func (r *AppRepository) CreateDeployment(ctx context.Context, d *models.Deployment) error {
	if err := r.db.WithContext(ctx).Create(d).Error; err != nil {
		return fmt.Errorf("AppRepo.CreateDeployment: %w", err)
	}
	return nil
}

func (r *AppRepository) FindDeploymentByID(ctx context.Context, id uuid.UUID) (*models.Deployment, error) {
	var d models.Deployment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&d).Error; err != nil {
		return nil, fmt.Errorf("AppRepo.FindDeploymentByID: %w", err)
	}
	return &d, nil
}

func (r *AppRepository) ListDeployments(ctx context.Context, appID uuid.UUID, limit int) ([]models.Deployment, error) {
	var deployments []models.Deployment
	q := r.db.WithContext(ctx).Where("app_id = ?", appID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&deployments).Error; err != nil {
		return nil, fmt.Errorf("AppRepo.ListDeployments: %w", err)
	}
	return deployments, nil
}

func (r *AppRepository) UpdateDeployment(ctx context.Context, d *models.Deployment) error {
	if err := r.db.WithContext(ctx).Save(d).Error; err != nil {
		return fmt.Errorf("AppRepo.UpdateDeployment: %w", err)
	}
	return nil
}

func (r *AppRepository) GetNextDeployVersion(ctx context.Context, appID uuid.UUID) (int, error) {
	var maxVersion int
	err := r.db.WithContext(ctx).Model(&models.Deployment{}).
		Where("app_id = ?", appID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error
	if err != nil {
		return 0, fmt.Errorf("AppRepo.GetNextDeployVersion: %w", err)
	}
	return maxVersion + 1, nil
}
