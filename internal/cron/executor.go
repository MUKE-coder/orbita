package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/docker"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

type Executor struct {
	cronRepo      *repository.CronRepository
	orgRepo       *repository.OrgRepository
	dockerClient  *docker.Client
	encryptionKey []byte
}

func NewExecutor(cronRepo *repository.CronRepository, orgRepo *repository.OrgRepository, dockerClient *docker.Client, encryptionKey []byte) *Executor {
	return &Executor{
		cronRepo:      cronRepo,
		orgRepo:       orgRepo,
		dockerClient:  dockerClient,
		encryptionKey: encryptionKey,
	}
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

	// Execute with timeout, retrying on failure per job config
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(job.Timeout)*time.Second)
	defer cancel()

	exitCode, logs, err := e.runContainer(execCtx, job)
	for attempt := 0; err == nil && exitCode != 0 && attempt < job.MaxRetries; attempt++ {
		select {
		case <-execCtx.Done():
			err = execCtx.Err()
		case <-time.After(5 * time.Second):
			log.Info().Str("job", job.Name).Int("attempt", attempt+2).Msg("Retrying cron job")
			exitCode, logs, err = e.runContainer(execCtx, job)
		}
	}

	finishedAt := time.Now()
	run.FinishedAt = &finishedAt
	duration := int(finishedAt.Sub(now).Milliseconds())
	run.DurationMs = &duration
	run.ExitCode = &exitCode

	snippet := logs
	if len(snippet) > 100*1024 {
		snippet = snippet[:100*1024]
	}

	switch {
	case execCtx.Err() == context.DeadlineExceeded:
		run.Status = models.CronRunStatusTimeout
		snippet = strings.TrimSpace(snippet + "\n[orbita] job killed: timeout reached")
	case err != nil:
		run.Status = models.CronRunStatusFailed
		snippet = strings.TrimSpace(snippet + "\n[orbita] " + err.Error())
		log.Error().Err(err).Str("job", job.Name).Msg("Cron job failed")
	case exitCode != 0:
		run.Status = models.CronRunStatusFailed
		log.Warn().Int("exit_code", exitCode).Str("job", job.Name).Msg("Cron job exited nonzero")
	default:
		run.Status = models.CronRunStatusSuccess
	}
	run.LogSnippet = &snippet

	_ = e.cronRepo.UpdateRun(ctx, run)

	// Update job's last run time
	job.LastRunAt = &finishedAt
	_ = e.cronRepo.Update(ctx, job)
}

// runContainer runs the job image to completion in the org's network and
// returns exit code + captured logs.
func (e *Executor) runContainer(ctx context.Context, job *models.CronJob) (int, string, error) {
	org, err := e.orgRepo.FindOrgByID(ctx, job.OrganizationID)
	if err != nil {
		return -1, "", fmt.Errorf("runContainer: resolve org: %w", err)
	}

	envVars, err := e.decryptEnvConfig(job)
	if err != nil {
		return -1, "", fmt.Errorf("runContainer: env: %w", err)
	}

	var cmd []string
	if job.Command != nil && strings.TrimSpace(*job.Command) != "" {
		cmd = []string{"/bin/sh", "-c", *job.Command}
	}

	spec := docker.OneOffSpec{
		Image:       job.Image,
		Command:     cmd,
		EnvVars:     envVars,
		NetworkName: docker.GetOrgNetworkName(org.Slug),
		Labels: map[string]string{
			"orbita.cron.id": job.ID.String(),
			"orbita.org":     org.Slug,
			"orbita.managed": "true",
		},
		MemoryLimit: int64(job.MemoryLimit) * 1024 * 1024,
		CPULimit:    int64(job.CPULimit) * 1_000_000, // 1000 = 1 core
	}

	return e.dockerClient.RunOneOffContainer(ctx, spec)
}

// decryptEnvConfig decodes the job's encrypted env JSON ({"KEY":"value"}).
func (e *Executor) decryptEnvConfig(job *models.CronJob) (map[string]string, error) {
	if job.EnvConfig == nil || *job.EnvConfig == "" {
		return nil, nil
	}
	orgKey, err := auth.DeriveOrgKey(e.encryptionKey, job.OrganizationID)
	if err != nil {
		return nil, err
	}
	raw := *job.EnvConfig
	if decrypted, err := auth.Decrypt(raw, orgKey); err == nil {
		raw = decrypted
	}
	var envVars map[string]string
	if err := json.Unmarshal([]byte(raw), &envVars); err != nil {
		return nil, fmt.Errorf("parse env config: %w", err)
	}
	return envVars, nil
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
