package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Email           string         `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash    string         `json:"-" gorm:"not null"`
	Name            string         `json:"name" gorm:"not null"`
	AvatarURL       *string        `json:"avatar_url"`
	IsSuperAdmin    bool           `json:"is_super_admin" gorm:"default:false;not null"`
	IsEmailVerified bool           `json:"is_email_verified" gorm:"default:false;not null"`
	TOTPSecret      *string        `json:"-"`
	TOTPEnabled     bool           `json:"totp_enabled" gorm:"default:false;not null"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}
