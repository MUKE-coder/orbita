package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ManagedDatabase struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	EnvironmentID    uuid.UUID      `json:"environment_id" gorm:"type:uuid;not null;index"`
	OrganizationID   uuid.UUID      `json:"organization_id" gorm:"type:uuid;not null;index"`
	Name             string         `json:"name" gorm:"not null"`
	Engine           string         `json:"engine" gorm:"not null"`
	Version          string         `json:"version" gorm:"not null"`
	ConnectionConfig *string        `json:"-"`
	VolumeName       *string        `json:"volume_name"`
	DockerServiceID  *string        `json:"docker_service_id"`
	Status           string         `json:"status" gorm:"not null;default:creating"`
	Port             *int           `json:"port"`
	CPULimit         int            `json:"cpu_limit" gorm:"not null;default:0"`
	MemoryLimit      int            `json:"memory_limit" gorm:"not null;default:0"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

type Backup struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	SourceID       uuid.UUID  `json:"source_id" gorm:"type:uuid;not null;index"`
	SourceType     string     `json:"source_type" gorm:"not null"`
	OrganizationID uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null;index"`
	Status         string     `json:"status" gorm:"not null;default:pending"`
	SizeBytes      int64      `json:"size_bytes" gorm:"not null;default:0"`
	StoragePath    *string    `json:"storage_path"`
	ErrorMessage   *string    `json:"error_message"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
}

type BackupSchedule struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	SourceID          uuid.UUID  `json:"source_id" gorm:"type:uuid;not null;index"`
	SourceType        string     `json:"source_type" gorm:"not null"`
	OrganizationID    uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null"`
	Frequency         string     `json:"frequency" gorm:"not null;default:daily"`
	RetentionCount    int        `json:"retention_count" gorm:"not null;default:7"`
	DestinationConfig *string    `json:"-"`
	Enabled           bool       `json:"enabled" gorm:"not null;default:true"`
	LastRunAt         *time.Time `json:"last_run_at"`
	NextRunAt         *time.Time `json:"next_run_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

const (
	EnginePostgres = "postgres"
	EngineMySQL    = "mysql"
	EngineMariaDB  = "mariadb"
	EngineMongoDB  = "mongodb"
	EngineRedis    = "redis"

	DBStatusCreating = "creating"
	DBStatusRunning  = "running"
	DBStatusStopped  = "stopped"
	DBStatusFailed   = "failed"

	BackupStatusPending    = "pending"
	BackupStatusRunning    = "running"
	BackupStatusCompleted  = "completed"
	BackupStatusFailed     = "failed"
)

var EngineImages = map[string]map[string]string{
	EnginePostgres: {"15": "postgres:15-alpine", "16": "postgres:16-alpine"},
	EngineMySQL:    {"8": "mysql:8"},
	EngineMariaDB:  {"10": "mariadb:10", "11": "mariadb:11"},
	EngineMongoDB:  {"6": "mongo:6", "7": "mongo:7"},
	EngineRedis:    {"7": "redis:7-alpine"},
}
