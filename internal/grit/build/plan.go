// Package build turns a Grit manifest + detected grit.json into a concrete
// per-service build/deploy plan for Orbita's engine. It does NOT generate
// Dockerfiles: per grit-knowledge/04 the correct build recipe for a Grit app is
// the set of Dockerfiles Grit already committed to the repo, built with the
// exact contexts below. This package computes those inputs (Dockerfile path,
// build context, port, build args, target domain) for each service.
package build

import (
	"fmt"
	"strings"

	"github.com/orbita-sh/orbita/internal/grit"
)

// ServicePlan is everything Orbita needs to create-or-update one application
// for one Grit service.
type ServicePlan struct {
	Role           string            // app | api | web | admin | docs
	AppName        string            // e.g. "stoka-api"
	DockerfilePath string            // relative to the build context
	BuildContext   string            // "" = repo root, "apps/api" = subdir
	Port           int               // container listen port
	BuildArgs      map[string]string // e.g. NEXT_PUBLIC_API_URL (Next.js apps)
	Domain         string            // public hostname from grit.yaml (may be empty)
	IsAPI          bool              // migrations run against this service; hosts dashboards
}

// Plan is the full build/deploy plan for a Grit app.
type Plan struct {
	GritApp  string        // logical Grit app name (grouping key)
	Mode     string        // architecture mode
	Migrate  bool          // run the migration hook on deploy
	Services []ServicePlan // one per deployable service
}

// APIService returns the plan's Go API service (runs migrations, hosts
// Pulse/Sentinel/Studio), or false if none.
func (p *Plan) APIService() (ServicePlan, bool) {
	for _, s := range p.Services {
		if s.IsAPI {
			return s, true
		}
	}
	return ServicePlan{}, false
}

// BuildPlan derives the full plan from a validated manifest + grit.json.
func BuildPlan(m *grit.Manifest, g *grit.GritJSON) (*Plan, error) {
	services, err := grit.DeriveServices(g)
	if err != nil {
		return nil, fmt.Errorf("BuildPlan: %w", err)
	}

	// The public API URL is baked into the Next.js bundles at build time
	// (grit-knowledge/04). Derive it from domains.api; fall back to the web
	// domain's api. subdomain only if the user gave a web domain.
	apiURL := publicAPIURL(m)

	plan := &Plan{
		GritApp:  m.App,
		Mode:     g.Architecture,
		Migrate:  m.MigrateEnabled(),
		Services: make([]ServicePlan, 0, len(services)),
	}

	for _, s := range services {
		sp := ServicePlan{
			Role:           s.Role,
			AppName:        fmt.Sprintf("%s-%s", m.App, s.Role),
			DockerfilePath: s.DockerfilePath,
			BuildContext:   s.BuildContext,
			Port:           s.Port,
			Domain:         domainForRole(m, s.Role),
			IsAPI:          s.IsAPI,
		}
		// Honor an explicit service override's port/path if provided.
		if ov, ok := m.Services[s.Role]; ok {
			if ov.Port > 0 {
				sp.Port = ov.Port
			}
		}
		if s.NeedsAPIURL && apiURL != "" {
			sp.BuildArgs = map[string]string{"NEXT_PUBLIC_API_URL": apiURL}
		}
		plan.Services = append(plan.Services, sp)
	}
	return plan, nil
}

// domainForRole maps a Grit service role to its grit.yaml domain. In single
// mode the one container also answers on the web domain.
func domainForRole(m *grit.Manifest, role string) string {
	switch role {
	case grit.RoleApp, grit.RoleWeb:
		return m.Domains.Web
	case grit.RoleAdmin:
		return m.Domains.Admin
	case grit.RoleAPI:
		return m.Domains.API
	case grit.RoleDocs:
		return m.Domains.Docs
	default:
		return ""
	}
}

// publicAPIURL returns the https URL the front-ends should call. Prefers
// domains.api; if absent but domains.web is set, uses api.<web-domain>.
func publicAPIURL(m *grit.Manifest) string {
	if m.Domains.API != "" {
		return "https://" + m.Domains.API
	}
	if m.Domains.Web != "" {
		return "https://api." + strings.TrimPrefix(m.Domains.Web, "www.")
	}
	return ""
}
