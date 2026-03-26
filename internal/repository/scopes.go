package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func OrgScope(orgID uuid.UUID) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("organization_id = ?", orgID)
	}
}
