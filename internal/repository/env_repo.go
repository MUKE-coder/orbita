package repository

import (
	"context"
	"errors"
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
	// Look up the existing row by the unique tuple (resource, key) — NOT by the
	// caller-supplied primary key, which is freshly generated on every set and
	// would make FirstOrCreate miss the existing row and INSERT a duplicate.
	var existing models.EnvVariable
	err := r.db.WithContext(ctx).
		Where("resource_id = ? AND resource_type = ? AND key = ? AND deleted_at IS NULL",
			ev.ResourceID, ev.ResourceType, ev.Key).
		First(&existing).Error

	if err == nil {
		existing.ValueEncrypted = ev.ValueEncrypted
		existing.IsSecret = ev.IsSecret
		existing.OrganizationID = ev.OrganizationID
		if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return fmt.Errorf("EnvRepo.Upsert update: %w", err)
		}
		*ev = existing
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("EnvRepo.Upsert lookup: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(ev).Error; err != nil {
		return fmt.Errorf("EnvRepo.Upsert create: %w", err)
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
