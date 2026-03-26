package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CronJob struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	EnvironmentID     uuid.UUID      `json:"environment_id" gorm:"type:uuid;not null"`
	OrganizationID    uuid.UUID      `json:"organization_id" gorm:"type:uuid;not null;index"`
	Name              string         `json:"name" gorm:"not null"`
	Schedule          string         `json:"schedule" gorm:"not null"`
	Image             string         `json:"image" gorm:"not null"`
	Command           *string        `json:"command"`
	EnvConfig         *string        `json:"-"`
	Timeout           int            `json:"timeout" gorm:"not null;default:3600"`
	ConcurrencyPolicy string         `json:"concurrency_policy" gorm:"not null;default:forbid"`
	MaxRetries        int            `json:"max_retries" gorm:"not null;default:0"`
	CPULimit          int            `json:"cpu_limit" gorm:"not null;default:0"`
	MemoryLimit       int            `json:"memory_limit" gorm:"not null;default:0"`
	Enabled           bool           `json:"enabled" gorm:"not null;default:true"`
	LastRunAt         *time.Time     `json:"last_run_at"`
	NextRunAt         *time.Time     `json:"next_run_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

type CronRun struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CronJobID  uuid.UUID  `json:"cron_job_id" gorm:"type:uuid;not null;index"`
	StartedAt  time.Time  `json:"started_at" gorm:"not null"`
	FinishedAt *time.Time `json:"finished_at"`
	Status     string     `json:"status" gorm:"not null;default:running"`
	ExitCode   *int       `json:"exit_code"`
	LogSnippet *string    `json:"log_snippet"`
	DurationMs *int       `json:"duration_ms"`
	CreatedAt  time.Time  `json:"created_at"`
}

const (
	CronRunStatusRunning  = "running"
	CronRunStatusSuccess  = "success"
	CronRunStatusFailed   = "failed"
	CronRunStatusTimeout  = "timeout"
	CronRunStatusSkipped  = "skipped"

	ConcurrencyAllow   = "allow"
	ConcurrencyForbid  = "forbid"
	ConcurrencyReplace = "replace"
)
