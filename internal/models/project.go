package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrganizationID uuid.UUID      `json:"organization_id" gorm:"type:uuid;not null;index"`
	Name           string         `json:"name" gorm:"not null"`
	Description    *string        `json:"description"`
	Emoji          string         `json:"emoji" gorm:"default:'🚀'"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	Environments   []Environment  `json:"environments,omitempty" gorm:"foreignKey:ProjectID"`
}

type Environment struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ProjectID uuid.UUID      `json:"project_id" gorm:"type:uuid;not null;index"`
	Name      string         `json:"name" gorm:"not null"`
	Type      string         `json:"type" gorm:"not null;default:production"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	EnvTypeProduction = "production"
	EnvTypeStaging    = "staging"
	EnvTypeCustom     = "custom"
)
