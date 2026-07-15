package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	gritbuild "github.com/orbita-sh/orbita/internal/grit/build"
	"github.com/orbita-sh/orbita/internal/models"
)

// PlanReconcile validates + derives the full plan WITHOUT mutating anything —
// backs `grit deploy --plan`. It reports exactly what a reconcile would create
// or change.
func (s *GritService) PlanReconcile(ctx context.Context, in ReconcileInput) (*ReconcileResult, error) {
	m := in.Manifest
	plan, err := gritbuild.BuildPlan(m, in.GritJSON)
	if err != nil {
		return nil, err
	}

	// Determine which services already exist (so the plan can show created vs
	// updated) without creating anything.
	envID := uuid.Nil
	projectID := ""
	existingRoles := map[string]models.Application{}
	if project, err := s.projectSvc.EnsureProjectReadOnly(ctx, in.OrgID, m.App); err == nil && project != nil {
		projectID = project.ID.String()
		if id, err := s.projectSvc.ProductionEnvID(ctx, project, in.OrgID); err == nil {
			envID = id
			if apps, err := s.appRepo.ListGritServices(ctx, envID, in.OrgID, m.App); err == nil {
				for _, a := range apps {
					if a.GritRole != nil {
						existingRoles[*a.GritRole] = a
					}
				}
			}
		}
	}

	var services []ServiceReconcileResult
	for _, sp := range plan.Services {
		r := ServiceReconcileResult{Role: sp.Role, AppName: sp.AppName, Domain: sp.Domain, Created: true}
		if a, ok := existingRoles[sp.Role]; ok {
			r.Created = false
			r.AppID = a.ID.String()
		}
		services = append(services, r)
	}

	return &ReconcileResult{
		GritApp:       m.App,
		Mode:          plan.Mode,
		ProjectID:     projectID,
		EnvironmentID: envIDString(envID),
		Services:      services,
		Addons:        m.Addons,
		Migrate:       plan.Migrate,
		DashboardURLs: DashboardURLs(m),
	}, nil
}

func envIDString(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

// ResolveEnvID resolves the environment for a Grit app: an explicit id, or the
// app's project production env.
func (s *GritService) ResolveEnvID(ctx context.Context, orgID uuid.UUID, gritApp, explicitEnvID string) (uuid.UUID, error) {
	if explicitEnvID != "" {
		id, err := uuid.Parse(explicitEnvID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid environment_id")
		}
		return id, nil
	}
	project, err := s.projectSvc.EnsureProjectReadOnly(ctx, orgID, gritApp)
	if err != nil || project == nil {
		return uuid.Nil, fmt.Errorf("grit app %q not found — reconcile first", gritApp)
	}
	return s.projectSvc.ProductionEnvID(ctx, project, orgID)
}

// Status returns the per-service status + links for a Grit app.
func (s *GritService) Status(ctx context.Context, orgID uuid.UUID, gritApp, explicitEnvID string) (*GritDeployResult, error) {
	envID, err := s.ResolveEnvID(ctx, orgID, gritApp, explicitEnvID)
	if err != nil {
		return nil, err
	}
	services, err := s.appRepo.ListGritServices(ctx, envID, orgID, gritApp)
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("grit app %q has no services", gritApp)
	}

	result := &GritDeployResult{GritApp: gritApp}
	for i := range services {
		result.Services = append(result.Services, s.serviceLink(ctx, orgID, &services[i]))
	}
	s.fillLinks(ctx, orgID, gritApp, envID, result)
	return result, nil
}

// Rollback reverts every service of a Grit app to its previous deploy. Order
// doesn't matter for rollback (no migration — Grit migrations are additive, so
// older code tolerates the newer schema; grit-knowledge/06).
func (s *GritService) Rollback(ctx context.Context, orgID uuid.UUID, orgSlug, gritApp, explicitEnvID string, userID *uuid.UUID) (*GritDeployResult, error) {
	envID, err := s.ResolveEnvID(ctx, orgID, gritApp, explicitEnvID)
	if err != nil {
		return nil, err
	}
	services, err := s.appRepo.ListGritServices(ctx, envID, orgID, gritApp)
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("grit app %q has no services", gritApp)
	}

	result := &GritDeployResult{GritApp: gritApp}
	for i := range services {
		app := &services[i]
		prev, err := s.previousSuccessfulDeploy(ctx, app.ID)
		if err != nil {
			result.Services = append(result.Services, GritServiceLink{Role: roleOf(app), AppID: app.ID.String(), Status: "no-previous-deploy"})
			continue
		}
		if _, err := s.appSvc.Rollback(ctx, app.ID, prev, orgID, orgSlug, userID); err != nil {
			result.Services = append(result.Services, GritServiceLink{Role: roleOf(app), AppID: app.ID.String(), Status: "rollback-failed"})
			continue
		}
		result.Services = append(result.Services, s.serviceLink(ctx, orgID, app))
	}
	s.fillLinks(ctx, orgID, gritApp, envID, result)
	return result, nil
}

// previousSuccessfulDeploy returns the deployment id to roll back to: the most
// recent successful deploy before the current one.
func (s *GritService) previousSuccessfulDeploy(ctx context.Context, appID uuid.UUID) (uuid.UUID, error) {
	deps, err := s.appRepo.ListDeployments(ctx, appID, 20)
	if err != nil {
		return uuid.Nil, err
	}
	successful := 0
	for _, d := range deps {
		if d.Status == models.DeployStatusSuccess {
			successful++
			if successful == 2 { // the one before the current live version
				return d.ID, nil
			}
		}
	}
	return uuid.Nil, fmt.Errorf("no previous successful deploy")
}
