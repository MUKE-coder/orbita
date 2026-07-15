package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort          string
	DatabaseURL         string
	RedisURL            string
	JWTSecret           string
	JWTRefreshSecret    string
	EncryptionMasterKey string
	ResendAPIKey        string
	ResendFromEmail     string
	DockerSocket        string
	TraefikConfigDir    string
	CORSOrigins         string
	AppBaseURL          string
	IsProduction        bool
	SuperAdminEmail     string
	BackupDir           string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://orbita:orbita@localhost:5432/orbita?sslmode=disable"),
		RedisURL:            getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:           getEnv("JWT_SECRET", ""),
		JWTRefreshSecret:    getEnv("JWT_REFRESH_SECRET", ""),
		EncryptionMasterKey: getEnv("ENCRYPTION_MASTER_KEY", ""),
		ResendAPIKey:        getEnv("RESEND_API_KEY", ""),
		ResendFromEmail:     getEnv("RESEND_FROM_EMAIL", "orbita@localhost"),
		DockerSocket:        getEnv("DOCKER_SOCKET", "/var/run/docker.sock"),
		TraefikConfigDir:    getEnv("TRAEFIK_CONFIG_DIR", "/etc/orbita/traefik"),
		CORSOrigins:         getEnv("CORS_ORIGINS", "http://localhost:5173"),
		AppBaseURL:          getEnv("APP_BASE_URL", "http://localhost:8080"),
		IsProduction:        getEnvBool("IS_PRODUCTION", false),
		SuperAdminEmail:     getEnv("SUPER_ADMIN_EMAIL", ""),
		BackupDir:           getEnv("BACKUP_DIR", "/var/lib/orbita/backups"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.JWTRefreshSecret == "" {
		return nil, fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	// Every org encryption key is HKDF-derived from this value. An empty or
	// short key would silently weaken all tenant secrets at rest.
	if len(cfg.EncryptionMasterKey) < 32 {
		return nil, fmt.Errorf("ENCRYPTION_MASTER_KEY is required and must be at least 32 characters (generate with: openssl rand -hex 32)")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return b
}
