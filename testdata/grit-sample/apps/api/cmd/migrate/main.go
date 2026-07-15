package main

import (
	"log"

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
	log.Println("Running migrations...")
	if err := models.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("Migrations applied.")
}
