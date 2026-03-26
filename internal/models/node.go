package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Node struct {
	ID            uuid.UUID       `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name          string          `json:"name" gorm:"not null"`
	IP            string          `json:"ip" gorm:"not null"`
	SSHPort       int             `json:"ssh_port" gorm:"not null;default:22"`
	SSHKeyID      *uuid.UUID      `json:"ssh_key_id" gorm:"type:uuid"`
	Role          string          `json:"role" gorm:"not null;default:worker"`
	Status        string          `json:"status" gorm:"not null;default:pending"`
	Labels        json.RawMessage `json:"labels" gorm:"type:jsonb;default:'{}'"`
	CPUCores      int             `json:"cpu_cores" gorm:"not null;default:0"`
	RAMMB         int             `json:"ram_mb" gorm:"not null;default:0"`
	DiskGB        int             `json:"disk_gb" gorm:"not null;default:0"`
	DockerVersion *string         `json:"docker_version"`
	SwarmNodeID   *string         `json:"swarm_node_id"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `json:"-" gorm:"index"`
}

const (
	NodeRolePrimary = "primary"
	NodeRoleWorker  = "worker"

	NodeStatusPending  = "pending"
	NodeStatusOnline   = "online"
	NodeStatusOffline  = "offline"
	NodeStatusDraining = "draining"
)

type NodeMetrics struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsed    int64   `json:"memory_used"`
	MemoryTotal   int64   `json:"memory_total"`
	DiskUsed      int64   `json:"disk_used"`
	DiskTotal     int64   `json:"disk_total"`
	ContainerCount int    `json:"container_count"`
	Uptime        int64   `json:"uptime_seconds"`
}
