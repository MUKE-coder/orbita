package models

import (
	"time"

	"github.com/google/uuid"
)

// PlatformSettings holds instance-wide configuration the super admin can edit
// from the dashboard. Single row, id always 1 (enforced by a CHECK constraint).
//
// These override the equivalent environment variables so an operator can turn
// on email after install without editing compose and restarting.
type PlatformSettings struct {
	ID int `json:"-" gorm:"primaryKey;default:1"`

	// AES-256-GCM under the platform-derived key. Never serialised — the API
	// exposes only whether a key is present.
	ResendAPIKeyEnc *string `json:"-" gorm:"column:resend_api_key_enc"`

	EmailFrom     *string    `json:"email_from"`
	EmailFromName *string    `json:"email_from_name"`
	UpdatedBy     *uuid.UUID `json:"updated_by" gorm:"type:uuid"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (PlatformSettings) TableName() string { return "platform_settings" }
