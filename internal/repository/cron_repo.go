package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type CronRepository struct {
	db *gorm.DB
}

func NewCronRepository(db *gorm.DB) *CronRepository {
	return &CronRepository{db: db}
}

func (r *CronRepository) Create(ctx context.Context, job *models.CronJob) error {
	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		return fmt.Errorf("CronRepo.Create: %w", err)
	}
	return nil
}

func (r *CronRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*models.CronJob, error) {
	var job models.CronJob
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("id = ?", id).First(&job).Error; err != nil {
		return nil, fmt.Errorf("CronRepo.FindByID: %w", err)
	}
	return &job, nil
}

func (r *CronRepository) ListByOrgID(ctx context.Context, orgID uuid.UUID) ([]models.CronJob, error) {
	var jobs []models.CronJob
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Order("created_at DESC").Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("CronRepo.ListByOrgID: %w", err)
	}
	return jobs, nil
}

func (r *CronRepository) ListEnabled(ctx context.Context) ([]models.CronJob, error) {
	var jobs []models.CronJob
	if err := r.db.WithContext(ctx).Where("enabled = true").Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("CronRepo.ListEnabled: %w", err)
	}
	return jobs, nil
}

func (r *CronRepository) Update(ctx context.Context, job *models.CronJob) error {
	if err := r.db.WithContext(ctx).Save(job).Error; err != nil {
		return fmt.Errorf("CronRepo.Update: %w", err)
	}
	return nil
}

func (r *CronRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("id = ?", id).Delete(&models.CronJob{}).Error; err != nil {
		return fmt.Errorf("CronRepo.Delete: %w", err)
	}
	return nil
}

// Runs

func (r *CronRepository) CreateRun(ctx context.Context, run *models.CronRun) error {
	if err := r.db.WithContext(ctx).Create(run).Error; err != nil {
		return fmt.Errorf("CronRepo.CreateRun: %w", err)
	}
	return nil
}

func (r *CronRepository) UpdateRun(ctx context.Context, run *models.CronRun) error {
	if err := r.db.WithContext(ctx).Save(run).Error; err != nil {
		return fmt.Errorf("CronRepo.UpdateRun: %w", err)
	}
	return nil
}

func (r *CronRepository) ListRuns(ctx context.Context, cronJobID uuid.UUID, limit int) ([]models.CronRun, error) {
	var runs []models.CronRun
	q := r.db.WithContext(ctx).Where("cron_job_id = ?", cronJobID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&runs).Error; err != nil {
		return nil, fmt.Errorf("CronRepo.ListRuns: %w", err)
	}
	return runs, nil
}

func (r *CronRepository) FindRunByID(ctx context.Context, id uuid.UUID) (*models.CronRun, error) {
	var run models.CronRun
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&run).Error; err != nil {
		return nil, fmt.Errorf("CronRepo.FindRunByID: %w", err)
	}
	return &run, nil
}

func (r *CronRepository) HasRunningRun(ctx context.Context, cronJobID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.CronRun{}).
		Where("cron_job_id = ? AND status = ?", cronJobID, models.CronRunStatusRunning).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("CronRepo.HasRunningRun: %w", err)
	}
	return count > 0, nil
}
