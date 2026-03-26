package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Application struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	EnvironmentID   uuid.UUID       `json:"environment_id" gorm:"type:uuid;not null;index"`
	OrganizationID  uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index"`
	Name            string          `json:"name" gorm:"not null"`
	SourceType      string          `json:"source_type" gorm:"not null;default:docker-image"`
	SourceConfig    json.RawMessage `json:"source_config" gorm:"type:jsonb;default:'{}'"`
	BuildConfig     json.RawMessage `json:"build_config" gorm:"type:jsonb;default:'{}'"`
	DeployConfig    json.RawMessage `json:"deploy_config" gorm:"type:jsonb;default:'{}'"`
	Status          string          `json:"status" gorm:"not null;default:created"`
	DockerServiceID *string         `json:"docker_service_id"`
	Replicas        int             `json:"replicas" gorm:"not null;default:1"`
	Port            *int            `json:"port"`
	AutoDeploy      bool            `json:"auto_deploy" gorm:"not null;default:false"`
	WebhookSecret   *string         `json:"-"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `json:"-" gorm:"index"`
}

type Deployment struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	AppID        uuid.UUID  `json:"app_id" gorm:"type:uuid;not null;index"`
	Version      int        `json:"version" gorm:"not null;default:1"`
	ImageRef     string     `json:"image_ref" gorm:"not null"`
	DeployConfig json.RawMessage `json:"deploy_config" gorm:"type:jsonb;default:'{}'"`
	Status       string     `json:"status" gorm:"not null;default:pending"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	TriggeredBy  *uuid.UUID `json:"triggered_by" gorm:"type:uuid"`
	TriggerType  string     `json:"trigger_type" gorm:"not null;default:manual"`
	ErrorMessage *string    `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
}

const (
	AppStatusCreated   = "created"
	AppStatusDeploying = "deploying"
	AppStatusRunning   = "running"
	AppStatusStopped   = "stopped"
	AppStatusFailed    = "failed"
	AppStatusRemoving  = "removing"

	DeployStatusPending  = "pending"
	DeployStatusBuilding = "building"
	DeployStatusRunning  = "running"
	DeployStatusSuccess  = "success"
	DeployStatusFailed   = "failed"
	DeployStatusRolledBack = "rolled_back"

	SourceTypeDockerImage = "docker-image"
	SourceTypeGit         = "git"
	SourceTypeCompose     = "docker-compose"

	TriggerManual  = "manual"
	TriggerWebhook = "webhook"
	TriggerPush    = "push"
)
