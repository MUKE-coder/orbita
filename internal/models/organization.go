package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Organization struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Slug        string         `json:"slug" gorm:"uniqueIndex;not null"`
	Description *string        `json:"description"`
	PlanID      *uuid.UUID     `json:"plan_id" gorm:"type:uuid"`
	Plan        *ResourcePlan  `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
	CreatedBy   uuid.UUID      `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Members     []OrgMember    `json:"members,omitempty" gorm:"foreignKey:OrgID"`
}

type OrgMember struct {
	OrgID    uuid.UUID `json:"org_id" gorm:"type:uuid;primaryKey"`
	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	Role     string    `json:"role" gorm:"not null;default:viewer"`
	JoinedAt time.Time `json:"joined_at"`
	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type OrgInvite struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrgID     uuid.UUID  `json:"org_id" gorm:"type:uuid;not null;index"`
	Email     string     `json:"email" gorm:"not null"`
	Role      string     `json:"role" gorm:"not null;default:developer"`
	TokenHash string     `json:"-" gorm:"not null;index"`
	InvitedBy uuid.UUID  `json:"invited_by" gorm:"type:uuid;not null"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
	Inviter   *User      `json:"inviter,omitempty" gorm:"foreignKey:InvitedBy"`
}

type ResourcePlan struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name         string         `json:"name" gorm:"uniqueIndex;not null"`
	MaxCPUCores  int            `json:"max_cpu_cores" gorm:"not null;default:1"`
	MaxRAMMB     int            `json:"max_ram_mb" gorm:"not null;default:1024"`
	MaxDiskGB    int            `json:"max_disk_gb" gorm:"not null;default:10"`
	MaxApps      int            `json:"max_apps" gorm:"not null;default:5"`
	MaxDatabases int            `json:"max_databases" gorm:"not null;default:3"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

const (
	RoleOwner     = "owner"
	RoleAdmin     = "admin"
	RoleDeveloper = "developer"
	RoleViewer    = "viewer"
)

var RoleHierarchy = map[string]int{
	RoleViewer:    0,
	RoleDeveloper: 1,
	RoleAdmin:     2,
	RoleOwner:     3,
}

func HasMinRole(userRole, minRole string) bool {
	return RoleHierarchy[userRole] >= RoleHierarchy[minRole]
}
