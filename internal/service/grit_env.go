package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/grit"
	"github.com/orbita-sh/orbita/internal/models"
)

func jsonMarshal(v interface{}) ([]byte, error)   { return json.Marshal(v) }
func jsonUnmarshal(b []byte, v interface{}) error { return json.Unmarshal(b, v) }

// ComputeGritAPIEnv builds the full runtime env for a Grit API service, in
// precedence order (later wins): user env-file values → addon URLs → dashboard
// creds → platform overrides. This mirrors grit-knowledge/05: Orbita's
// in-network addon hosts and APP_ENV=production always win over the user's
// local .env.
func ComputeGritAPIEnv(m *grit.Manifest, userEnv, addonEnv map[string]string) map[string]string {
	env := map[string]string{}

	// 1. User-provided values from env.from (lowest precedence).
	for k, v := range userEnv {
		env[k] = v
	}

	// 2. Addon connection URLs (postgres/redis/minio) provisioned by Orbita.
	for k, v := range addonEnv {
		env[k] = v
	}

	// 3. Embedded dashboards: enable + strong generated passwords if absent.
	for k, v := range DashboardEnv(m, userEnv) {
		env[k] = v
	}

	// 4. Platform overrides (highest precedence).
	env["APP_ENV"] = "production"
	// Marker so the deploy engine knows whether to run the migration hook.
	if m.MigrateEnabled() {
		env["ORBITA_GRIT_MIGRATE"] = "true"
	} else {
		env["ORBITA_GRIT_MIGRATE"] = "false"
	}
	if env["PORT"] == "" {
		env["PORT"] = "8080"
	}
	if m.Domains.API != "" {
		env["API_URL"] = "https://" + m.Domains.API
	}
	if m.Domains.Admin != "" {
		env["NEXT_PUBLIC_ADMIN_URL"] = "https://" + m.Domains.Admin
	}
	if cors := corsOrigins(m); cors != "" {
		env["CORS_ORIGINS"] = cors
	}
	if m.Domains.Web != "" {
		env["OAUTH_FRONTEND_URL"] = "https://" + m.Domains.Web
	}

	return env
}

// corsOrigins builds a comma-separated allow-list from the app's public domains.
func corsOrigins(m *grit.Manifest) string {
	var origins []string
	for _, h := range []string{m.Domains.Web, m.Domains.Admin} {
		if h != "" {
			origins = append(origins, "https://"+h)
		}
	}
	return strings.Join(origins, ",")
}

// DashboardEnv returns the Pulse/Sentinel/Studio *_ENABLED + credential env for
// a Grit app, honoring orbita.yaml toggles and generating strong passwords when
// the user didn't supply them (never ship weak/default creds to production).
func DashboardEnv(m *grit.Manifest, userEnv map[string]string) map[string]string {
	env := map[string]string{}

	set := func(prefix string, enabled bool) {
		if !enabled {
			env[prefix+"_ENABLED"] = "false"
			return
		}
		env[prefix+"_ENABLED"] = "true"
		userField := prefix + "_USERNAME"
		passField := prefix + "_PASSWORD"
		if userEnv[userField] == "" {
			env[userField] = "admin"
		}
		if userEnv[passField] == "" {
			env[passField] = randomPassword(18)
		}
	}

	set("PULSE", m.ObservabilityEnabled())
	set("SENTINEL", m.SecurityEnabled())
	set("GORM_STUDIO", m.StudioEnabled())
	return env
}

// DashboardURLs returns the reachable dashboard links for the enabled dashboards
// (all served by the API on the api domain — grit-knowledge/07).
func DashboardURLs(m *grit.Manifest) map[string]string {
	if m.Domains.API == "" {
		return nil
	}
	base := "https://" + m.Domains.API
	urls := map[string]string{}
	if m.ObservabilityEnabled() {
		urls["pulse"] = base + "/pulse/ui"
	}
	if m.SecurityEnabled() {
		urls["sentinel"] = base + "/sentinel/ui"
	}
	if m.StudioEnabled() {
		urls["studio"] = base + "/studio"
	}
	urls["health"] = base + "/api/health"
	return urls
}

// provisionAddons ensures each requested addon exists in the org network and
// returns the env it contributes (grit-knowledge/05). Idempotent: existing
// addons (by deterministic name) are reused. Postgres/Redis reuse the managed-
// database path; MinIO uses the dedicated provisioner.
func (s *GritService) provisionAddons(ctx context.Context, orgID uuid.UUID, orgSlug string, envID uuid.UUID, m *grit.Manifest) (map[string]string, error) {
	env := map[string]string{}

	for _, addon := range m.Addons {
		switch addon {
		case "postgres":
			conn, err := s.ensureManagedDB(ctx, orgID, orgSlug, envID, m.App+"-postgres", models.EnginePostgres, "16")
			if err != nil {
				return nil, err
			}
			// DATABASE_URL wins over POSTGRES_* parts in Grit's config.
			env["DATABASE_URL"] = conn

		case "redis":
			conn, err := s.ensureManagedDB(ctx, orgID, orgSlug, envID, m.App+"-redis", models.EngineRedis, "7")
			if err != nil {
				return nil, err
			}
			env["REDIS_URL"] = conn

		case "minio":
			mr, err := s.ensureMinio(ctx, orgID, orgSlug, envID, m.App)
			if err != nil {
				return nil, err
			}
			env["STORAGE_DRIVER"] = "minio"
			env["MINIO_ENDPOINT"] = mr.Endpoint
			env["MINIO_ACCESS_KEY"] = mr.AccessKey
			env["MINIO_SECRET_KEY"] = mr.SecretKey
			env["MINIO_BUCKET"] = m.App + "-uploads"
			env["MINIO_USE_SSL"] = "false"

		default:
			log.Warn().Str("addon", addon).Msg("Unknown Grit addon; skipping")
		}
	}

	return env, nil
}

// ensureManagedDB finds an existing managed DB of the given name in the env or
// provisions one, returning its (decrypted) connection string.
func (s *GritService) ensureManagedDB(ctx context.Context, orgID uuid.UUID, orgSlug string, envID uuid.UUID, name, engine, version string) (string, error) {
	dbs, err := s.dbSvc.ListDatabases(ctx, orgID)
	if err != nil {
		return "", err
	}
	for _, db := range dbs {
		if db.EnvironmentID == envID && db.Name == name {
			return s.dbSvc.GetConnectionString(ctx, db.ID, orgID)
		}
	}

	mdb, err := s.dbSvc.CreateDatabase(ctx, orgID, orgSlug, CreateDBInput{
		Name:          name,
		Engine:        engine,
		Version:       version,
		EnvironmentID: envID,
	})
	if err != nil {
		return "", fmt.Errorf("provision %s: %w", engine, err)
	}
	return s.dbSvc.GetConnectionString(ctx, mdb.ID, orgID)
}

// ensureMinio finds an existing MinIO addon for the app (by service label) or
// provisions one. For idempotency it tracks the MinIO service via a managed
// resource marker in the app's addon set is overkill; instead we key on the
// deterministic service name and let provisioning be re-run-safe.
func (s *GritService) ensureMinio(ctx context.Context, orgID uuid.UUID, orgSlug string, envID uuid.UUID, appName string) (*minioAddon, error) {
	// MinIO isn't a managed database, so reuse-detection is best-effort: if a
	// prior reconcile stored its creds as env on the API app we'd reuse them,
	// but for simplicity and safety we (re)provision the service idempotently by
	// name — CreateService upserts by service name at the Docker layer.
	res, err := s.orch.ProvisionMinio(ctx, orgSlug, appName)
	if err != nil {
		return nil, err
	}
	return &minioAddon{Endpoint: res.Endpoint, AccessKey: res.AccessKey, SecretKey: res.SecretKey}, nil
}

type minioAddon struct {
	Endpoint  string
	AccessKey string
	SecretKey string
}
