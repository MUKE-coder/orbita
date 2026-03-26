package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/orbita-sh/orbita/internal/database"
)

func main() {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://orbita:orbita@localhost:5432/orbita?sslmode=disable"
	}

	migrationsPath := "migrations"

	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate [up|down]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "up":
		fmt.Println("Running migrations...")
		if err := database.RunMigrations(databaseURL, migrationsPath); err != nil {
			fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations applied successfully.")
	case "down":
		fmt.Println("Rolling back last migration...")
		if err := database.RollbackMigrations(databaseURL, migrationsPath); err != nil {
			fmt.Fprintf(os.Stderr, "Rollback failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Rollback completed successfully.")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s. Use 'up' or 'down'.\n", os.Args[1])
		os.Exit(1)
	}
}
