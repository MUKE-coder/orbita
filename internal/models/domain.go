package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Domain struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ResourceID     uuid.UUID       `json:"resource_id" gorm:"type:uuid;not null;index"`
	ResourceType   string          `json:"resource_type" gorm:"not null"`
	OrganizationID uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index"`
	Domain         string          `json:"domain" gorm:"uniqueIndex;not null"`
	SSLEnabled     bool            `json:"ssl_enabled" gorm:"not null;default:true"`
	SSLConfig      json.RawMessage `json:"ssl_config" gorm:"type:jsonb;default:'{}'"`
	Status         string          `json:"status" gorm:"not null;default:pending"`
	Verified       bool            `json:"verified" gorm:"not null;default:false"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `json:"-" gorm:"index"`
}

const (
	DomainStatusPending  = "pending"
	DomainStatusActive   = "active"
	DomainStatusError    = "error"

	ResourceTypeApp      = "application"
	ResourceTypeDatabase = "database"
	ResourceTypeService  = "service"
)
