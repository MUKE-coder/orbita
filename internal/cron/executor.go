package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

type Executor struct {
	cronRepo *repository.CronRepository
}

func NewExecutor(cronRepo *repository.CronRepository) *Executor {
	return &Executor{cronRepo: cronRepo}
}

func (e *Executor) ExecuteJob(ctx context.Context, job *models.CronJob) {
	log.Info().Str("job", job.Name).Str("job_id", job.ID.String()).Msg("Executing cron job")

	// Check concurrency policy
	if job.ConcurrencyPolicy == models.ConcurrencyForbid {
		hasRunning, err := e.cronRepo.HasRunningRun(ctx, job.ID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to check running runs")
			return
		}
		if hasRunning {
			log.Info().Str("job", job.Name).Msg("Skipping: previous run still active (policy: forbid)")
			e.createSkippedRun(ctx, job.ID)
			return
		}
	}

	now := time.Now()
	run := &models.CronRun{
		ID:        uuid.New(),
		CronJobID: job.ID,
		StartedAt: now,
		Status:    models.CronRunStatusRunning,
	}

	if err := e.cronRepo.CreateRun(ctx, run); err != nil {
		log.Error().Err(err).Msg("Failed to create cron run")
		return
	}

	// Execute with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(job.Timeout)*time.Second)
	defer cancel()

	err := e.executeRun(execCtx, job, run)

	finishedAt := time.Now()
	run.FinishedAt = &finishedAt
	duration := int(finishedAt.Sub(now).Milliseconds())
	run.DurationMs = &duration

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			run.Status = models.CronRunStatusTimeout
		} else {
			run.Status = models.CronRunStatusFailed
		}
		exitCode := 1
		run.ExitCode = &exitCode
		logSnip := err.Error()
		run.LogSnippet = &logSnip
		log.Error().Err(err).Str("job", job.Name).Msg("Cron job failed")
	} else {
		run.Status = models.CronRunStatusSuccess
		exitCode := 0
		run.ExitCode = &exitCode
		logSnip := fmt.Sprintf("Job %s completed successfully in %dms", job.Name, duration)
		run.LogSnippet = &logSnip
	}

	_ = e.cronRepo.UpdateRun(ctx, run)

	// Update job's last run time
	job.LastRunAt = &finishedAt
	_ = e.cronRepo.Update(ctx, job)
}

func (e *Executor) executeRun(ctx context.Context, job *models.CronJob, run *models.CronRun) error {
	// TODO: real impl
	// 1. Pull image if needed
	// 2. Create Docker container (not service — run-to-completion):
	//    docker run --rm --network orbita-org-{slug} --cgroup-parent orbita-org-{slug} ...
	// 3. Apply timeout context
	// 4. Stream and capture logs (store first 100KB in cron_runs.logs)
	// 5. Record exit code, duration, status
	// 6. On failure: retry N times with 5s backoff

	log.Info().
		Str("job", job.Name).
		Str("image", job.Image).
		Msg("Running cron container (stub)")

	// Simulate execution
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

func (e *Executor) createSkippedRun(ctx context.Context, jobID uuid.UUID) {
	now := time.Now()
	exitCode := -1
	logSnip := "Skipped: previous run still active"
	run := &models.CronRun{
		ID:         uuid.New(),
		CronJobID:  jobID,
		StartedAt:  now,
		FinishedAt: &now,
		Status:     models.CronRunStatusSkipped,
		ExitCode:   &exitCode,
		LogSnippet: &logSnip,
	}
	_ = e.cronRepo.CreateRun(ctx, run)
}
