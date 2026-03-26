package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

// Templates

func (r *ServiceRepository) ListTemplates(ctx context.Context) ([]models.Template, error) {
	var templates []models.Template
	if err := r.db.WithContext(ctx).Where("is_active = true").Order("category, name").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("ServiceRepo.ListTemplates: %w", err)
	}
	return templates, nil
}

func (r *ServiceRepository) FindTemplateByID(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	var t models.Template
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error; err != nil {
		return nil, fmt.Errorf("ServiceRepo.FindTemplateByID: %w", err)
	}
	return &t, nil
}

// Services

func (r *ServiceRepository) Create(ctx context.Context, svc *models.DeployedService) error {
	if err := r.db.WithContext(ctx).Create(svc).Error; err != nil {
		return fmt.Errorf("ServiceRepo.Create: %w", err)
	}
	return nil
}

func (r *ServiceRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*models.DeployedService, error) {
	var svc models.DeployedService
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).Preload("Template").
		Where("id = ?", id).First(&svc).Error; err != nil {
		return nil, fmt.Errorf("ServiceRepo.FindByID: %w", err)
	}
	return &svc, nil
}

func (r *ServiceRepository) ListByOrgID(ctx context.Context, orgID uuid.UUID) ([]models.DeployedService, error) {
	var services []models.DeployedService
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).Preload("Template").
		Order("created_at DESC").Find(&services).Error; err != nil {
		return nil, fmt.Errorf("ServiceRepo.ListByOrgID: %w", err)
	}
	return services, nil
}

func (r *ServiceRepository) Update(ctx context.Context, svc *models.DeployedService) error {
	if err := r.db.WithContext(ctx).Save(svc).Error; err != nil {
		return fmt.Errorf("ServiceRepo.Update: %w", err)
	}
	return nil
}

func (r *ServiceRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("id = ?", id).Delete(&models.DeployedService{}).Error; err != nil {
		return fmt.Errorf("ServiceRepo.Delete: %w", err)
	}
	return nil
}
