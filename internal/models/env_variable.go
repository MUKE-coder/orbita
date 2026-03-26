package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EnvVariable struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ResourceID     uuid.UUID      `json:"resource_id" gorm:"type:uuid;not null"`
	ResourceType   string         `json:"resource_type" gorm:"not null"`
	OrganizationID uuid.UUID      `json:"organization_id" gorm:"type:uuid;not null;index"`
	Key            string         `json:"key" gorm:"not null"`
	ValueEncrypted string         `json:"-" gorm:"not null"`
	IsSecret       bool           `json:"is_secret" gorm:"not null;default:false"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// EnvVarDisplay is the API response struct with value handling
type EnvVarDisplay struct {
	ID           uuid.UUID `json:"id"`
	Key          string    `json:"key"`
	Value        string    `json:"value"`
	IsSecret     bool      `json:"is_secret"`
	ResourceID   uuid.UUID `json:"resource_id"`
	ResourceType string    `json:"resource_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
