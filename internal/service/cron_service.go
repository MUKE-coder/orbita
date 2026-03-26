package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/cron"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

var (
	ErrCronJobNotFound = errors.New("cron job not found")
	ErrCronRunNotFound = errors.New("cron run not found")
)

type CronService struct {
	cronRepo  *repository.CronRepository
	scheduler *cron.Scheduler
}

func NewCronService(cronRepo *repository.CronRepository, scheduler *cron.Scheduler) *CronService {
	return &CronService{
		cronRepo:  cronRepo,
		scheduler: scheduler,
	}
}

type CreateCronInput struct {
	Name              string    `json:"name"`
	Schedule          string    `json:"schedule"`
	Image             string    `json:"image"`
	Command           *string   `json:"command"`
	EnvironmentID     uuid.UUID `json:"environment_id"`
	Timeout           int       `json:"timeout"`
	ConcurrencyPolicy string   `json:"concurrency_policy"`
	MaxRetries        int       `json:"max_retries"`
	CPULimit          int       `json:"cpu_limit"`
	MemoryLimit       int       `json:"memory_limit"`
}

func (s *CronService) CreateCronJob(ctx context.Context, orgID uuid.UUID, input CreateCronInput) (*models.CronJob, error) {
	if input.Timeout <= 0 {
		input.Timeout = 3600
	}
	if input.ConcurrencyPolicy == "" {
		input.ConcurrencyPolicy = models.ConcurrencyForbid
	}

	job := &models.CronJob{
		ID:                uuid.New(),
		EnvironmentID:     input.EnvironmentID,
		OrganizationID:    orgID,
		Name:              input.Name,
		Schedule:          input.Schedule,
		Image:             input.Image,
		Command:           input.Command,
		Timeout:           input.Timeout,
		ConcurrencyPolicy: input.ConcurrencyPolicy,
		MaxRetries:        input.MaxRetries,
		CPULimit:          input.CPULimit,
		MemoryLimit:       input.MemoryLimit,
		Enabled:           true,
	}

	if err := s.cronRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("CreateCronJob: %w", err)
	}

	// Register with scheduler
	if err := s.scheduler.AddJob(job); err != nil {
		return nil, fmt.Errorf("CreateCronJob: schedule: %w", err)
	}

	_ = s.cronRepo.Update(ctx, job)

	return job, nil
}

func (s *CronService) GetCronJob(ctx context.Context, id, orgID uuid.UUID) (*models.CronJob, error) {
	job, err := s.cronRepo.FindByID(ctx, id, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCronJobNotFound
		}
		return nil, fmt.Errorf("GetCronJob: %w", err)
	}
	return job, nil
}

func (s *CronService) ListCronJobs(ctx context.Context, orgID uuid.UUID) ([]models.CronJob, error) {
	return s.cronRepo.ListByOrgID(ctx, orgID)
}

func (s *CronService) UpdateCronJob(ctx context.Context, id, orgID uuid.UUID, updates map[string]interface{}) (*models.CronJob, error) {
	job, err := s.cronRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, ErrCronJobNotFound
	}

	if name, ok := updates["name"].(string); ok && name != "" {
		job.Name = name
	}
	if schedule, ok := updates["schedule"].(string); ok && schedule != "" {
		job.Schedule = schedule
	}
	if image, ok := updates["image"].(string); ok && image != "" {
		job.Image = image
	}
	if timeout, ok := updates["timeout"].(float64); ok {
		job.Timeout = int(timeout)
	}
	if policy, ok := updates["concurrency_policy"].(string); ok {
		job.ConcurrencyPolicy = policy
	}
	if retries, ok := updates["max_retries"].(float64); ok {
		job.MaxRetries = int(retries)
	}

	if err := s.cronRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("UpdateCronJob: %w", err)
	}

	_ = s.scheduler.UpdateJob(job)
	return job, nil
}

func (s *CronService) DeleteCronJob(ctx context.Context, id, orgID uuid.UUID) error {
	_, err := s.cronRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return ErrCronJobNotFound
	}
	s.scheduler.RemoveJob(id)
	return s.cronRepo.Delete(ctx, id, orgID)
}

func (s *CronService) ToggleCronJob(ctx context.Context, id, orgID uuid.UUID) (*models.CronJob, error) {
	job, err := s.cronRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, ErrCronJobNotFound
	}

	job.Enabled = !job.Enabled
	if err := s.cronRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("ToggleCronJob: %w", err)
	}

	_ = s.scheduler.UpdateJob(job)
	return job, nil
}

func (s *CronService) TriggerCronJob(ctx context.Context, id, orgID uuid.UUID) error {
	job, err := s.cronRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return ErrCronJobNotFound
	}
	return s.scheduler.TriggerJob(ctx, job)
}

func (s *CronService) ListRuns(ctx context.Context, cronJobID uuid.UUID, limit int) ([]models.CronRun, error) {
	return s.cronRepo.ListRuns(ctx, cronJobID, limit)
}

func (s *CronService) GetRunLogs(ctx context.Context, runID uuid.UUID) (string, error) {
	run, err := s.cronRepo.FindRunByID(ctx, runID)
	if err != nil {
		return "", ErrCronRunNotFound
	}
	if run.LogSnippet == nil {
		return "", nil
	}
	return *run.LogSnippet, nil
}
