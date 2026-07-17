package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/grit"
	gritbuild "github.com/orbita-sh/orbita/internal/grit/build"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/orchestrator"
	"github.com/orbita-sh/orbita/internal/repository"
)

// GritService reconciles an orbita.yaml + grit.json into Orbita resources (apps,
// addons, env, domains) and drives the Grit-aware deploy (build → migrate →
// route → live). It composes the existing services rather than duplicating them.
type GritService struct {
	appRepo       *repository.AppRepository
	orgRepo       *repository.OrgRepository
	projectSvc    *ProjectService
	appSvc        *AppService
	dbSvc         *DBService
	envSvc        *EnvService
	domainSvc     *DomainService
	orch          *orchestrator.Orchestrator
	encryptionKey []byte
}

func NewGritService(
	appRepo *repository.AppRepository,
	orgRepo *repository.OrgRepository,
	projectSvc *ProjectService,
	appSvc *AppService,
	dbSvc *DBService,
	envSvc *EnvService,
	domainSvc *DomainService,
	orch *orchestrator.Orchestrator,
	encryptionKey []byte,
) *GritService {
	return &GritService{
		appRepo:       appRepo,
		orgRepo:       orgRepo,
		projectSvc:    projectSvc,
		appSvc:        appSvc,
		dbSvc:         dbSvc,
		envSvc:        envSvc,
		domainSvc:     domainSvc,
		orch:          orch,
		encryptionKey: encryptionKey,
	}
}

// ReconcileInput is what the CLI submits: the manifest, the detected grit.json,
// the local env-file values, and how to reach the repo.
type ReconcileInput struct {
	OrgID         uuid.UUID
	OrgSlug       string
	Manifest      *grit.Manifest
	GritJSON      *grit.GritJSON
	EnvValues     map[string]string // parsed from env.from (never logged)
	GitConnID     *uuid.UUID        // optional; nil for public repos
	EnvironmentID *uuid.UUID        // optional; defaults to the app's production env
}

// ServiceReconcileResult reports what happened to one service.
type ServiceReconcileResult struct {
	Role    string `json:"role"`
	AppName string `json:"app_name"`
	AppID   string `json:"app_id"`
	Domain  string `json:"domain,omitempty"`
	Created bool   `json:"created"`
}

// ReconcileResult summarizes a reconcile (also serves the --plan dry run).
type ReconcileResult struct {
	GritApp       string                   `json:"grit_app"`
	Mode          string                   `json:"mode"`
	ProjectID     string                   `json:"project_id"`
	EnvironmentID string                   `json:"environment_id"`
	Services      []ServiceReconcileResult `json:"services"`
	Addons        []string                 `json:"addons"`
	Migrate       bool                     `json:"migrate"`
	DashboardURLs map[string]string        `json:"dashboard_urls,omitempty"`
}

// Reconcile creates-or-updates all Orbita resources for a Grit app. Idempotent:
// re-running reconciles to the same state (no duplicate apps/addons). It does not
// deploy — call Deploy after.
func (s *GritService) Reconcile(ctx context.Context, in ReconcileInput) (*ReconcileResult, error) {
	m := in.Manifest
	if err := grit.ValidateForDeploy(m, in.GritJSON); err != nil {
		return nil, err
	}

	plan, err := gritbuild.BuildPlan(m, in.GritJSON)
	if err != nil {
		return nil, err
	}

	// 1. Ensure project + production environment for this Grit app.
	project, err := s.projectSvc.EnsureProject(ctx, in.OrgID, m.App)
	if err != nil {
		return nil, fmt.Errorf("Reconcile: project: %w", err)
	}
	var envID uuid.UUID
	if in.EnvironmentID != nil {
		envID = *in.EnvironmentID
	} else {
		envID, err = s.projectSvc.ProductionEnvID(ctx, project, in.OrgID)
		if err != nil {
			return nil, fmt.Errorf("Reconcile: environment: %w", err)
		}
	}

	// 2. Provision addons (idempotent) and collect the env they contribute.
	addonEnv, err := s.provisionAddons(ctx, in.OrgID, in.OrgSlug, envID, m)
	if err != nil {
		return nil, fmt.Errorf("Reconcile: addons: %w", err)
	}

	// 3. Compute the full API env = user env-file + addon env + dashboards +
	//    platform overrides.
	apiEnv := ComputeGritAPIEnv(m, in.EnvValues, addonEnv)

	// 4. Create-or-update one application per service.
	var results []ServiceReconcileResult
	for _, sp := range plan.Services {
		app, created, err := s.upsertServiceApp(ctx, in, envID, m, sp)
		if err != nil {
			return nil, fmt.Errorf("Reconcile: service %s: %w", sp.Role, err)
		}

		// The API service carries the full runtime env (incl. addon URLs and
		// dashboard creds). Front-ends get their public API URL at build time.
		if sp.IsAPI {
			if err := s.injectEnv(ctx, in.OrgID, app.ID, apiEnv); err != nil {
				return nil, fmt.Errorf("Reconcile: inject env: %w", err)
			}
		}

		// Attach the service's domain (idempotent).
		if sp.Domain != "" {
			if err := s.ensureDomain(ctx, in.OrgID, app.ID, sp.Domain, sp.Port); err != nil {
				log.Warn().Err(err).Str("domain", sp.Domain).Msg("Grit reconcile: domain attach failed")
			}
		}

		results = append(results, ServiceReconcileResult{
			Role: sp.Role, AppName: sp.AppName, AppID: app.ID.String(), Domain: sp.Domain, Created: created,
		})
	}

	res := &ReconcileResult{
		GritApp:       m.App,
		Mode:          plan.Mode,
		ProjectID:     project.ID.String(),
		EnvironmentID: envID.String(),
		Services:      results,
		Addons:        m.Addons,
		Migrate:       plan.Migrate,
		DashboardURLs: DashboardURLs(m),
	}
	return res, nil
}

// upsertServiceApp finds the existing app for (env, gritApp, role) or creates it,
// keeping its source config in sync with the plan.
func (s *GritService) upsertServiceApp(ctx context.Context, in ReconcileInput, envID uuid.UUID, m *grit.Manifest, sp gritbuild.ServicePlan) (*models.Application, bool, error) {
	existing, err := s.appRepo.ListGritServices(ctx, envID, in.OrgID, m.App)
	if err != nil {
		return nil, false, err
	}
	for i := range existing {
		if existing[i].GritRole != nil && *existing[i].GritRole == sp.Role {
			app := &existing[i]
			// Keep build recipe + auto-deploy in sync on re-reconcile.
			app.SourceConfig = s.gritSourceConfig(m, sp, in.GitConnID)
			port := sp.Port
			app.Port = &port
			if err := s.appRepo.Update(ctx, app); err != nil {
				return nil, false, err
			}
			return app, false, nil
		}
	}

	port := sp.Port
	input := CreateAppInput{
		Name:            sp.AppName,
		EnvironmentID:   envID,
		SourceType:      models.SourceTypeGrit,
		Port:            &port,
		Replicas:        1,
		GitConnectionID: in.GitConnID,
		RepoFullName:    m.Repo,
		RepoURL:         m.RepoURL, // optional override; else derived from Repo
		Branch:          m.BranchOrDefault(),
		DockerfilePath:  sp.DockerfilePath,
		BuildContext:    sp.BuildContext,
		BuildArgs:       sp.BuildArgs,
		GritApp:         m.App,
		GritRole:        sp.Role,
	}
	app, err := s.appSvc.CreateApp(ctx, in.OrgID, input)
	if err != nil {
		return nil, false, err
	}
	return app, true, nil
}

// gritSourceConfig rebuilds an app's source_config JSON from the plan (used on
// re-reconcile to pick up recipe changes).
//
// gitConnID must be threaded through here, not just set at create time: the
// re-reconcile path overwrites SourceConfig wholesale, so leaving the connection
// out silently dropped it on the second deploy. A private repo then built once
// (create) and failed on every subsequent deploy with "could not read Username
// for 'https://github.com'".
func (s *GritService) gritSourceConfig(m *grit.Manifest, sp gritbuild.ServicePlan, gitConnID *uuid.UUID) []byte {
	repoURL := m.RepoURL
	if repoURL == "" {
		if owner, name, ok := m.RepoOwnerName(); ok {
			repoURL = fmt.Sprintf("https://github.com/%s/%s.git", owner, name)
		}
	}
	cfg := map[string]interface{}{
		"repo_full_name":  m.Repo,
		"repo_url":        repoURL,
		"branch":          m.BranchOrDefault(),
		"dockerfile_path": sp.DockerfilePath,
		"build_context":   sp.BuildContext,
		"grit_role":       sp.Role,
	}
	if gitConnID != nil {
		cfg["git_connection_id"] = gitConnID.String()
	}
	if len(sp.BuildArgs) > 0 {
		cfg["build_args"] = sp.BuildArgs
	}
	return mustJSON(cfg)
}

// injectEnv upserts every key of the computed env, marking sensitive keys as
// secret so they're masked in the API.
func (s *GritService) injectEnv(ctx context.Context, orgID, appID uuid.UUID, env map[string]string) error {
	for k, v := range env {
		if err := s.envSvc.SetEnvVar(ctx, appID, models.ResourceTypeApp, k, v, isSecretKey(k), orgID); err != nil {
			return err
		}
	}
	return nil
}

// ensureDomain attaches a domain to an app if not already present.
func (s *GritService) ensureDomain(ctx context.Context, orgID, appID uuid.UUID, domain string, port int) error {
	existing, err := s.domainSvc.ListDomainsByResource(ctx, appID, models.ResourceTypeApp)
	if err == nil {
		for _, d := range existing {
			if d.Domain == domain {
				return nil // already attached
			}
		}
	}
	_, err = s.domainSvc.AddDomain(ctx, appID, models.ResourceTypeApp, domain, orgID, true, port)
	return err
}

// mustJSON marshals or panics (inputs are always marshalable maps).
func mustJSON(v interface{}) []byte {
	b, err := jsonMarshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// randomPassword returns a URL-safe strong password.
func randomPassword(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// isSecretKey heuristically flags env keys whose values should be masked.
func isSecretKey(key string) bool {
	k := strings.ToUpper(key)
	for _, tok := range []string{"SECRET", "PASSWORD", "TOKEN", "KEY", "DATABASE_URL", "REDIS_URL", "DSN"} {
		if strings.Contains(k, tok) {
			return true
		}
	}
	return false
}
