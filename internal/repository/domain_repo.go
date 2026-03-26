package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type DomainRepository struct {
	db *gorm.DB
}

func NewDomainRepository(db *gorm.DB) *DomainRepository {
	return &DomainRepository{db: db}
}

func (r *DomainRepository) Create(ctx context.Context, d *models.Domain) error {
	if err := r.db.WithContext(ctx).Create(d).Error; err != nil {
		return fmt.Errorf("DomainRepo.Create: %w", err)
	}
	return nil
}

func (r *DomainRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Domain, error) {
	var d models.Domain
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&d).Error; err != nil {
		return nil, fmt.Errorf("DomainRepo.FindByID: %w", err)
	}
	return &d, nil
}

func (r *DomainRepository) FindByDomain(ctx context.Context, domain string) (*models.Domain, error) {
	var d models.Domain
	if err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&d).Error; err != nil {
		return nil, fmt.Errorf("DomainRepo.FindByDomain: %w", err)
	}
	return &d, nil
}

func (r *DomainRepository) ListByResource(ctx context.Context, resourceID uuid.UUID, resourceType string) ([]models.Domain, error) {
	var domains []models.Domain
	if err := r.db.WithContext(ctx).Where("resource_id = ? AND resource_type = ?", resourceID, resourceType).
		Order("created_at DESC").Find(&domains).Error; err != nil {
		return nil, fmt.Errorf("DomainRepo.ListByResource: %w", err)
	}
	return domains, nil
}

func (r *DomainRepository) ListByOrgID(ctx context.Context, orgID uuid.UUID) ([]models.Domain, error) {
	var domains []models.Domain
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Order("created_at DESC").Find(&domains).Error; err != nil {
		return nil, fmt.Errorf("DomainRepo.ListByOrgID: %w", err)
	}
	return domains, nil
}

func (r *DomainRepository) Update(ctx context.Context, d *models.Domain) error {
	if err := r.db.WithContext(ctx).Save(d).Error; err != nil {
		return fmt.Errorf("DomainRepo.Update: %w", err)
	}
	return nil
}

func (r *DomainRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Domain{}).Error; err != nil {
		return fmt.Errorf("DomainRepo.Delete: %w", err)
	}
	return nil
}
