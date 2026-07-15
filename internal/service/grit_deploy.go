package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/grit"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/orchestrator"
)

// GritDeployResult reports the outcome of a Grit deploy.
type GritDeployResult struct {
	GritApp      string            `json:"grit_app"`
	Migrated     bool              `json:"migrated"`
	Services     []GritServiceLink `json:"services"`
	LiveURL      string            `json:"live_url,omitempty"`
	APIURL       string            `json:"api_url,omitempty"`
	DashboardURL map[string]string `json:"dashboard_urls,omitempty"`
}

// GritServiceLink is one deployed service's status.
type GritServiceLink struct {
	Role   string `json:"role"`
	AppID  string `json:"app_id"`
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}

// Deploy builds and deploys every service of a reconciled Grit app in the
// correct order: the API first (so its image exists for the migration hook),
// then migrations under an advisory lock (gating cutover), then the front-ends.
// A migration failure aborts before any front-end cuts over.
func (s *GritService) Deploy(ctx context.Context, orgID uuid.UUID, orgSlug, gritApp string, envID uuid.UUID, userID *uuid.UUID) (*GritDeployResult, error) {
	services, err := s.appRepo.ListGritServices(ctx, envID, orgID, gritApp)
	if err != nil {
		return nil, fmt.Errorf("Deploy: list services: %w", err)
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("Deploy: no services found for grit app %q — reconcile first", gritApp)
	}

	// Deploy the API service first.
	ordered := orderServices(services)
	result := &GritDeployResult{GritApp: gritApp}

	var apiApp *models.Application
	for i := range ordered {
		if ordered[i].GritRole != nil && isAPIRole(*ordered[i].GritRole) {
			apiApp = &ordered[i]
			break
		}
	}
	if apiApp == nil {
		return nil, fmt.Errorf("Deploy: grit app %q has no API service", gritApp)
	}

	// 1. Deploy the API.
	if _, err := s.appSvc.DeployWithTrigger(ctx, apiApp.ID, orgID, orgSlug, userID, models.TriggerManual); err != nil {
		return nil, fmt.Errorf("Deploy: api build/deploy failed: %w", err)
	}

	// 2. Migrations under advisory lock, before front-ends cut over.
	migrated := false
	if s.shouldMigrate(ctx, orgID, apiApp) {
		if err := s.runMigrations(ctx, orgID, orgSlug, apiApp); err != nil {
			return nil, fmt.Errorf("Deploy: migrations failed (not cutting over): %w", err)
		}
		migrated = true
	}
	result.Migrated = migrated

	// 3. Deploy the front-ends (web/admin/docs).
	for i := range ordered {
		app := &ordered[i]
		if app.ID == apiApp.ID {
			continue
		}
		if _, err := s.appSvc.DeployWithTrigger(ctx, app.ID, orgID, orgSlug, userID, models.TriggerManual); err != nil {
			log.Error().Err(err).Str("role", roleOf(app)).Msg("Grit deploy: front-end failed")
			result.Services = append(result.Services, GritServiceLink{Role: roleOf(app), AppID: app.ID.String(), Status: "failed"})
			continue
		}
		result.Services = append(result.Services, s.serviceLink(ctx, orgID, app))
	}
	// API link (deployed first)
	result.Services = append(result.Services, s.serviceLink(ctx, orgID, apiApp))

	// 4. Public links.
	s.fillLinks(ctx, orgID, gritApp, envID, result)
	return result, nil
}

// shouldMigrate reports whether the migration hook should run: the app's stored
// grit migrate flag (default true) — read from the API app's env marker.
func (s *GritService) shouldMigrate(ctx context.Context, orgID uuid.UUID, apiApp *models.Application) bool {
	env, err := s.envSvc.GetEnvVarMap(ctx, apiApp.ID, models.ResourceTypeApp, orgID)
	if err != nil {
		return true // default on
	}
	if v, ok := env["ORBITA_GRIT_MIGRATE"]; ok {
		return v != "false"
	}
	return true
}

// runMigrations runs the Grit migrator for the API app under an advisory lock.
func (s *GritService) runMigrations(ctx context.Context, orgID uuid.UUID, orgSlug string, apiApp *models.Application) error {
	env, err := s.envSvc.GetEnvVarMap(ctx, apiApp.ID, models.ResourceTypeApp, orgID)
	if err != nil {
		return fmt.Errorf("runMigrations: env: %w", err)
	}
	dbURL := env["DATABASE_URL"]
	if dbURL == "" {
		return fmt.Errorf("runMigrations: no DATABASE_URL — a postgres addon is required to migrate")
	}

	var src orchestrator.SourceConfig
	_ = jsonUnmarshal(apiApp.SourceConfig, &src)
	repoURL := src.RepoURL
	if repoURL == "" && src.RepoFullName != "" {
		repoURL = fmt.Sprintf("https://github.com/%s.git", src.RepoFullName)
	}

	logs, err := s.orch.RunGritMigrations(ctx, orchestrator.GritMigrateSpec{
		OrgSlug:     orgSlug,
		RepoURL:     repoURL,
		Branch:      src.Branch,
		APIContext:  src.BuildContext,
		DatabaseURL: dbURL,
		Env:         env,
	})
	if err != nil {
		return err
	}
	log.Info().Str("app", apiApp.Name).Int("log_bytes", len(logs)).Msg("Grit migrations applied")
	return nil
}

func (s *GritService) serviceLink(ctx context.Context, orgID uuid.UUID, app *models.Application) GritServiceLink {
	status, _ := s.orch.GetApplicationStatus(ctx, app)
	link := GritServiceLink{Role: roleOf(app), AppID: app.ID.String(), Status: status}
	if domains, err := s.domainSvc.ListDomainsByResource(ctx, app.ID, models.ResourceTypeApp); err == nil && len(domains) > 0 {
		link.URL = "https://" + domains[0].Domain
	}
	return link
}

// fillLinks populates the live/api/dashboard URLs from the deployed services'
// domains.
func (s *GritService) fillLinks(ctx context.Context, orgID uuid.UUID, gritApp string, envID uuid.UUID, result *GritDeployResult) {
	services, _ := s.appRepo.ListGritServices(ctx, envID, orgID, gritApp)
	var apiDomain string
	for i := range services {
		role := roleOf(&services[i])
		domains, err := s.domainSvc.ListDomainsByResource(ctx, services[i].ID, models.ResourceTypeApp)
		if err != nil || len(domains) == 0 {
			continue
		}
		host := domains[0].Domain
		switch role {
		case grit.RoleWeb, grit.RoleApp:
			result.LiveURL = "https://" + host
		case grit.RoleAPI:
			apiDomain = host
			result.APIURL = "https://" + host
		}
	}
	if apiDomain != "" {
		result.DashboardURL = map[string]string{
			"pulse":    "https://" + apiDomain + "/pulse/ui",
			"sentinel": "https://" + apiDomain + "/sentinel/ui",
			"health":   "https://" + apiDomain + "/api/health",
		}
	}
}

func isAPIRole(role string) bool { return role == grit.RoleAPI || role == grit.RoleApp }

func roleOf(app *models.Application) string {
	if app.GritRole != nil {
		return *app.GritRole
	}
	return ""
}

// orderServices returns services with the API first, then a stable order.
func orderServices(in []models.Application) []models.Application {
	out := make([]models.Application, len(in))
	copy(out, in)
	sort.SliceStable(out, func(i, j int) bool {
		return rolePriority(roleOf(&out[i])) < rolePriority(roleOf(&out[j]))
	})
	return out
}

func rolePriority(role string) int {
	switch role {
	case grit.RoleAPI, grit.RoleApp:
		return 0
	case grit.RoleWeb:
		return 1
	case grit.RoleAdmin:
		return 2
	case grit.RoleDocs:
		return 3
	default:
		return 9
	}
}
