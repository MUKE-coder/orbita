package grit

import (
	"encoding/json"
	"fmt"
)

// Architecture modes (grit.json.architecture). Determines the deployable shape.
const (
	ModeSingle = "single"
	ModeDouble = "double"
	ModeTriple = "triple"
	ModeAPI    = "api"
	ModeMobile = "mobile"
)

// Grit service roles.
const (
	RoleApp   = "app"   // single mode: one container serving SPA + API on :8080
	RoleAPI   = "api"   // Go API on :8080
	RoleWeb   = "web"   // Next.js/Vite public site on :3000
	RoleAdmin = "admin" // Next.js admin panel on :3000
	RoleDocs  = "docs"  // Next.js docs on :3002
)

// GritJSON is the machine marker at every Grit repo root. Its presence is how a
// repo is detected as a Grit app; architecture selects the deploy strategy.
type GritJSON struct {
	Architecture string `json:"architecture"`
	Frontend     string `json:"frontend"` // next | vite (irrelevant for api mode)
	Version      string `json:"version"`
	Apps         struct {
		Expo    bool `json:"expo"`
		Desktop bool `json:"desktop"`
		Docs    bool `json:"docs"`
	} `json:"apps"`
}

// ParseGritJSON parses grit.json bytes.
func ParseGritJSON(data []byte) (*GritJSON, error) {
	var g GritJSON
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, fmt.Errorf("ParseGritJSON: %w", err)
	}
	if g.Architecture == "" {
		return nil, fmt.Errorf("ParseGritJSON: grit.json has no architecture field — not a valid Grit marker")
	}
	return &g, nil
}

// Service is one deployable Grit service, with the exact build recipe Orbita
// must use (grit-knowledge/04). BuildContext is the Docker build context
// relative to the repo root ("" = repo root); DockerfilePath is relative to the
// context.
type Service struct {
	Role           string   // app | api | web | admin | docs
	DockerfilePath string   // e.g. "Dockerfile" or "apps/web/Dockerfile"
	BuildContext   string   // "" = repo root, "apps/api" = subdir
	Port           int      // container listen port
	NeedsAPIURL    bool     // Next.js apps take NEXT_PUBLIC_API_URL as a build arg
	IsAPI          bool     // true for the Go API (app or api role) — migrations run here
}

// IsVPSDeployable reports whether a mode can be deployed to a VPS.
func IsVPSDeployable(mode string) bool {
	switch mode {
	case ModeSingle, ModeDouble, ModeTriple, ModeAPI:
		return true
	default:
		return false // mobile ships to app stores
	}
}

// DeriveServices returns the deployable service list for a Grit app, derived
// purely from grit.json (mode + docs flag). expo/desktop are never returned —
// they are clients of the API, not server workloads. See grit-knowledge/02, 04.
func DeriveServices(g *GritJSON) ([]Service, error) {
	switch g.Architecture {
	case ModeSingle:
		// One container: Go binary with the SPA embedded. Dockerfile + module
		// at the repo root. Serves SPA at / and API at /api/*.
		return []Service{
			{Role: RoleApp, DockerfilePath: "Dockerfile", BuildContext: "", Port: 8080, IsAPI: true},
		}, nil

	case ModeAPI:
		return []Service{
			{Role: RoleAPI, DockerfilePath: "Dockerfile", BuildContext: "apps/api", Port: 8080, IsAPI: true},
		}, nil

	case ModeDouble:
		return []Service{
			{Role: RoleAPI, DockerfilePath: "Dockerfile", BuildContext: "apps/api", Port: 8080, IsAPI: true},
			{Role: RoleWeb, DockerfilePath: "apps/web/Dockerfile", BuildContext: "", Port: 3000, NeedsAPIURL: true},
		}, nil

	case ModeTriple:
		svcs := []Service{
			{Role: RoleAPI, DockerfilePath: "Dockerfile", BuildContext: "apps/api", Port: 8080, IsAPI: true},
			{Role: RoleWeb, DockerfilePath: "apps/web/Dockerfile", BuildContext: "", Port: 3000, NeedsAPIURL: true},
			{Role: RoleAdmin, DockerfilePath: "apps/admin/Dockerfile", BuildContext: "", Port: 3000, NeedsAPIURL: true},
		}
		if g.Apps.Docs {
			svcs = append(svcs, Service{Role: RoleDocs, DockerfilePath: "apps/docs/Dockerfile", BuildContext: "", Port: 3002})
		}
		return svcs, nil

	case ModeMobile:
		return nil, fmt.Errorf("DeriveServices: mobile apps ship to app stores and cannot be deployed to a VPS")

	default:
		return nil, fmt.Errorf("DeriveServices: unknown architecture %q", g.Architecture)
	}
}

// APIService returns the Go API service (the one that runs migrations and hosts
// Pulse/Sentinel/Studio), or false if none.
func APIService(services []Service) (Service, bool) {
	for _, s := range services {
		if s.IsAPI {
			return s, true
		}
	}
	return Service{}, false
}

// DefaultPortForRole returns the conventional port for a service role.
func DefaultPortForRole(role string) int {
	switch role {
	case RoleDocs:
		return 3002
	case RoleWeb, RoleAdmin:
		return 3000
	default: // app, api
		return 8080
	}
}
