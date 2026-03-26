package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type DBRepository struct {
	db *gorm.DB
}

func NewDBRepository(db *gorm.DB) *DBRepository {
	return &DBRepository{db: db}
}

func (r *DBRepository) Create(ctx context.Context, mdb *models.ManagedDatabase) error {
	if err := r.db.WithContext(ctx).Create(mdb).Error; err != nil {
		return fmt.Errorf("DBRepo.Create: %w", err)
	}
	return nil
}

func (r *DBRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*models.ManagedDatabase, error) {
	var mdb models.ManagedDatabase
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("id = ?", id).First(&mdb).Error; err != nil {
		return nil, fmt.Errorf("DBRepo.FindByID: %w", err)
	}
	return &mdb, nil
}

func (r *DBRepository) ListByOrgID(ctx context.Context, orgID uuid.UUID) ([]models.ManagedDatabase, error) {
	var dbs []models.ManagedDatabase
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Order("created_at DESC").Find(&dbs).Error; err != nil {
		return nil, fmt.Errorf("DBRepo.ListByOrgID: %w", err)
	}
	return dbs, nil
}

func (r *DBRepository) Update(ctx context.Context, mdb *models.ManagedDatabase) error {
	if err := r.db.WithContext(ctx).Save(mdb).Error; err != nil {
		return fmt.Errorf("DBRepo.Update: %w", err)
	}
	return nil
}

func (r *DBRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("id = ?", id).Delete(&models.ManagedDatabase{}).Error; err != nil {
		return fmt.Errorf("DBRepo.Delete: %w", err)
	}
	return nil
}

// Backups

func (r *DBRepository) CreateBackup(ctx context.Context, b *models.Backup) error {
	if err := r.db.WithContext(ctx).Create(b).Error; err != nil {
		return fmt.Errorf("DBRepo.CreateBackup: %w", err)
	}
	return nil
}

func (r *DBRepository) FindBackupByID(ctx context.Context, id uuid.UUID) (*models.Backup, error) {
	var b models.Backup
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&b).Error; err != nil {
		return nil, fmt.Errorf("DBRepo.FindBackupByID: %w", err)
	}
	return &b, nil
}

func (r *DBRepository) ListBackups(ctx context.Context, sourceID uuid.UUID, limit int) ([]models.Backup, error) {
	var backups []models.Backup
	q := r.db.WithContext(ctx).Where("source_id = ?", sourceID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&backups).Error; err != nil {
		return nil, fmt.Errorf("DBRepo.ListBackups: %w", err)
	}
	return backups, nil
}

func (r *DBRepository) UpdateBackup(ctx context.Context, b *models.Backup) error {
	if err := r.db.WithContext(ctx).Save(b).Error; err != nil {
		return fmt.Errorf("DBRepo.UpdateBackup: %w", err)
	}
	return nil
}

// Backup Schedules

func (r *DBRepository) GetBackupSchedule(ctx context.Context, sourceID uuid.UUID) (*models.BackupSchedule, error) {
	var bs models.BackupSchedule
	if err := r.db.WithContext(ctx).Where("source_id = ?", sourceID).First(&bs).Error; err != nil {
		return nil, fmt.Errorf("DBRepo.GetBackupSchedule: %w", err)
	}
	return &bs, nil
}

func (r *DBRepository) UpsertBackupSchedule(ctx context.Context, bs *models.BackupSchedule) error {
	if err := r.db.WithContext(ctx).Save(bs).Error; err != nil {
		return fmt.Errorf("DBRepo.UpsertBackupSchedule: %w", err)
	}
	return nil
}
