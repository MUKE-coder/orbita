package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Template struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name            string          `json:"name" gorm:"not null"`
	Description     *string         `json:"description"`
	Category        string          `json:"category" gorm:"not null;default:other"`
	ComposeTemplate string          `json:"compose_template" gorm:"not null"`
	ParamsSchema    json.RawMessage `json:"params_schema" gorm:"type:jsonb;default:'[]'"`
	IconURL         *string         `json:"icon_url"`
	IsActive        bool            `json:"is_active" gorm:"not null;default:true"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type TemplateParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Default  string `json:"default"`
	Required bool   `json:"required"`
}

type DeployedService struct {
	ID               uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey;column:id"`
	EnvironmentID    uuid.UUID       `json:"environment_id" gorm:"type:uuid;not null"`
	OrganizationID   uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index"`
	TemplateID       *uuid.UUID      `json:"template_id" gorm:"type:uuid"`
	Name             string          `json:"name" gorm:"not null"`
	Config           json.RawMessage `json:"config" gorm:"type:jsonb;default:'{}'"`
	Status           string          `json:"status" gorm:"not null;default:creating"`
	DockerServiceIDs pq.StringArray  `json:"docker_service_ids" gorm:"type:text[];default:'{}'"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `json:"-" gorm:"index"`
	Template         *Template       `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
}

func (DeployedService) TableName() string {
	return "services"
}

const (
	ServiceStatusCreating = "creating"
	ServiceStatusRunning  = "running"
	ServiceStatusStopped  = "stopped"
	ServiceStatusFailed   = "failed"
)
