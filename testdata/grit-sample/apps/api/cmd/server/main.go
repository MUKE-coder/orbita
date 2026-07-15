package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"gritsample/apps/api/internal/config"
	"gritsample/apps/api/internal/database"
	"gritsample/apps/api/internal/models"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	// Health endpoint Orbita/Traefik use as the readiness gate.
	r.GET("/api/health", func(c *gin.Context) {
		sqlDB, e := db.DB()
		dbOK := e == nil && sqlDB.Ping() == nil
		status := "ok"
		if !dbOK {
			status = "degraded"
		}
		c.JSON(http.StatusOK, gin.H{"status": status, "database": gin.H{"ok": dbOK}})
	})

	r.GET("/api/notes", func(c *gin.Context) {
		var notes []models.Note
		db.Find(&notes)
		c.JSON(http.StatusOK, gin.H{"data": notes})
	})
	r.POST("/api/notes", func(c *gin.Context) {
		var n models.Note
		if err := c.ShouldBindJSON(&n); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Create(&n)
		c.JSON(http.StatusCreated, gin.H{"data": n})
	})

	log.Printf("grit-sample api listening on :%s", cfg.Port)
	_ = r.Run(":" + cfg.Port)
}
