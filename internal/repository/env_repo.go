package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type EnvRepository struct {
	db *gorm.DB
}

func NewEnvRepository(db *gorm.DB) *EnvRepository {
	return &EnvRepository{db: db}
}

func (r *EnvRepository) Upsert(ctx context.Context, ev *models.EnvVariable) error {
	result := r.db.WithContext(ctx).
		Where("resource_id = ? AND resource_type = ? AND key = ? AND deleted_at IS NULL",
			ev.ResourceID, ev.ResourceType, ev.Key).
		Assign(models.EnvVariable{
			ValueEncrypted: ev.ValueEncrypted,
			IsSecret:       ev.IsSecret,
			OrganizationID: ev.OrganizationID,
		}).
		FirstOrCreate(ev)
	if result.Error != nil {
		return fmt.Errorf("EnvRepo.Upsert: %w", result.Error)
	}
	// If record existed, update it
	if result.RowsAffected == 0 {
		if err := r.db.WithContext(ctx).Save(ev).Error; err != nil {
			return fmt.Errorf("EnvRepo.Upsert save: %w", err)
		}
	}
	return nil
}

func (r *EnvRepository) ListByResource(ctx context.Context, resourceID uuid.UUID, resourceType string) ([]models.EnvVariable, error) {
	var vars []models.EnvVariable
	if err := r.db.WithContext(ctx).
		Where("resource_id = ? AND resource_type = ?", resourceID, resourceType).
		Order("key ASC").Find(&vars).Error; err != nil {
		return nil, fmt.Errorf("EnvRepo.ListByResource: %w", err)
	}
	return vars, nil
}

func (r *EnvRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.EnvVariable, error) {
	var ev models.EnvVariable
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&ev).Error; err != nil {
		return nil, fmt.Errorf("EnvRepo.FindByID: %w", err)
	}
	return &ev, nil
}

func (r *EnvRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("id = ?", id).Delete(&models.EnvVariable{}).Error; err != nil {
		return fmt.Errorf("EnvRepo.Delete: %w", err)
	}
	return nil
}

func (r *EnvRepository) DeleteByKey(ctx context.Context, resourceID uuid.UUID, resourceType, key string) error {
	if err := r.db.WithContext(ctx).
		Where("resource_id = ? AND resource_type = ? AND key = ?", resourceID, resourceType, key).
		Delete(&models.EnvVariable{}).Error; err != nil {
		return fmt.Errorf("EnvRepo.DeleteByKey: %w", err)
	}
	return nil
}
