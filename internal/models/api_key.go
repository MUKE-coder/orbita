package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type APIKey struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID     uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	OrgID      *uuid.UUID     `json:"org_id" gorm:"type:uuid"`
	Name       string         `json:"name" gorm:"not null"`
	KeyHash    string         `json:"-" gorm:"not null;index"`
	KeyPrefix  string         `json:"key_prefix" gorm:"not null"`
	Scopes     pq.StringArray `json:"scopes" gorm:"type:text[];default:'{}'"`
	LastUsedAt *time.Time     `json:"last_used_at"`
	ExpiresAt  *time.Time     `json:"expires_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}
