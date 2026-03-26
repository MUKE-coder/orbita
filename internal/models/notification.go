package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID         uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	OrganizationID *uuid.UUID `json:"organization_id" gorm:"type:uuid;index"`
	Type           string     `json:"type" gorm:"not null"`
	Title          string     `json:"title" gorm:"not null"`
	Body           *string    `json:"body"`
	Read           bool       `json:"read" gorm:"not null;default:false"`
	CreatedAt      time.Time  `json:"created_at"`
}

type NotificationSetting struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null"`
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	EventType      string    `json:"event_type" gorm:"not null"`
	EmailEnabled   bool      `json:"email_enabled" gorm:"not null;default:true"`
	WebhookURL     *string   `json:"webhook_url"`
	WebhookEnabled bool      `json:"webhook_enabled" gorm:"not null;default:false"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AuditLog struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrganizationID *uuid.UUID      `json:"organization_id" gorm:"type:uuid;index"`
	UserID         *uuid.UUID      `json:"user_id" gorm:"type:uuid;index"`
	Action         string          `json:"action" gorm:"not null"`
	ResourceType   *string         `json:"resource_type"`
	ResourceID     *uuid.UUID      `json:"resource_id" gorm:"type:uuid"`
	Metadata       json.RawMessage `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	IP             *string         `json:"ip"`
	CreatedAt      time.Time       `json:"created_at"`
}

const (
	NotifTypeDeploy      = "deploy"
	NotifTypeCronFailure = "cron_failure"
	NotifTypeBackup      = "backup"
	NotifTypeQuota       = "quota_alert"
	NotifTypeMember      = "member_change"
)
