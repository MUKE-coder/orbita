package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID           uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	RefreshTokenHash string    `json:"-" gorm:"not null;index"`
	DeviceInfo       *string   `json:"device_info"`
	IPAddress        *string   `json:"ip_address"`
	ExpiresAt        time.Time `json:"expires_at" gorm:"not null;index"`
	CreatedAt        time.Time `json:"created_at"`
}
