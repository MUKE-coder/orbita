package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GitConnection struct {
	ID                    uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrganizationID        uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index"`
	Provider              string          `json:"provider" gorm:"not null"`
	AccessTokenEncrypted  string          `json:"-" gorm:"not null"`
	RefreshTokenEncrypted *string         `json:"-"`
	Metadata              json.RawMessage `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	DeletedAt             gorm.DeletedAt  `json:"-" gorm:"index"`
}

type RegistryCredential struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrganizationID    uuid.UUID      `json:"organization_id" gorm:"type:uuid;not null;index"`
	RegistryURL       string         `json:"registry_url" gorm:"not null"`
	Username          string         `json:"username" gorm:"not null"`
	PasswordEncrypted string         `json:"-" gorm:"not null"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	GitProviderGitHub = "github"
	GitProviderGitLab = "gitlab"
	GitProviderGitea  = "gitea"
)
