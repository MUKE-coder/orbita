package models

import (
	"time"

	"gorm.io/gorm"
)

// Note is a minimal Grit-style resource model.
type Note struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// Models returns every model to migrate (Grit's models.Models()).
func Models() []interface{} { return []interface{}{&Note{}} }

// Migrate runs GORM AutoMigrate over the model registry (idempotent, additive).
func Migrate(db *gorm.DB) error { return db.AutoMigrate(Models()...) }
