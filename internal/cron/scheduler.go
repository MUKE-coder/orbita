package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

type Scheduler struct {
	cron     *cron.Cron
	cronRepo *repository.CronRepository
	executor *Executor
	entries  map[uuid.UUID]cron.EntryID
	mu       sync.RWMutex
}

func NewScheduler(cronRepo *repository.CronRepository, executor *Executor) *Scheduler {
	return &Scheduler{
		cron:     cron.New(cron.WithSeconds()),
		cronRepo: cronRepo,
		executor: executor,
		entries:  make(map[uuid.UUID]cron.EntryID),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	// Load all enabled cron jobs
	jobs, err := s.cronRepo.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("Scheduler.Start: %w", err)
	}

	for i := range jobs {
		if err := s.AddJob(&jobs[i]); err != nil {
			log.Error().Err(err).Str("job_id", jobs[i].ID.String()).Msg("Failed to add cron job")
		}
	}

	s.cron.Start()
	log.Info().Int("jobs_loaded", len(jobs)).Msg("Cron scheduler started")

	return nil
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Info().Msg("Cron scheduler stopped")
}

func (s *Scheduler) AddJob(job *models.CronJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prepend "0 " for second field if schedule has 5 fields (standard cron)
	schedule := job.Schedule
	if len(schedule) > 0 && schedule[0] == '@' {
		// Named schedule like @daily, @hourly — use as-is
	} else {
		// Count fields
		fields := 0
		for _, c := range schedule {
			if c == ' ' {
				fields++
			}
		}
		if fields == 4 { // 5 fields → add seconds
			schedule = "0 " + schedule
		}
	}

	entryID, err := s.cron.AddFunc(schedule, func() {
		s.executor.ExecuteJob(context.Background(), job)
	})
	if err != nil {
		return fmt.Errorf("Scheduler.AddJob: %w", err)
	}

	s.entries[job.ID] = entryID

	// Update next run time
	entry := s.cron.Entry(entryID)
	nextRun := entry.Next
	job.NextRunAt = &nextRun

	log.Info().Str("job", job.Name).Str("schedule", job.Schedule).Time("next_run", nextRun).Msg("Cron job registered")

	return nil
}

func (s *Scheduler) RemoveJob(jobID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, ok := s.entries[jobID]; ok {
		s.cron.Remove(entryID)
		delete(s.entries, jobID)
	}
}

func (s *Scheduler) UpdateJob(job *models.CronJob) error {
	s.RemoveJob(job.ID)
	if job.Enabled {
		return s.AddJob(job)
	}
	return nil
}

func (s *Scheduler) TriggerJob(ctx context.Context, job *models.CronJob) error {
	go s.executor.ExecuteJob(ctx, job)
	return nil
}

func (s *Scheduler) GetNextRuns(jobID uuid.UUID, count int) []time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entryID, ok := s.entries[jobID]
	if !ok {
		return nil
	}

	entry := s.cron.Entry(entryID)
	var runs []time.Time
	next := entry.Next
	for i := 0; i < count; i++ {
		runs = append(runs, next)
		next = next.Add(entry.Next.Sub(time.Now()))
	}
	return runs
}
