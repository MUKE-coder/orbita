package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type SettingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get returns the single settings row, creating it if the migration's seed is
// missing (an instance upgraded mid-flight, say).
func (r *SettingsRepository) Get(ctx context.Context) (*models.PlatformSettings, error) {
	var s models.PlatformSettings
	err := r.db.WithContext(ctx).First(&s, 1).Error
	if err == gorm.ErrRecordNotFound {
		s = models.PlatformSettings{ID: 1}
		if err := r.db.WithContext(ctx).Create(&s).Error; err != nil {
			return nil, fmt.Errorf("Get settings: seed: %w", err)
		}
		return &s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Get settings: %w", err)
	}
	return &s, nil
}

// Update writes the settings row. Uses Select so explicit nils (clearing the
// API key) are persisted rather than skipped as zero values.
func (r *SettingsRepository) Update(ctx context.Context, s *models.PlatformSettings) error {
	s.ID = 1
	err := r.db.WithContext(ctx).
		Model(&models.PlatformSettings{ID: 1}).
		Select("resend_api_key_enc", "email_from", "email_from_name", "updated_by", "updated_at").
		Updates(s).Error
	if err != nil {
		return fmt.Errorf("Update settings: %w", err)
	}
	return nil
}
