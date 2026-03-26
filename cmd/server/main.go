package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	orbita "github.com/orbita-sh/orbita"
	"github.com/orbita-sh/orbita/internal/api"
	"github.com/orbita-sh/orbita/internal/config"
	"github.com/orbita-sh/orbita/internal/database"
	"github.com/orbita-sh/orbita/internal/mailer"
	orbitaRedis "github.com/orbita-sh/orbita/internal/redis"
	"github.com/orbita-sh/orbita/internal/repository"
	"github.com/orbita-sh/orbita/internal/service"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	if cfg.IsProduction {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx := context.Background()

	// Connect to database
	db, err := database.Connect(ctx, cfg.DatabaseURL, cfg.IsProduction)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	log.Info().Msg("Connected to PostgreSQL")

	// Run migrations
	if err := database.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	log.Info().Msg("Migrations applied")

	// Connect to Redis
	rdb, err := orbitaRedis.Connect(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	log.Info().Msg("Connected to Redis")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	orgRepo := repository.NewOrgRepository(db)

	// Initialize services
	mail := mailer.New(cfg.ResendAPIKey, cfg.ResendFromEmail)
	authService := service.NewAuthService(userRepo, mail, cfg)
	orgService := service.NewOrgService(orgRepo, userRepo, mail, cfg)

	// Prepare embedded static files
	staticFS, err := fs.Sub(orbita.StaticFiles, "web/dist")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to prepare static files")
	}

	// Setup router
	router := api.NewRouter(&api.RouterDeps{
		Config:      cfg,
		AuthService: authService,
		OrgService:  orgService,
		UserRepo:    userRepo,
		OrgRepo:     orgRepo,
		Redis:       rdb,
		StaticFS:    staticFS,
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router.Engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("port", cfg.ServerPort).Msg("Starting Orbita server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	// Close Redis
	if err := rdb.Close(); err != nil {
		log.Error().Err(err).Msg("Error closing Redis connection")
	}

	// Close database
	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing database connection")
		}
	}

	log.Info().Msg("Orbita server stopped")
}
