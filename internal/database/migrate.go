package database

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURL, migrationsPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("database.RunMigrations: create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("database.RunMigrations: run up: %w", err)
	}

	return nil
}

func RollbackMigrations(databaseURL, migrationsPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("database.RollbackMigrations: create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("database.RollbackMigrations: step down: %w", err)
	}

	return nil
}
