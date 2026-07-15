package grit

import "testing"

func TestParseAndDefaults(t *testing.T) {
	y := []byte(`
app: rental-manager
repo: MUKE-coder/rental-manager
addons:
  - postgres
  - redis
domains:
  web: hmkestates.com
  api: api.hmkestates.com
env:
  from: .env.production
`)
	m, err := ParseManifest(y)
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if m.App != "rental-manager" {
		t.Errorf("app = %q", m.App)
	}
	if m.BranchOrDefault() != "main" {
		t.Errorf("branch default = %q, want main", m.BranchOrDefault())
	}
	if !m.MigrateEnabled() {
		t.Error("migrate should default true")
	}
	if !m.ObservabilityEnabled() || !m.SecurityEnabled() {
		t.Error("observability/security should default true")
	}
	if m.StudioEnabled() {
		t.Error("studio should default false")
	}
	owner, name, ok := m.RepoOwnerName()
	if !ok || owner != "MUKE-coder" || name != "rental-manager" {
		t.Errorf("RepoOwnerName = %q/%q/%v", owner, name, ok)
	}
}

func TestToggleOverrides(t *testing.T) {
	y := []byte("app: x\nrepo: o/x\nmigrate: false\nobservability: false\nstudio: true\n")
	m, _ := ParseManifest(y)
	if m.MigrateEnabled() {
		t.Error("migrate should be false")
	}
	if m.ObservabilityEnabled() {
		t.Error("observability should be false")
	}
	if !m.StudioEnabled() {
		t.Error("studio should be true")
	}
	if !m.SecurityEnabled() {
		t.Error("security should still default true")
	}
}

func TestValidateManifest(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{"valid", "app: x\nrepo: owner/name\n", false},
		{"missing app", "repo: owner/name\n", true},
		{"missing repo", "app: x\n", true},
		{"bad repo form", "app: x\nrepo: notaslash\n", true},
		{"bad addon", "app: x\nrepo: o/n\naddons:\n  - mysql\n", true},
		{"domain with scheme", "app: x\nrepo: o/n\ndomains:\n  web: https://x.com\n", true},
		{"domain with port", "app: x\nrepo: o/n\ndomains:\n  api: x.com:8080\n", true},
		{"domain not fqdn", "app: x\nrepo: o/n\ndomains:\n  web: localhost\n", true},
		{"valid domains+addons", "app: x\nrepo: o/n\naddons:\n  - postgres\n  - minio\ndomains:\n  web: a.com\n  api: api.a.com\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := ParseManifest([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			err = ValidateManifest(m)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateManifest err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeriveServices(t *testing.T) {
	tests := []struct {
		mode      string
		docs      bool
		wantRoles []string
	}{
		{ModeSingle, false, []string{"app"}},
		{ModeAPI, false, []string{"api"}},
		{ModeDouble, false, []string{"api", "web"}},
		{ModeTriple, false, []string{"api", "web", "admin"}},
		{ModeTriple, true, []string{"api", "web", "admin", "docs"}},
	}
	for _, tt := range tests {
		g := &GritJSON{Architecture: tt.mode}
		g.Apps.Docs = tt.docs
		svcs, err := DeriveServices(g)
		if err != nil {
			t.Fatalf("%s: %v", tt.mode, err)
		}
		var roles []string
		for _, s := range svcs {
			roles = append(roles, s.Role)
		}
		if len(roles) != len(tt.wantRoles) {
			t.Fatalf("%s: roles = %v, want %v", tt.mode, roles, tt.wantRoles)
		}
		for i := range roles {
			if roles[i] != tt.wantRoles[i] {
				t.Errorf("%s: role[%d] = %q, want %q", tt.mode, i, roles[i], tt.wantRoles[i])
			}
		}
	}
}

func TestDeriveServicesBuildRecipe(t *testing.T) {
	// triple: api builds from apps/api with root Dockerfile; web/admin build
	// from repo root with their own Dockerfile and need NEXT_PUBLIC_API_URL.
	g := &GritJSON{Architecture: ModeTriple}
	svcs, _ := DeriveServices(g)
	byRole := map[string]Service{}
	for _, s := range svcs {
		byRole[s.Role] = s
	}

	api := byRole["api"]
	if api.BuildContext != "apps/api" || api.DockerfilePath != "Dockerfile" || api.Port != 8080 || !api.IsAPI {
		t.Errorf("api recipe wrong: %+v", api)
	}
	if api.NeedsAPIURL {
		t.Error("api should not need NEXT_PUBLIC_API_URL")
	}

	web := byRole["web"]
	if web.BuildContext != "" || web.DockerfilePath != "apps/web/Dockerfile" || web.Port != 3000 || !web.NeedsAPIURL {
		t.Errorf("web recipe wrong: %+v", web)
	}
	if web.IsAPI {
		t.Error("web should not be marked API")
	}
}

func TestDeriveServicesMobileRejected(t *testing.T) {
	if _, err := DeriveServices(&GritJSON{Architecture: ModeMobile}); err == nil {
		t.Error("mobile should be rejected")
	}
}

func TestParseGritJSON(t *testing.T) {
	// The real stoka marker (grit-knowledge README).
	g, err := ParseGritJSON([]byte(`{"architecture":"triple","frontend":"next","version":"3.59.0","apps":{"expo":true,"desktop":true,"docs":true}}`))
	if err != nil {
		t.Fatalf("ParseGritJSON: %v", err)
	}
	if g.Architecture != "triple" || g.Frontend != "next" || !g.Apps.Docs {
		t.Errorf("parsed wrong: %+v", g)
	}
	// expo/desktop must not affect the deployable set
	svcs, _ := DeriveServices(g)
	for _, s := range svcs {
		if s.Role == "expo" || s.Role == "desktop" {
			t.Errorf("expo/desktop must never be deployable: %+v", s)
		}
	}
	if len(svcs) != 4 { // api, web, admin, docs
		t.Errorf("stoka should derive 4 services, got %d", len(svcs))
	}
}

func TestParseGritJSONRejectsNonGrit(t *testing.T) {
	if _, err := ParseGritJSON([]byte(`{"name":"not-grit"}`)); err == nil {
		t.Error("json without architecture should be rejected")
	}
}

func TestValidateForDeploy(t *testing.T) {
	m, _ := ParseManifest([]byte("app: x\nrepo: o/n\nservices:\n  admin:\n    path: ./apps/admin\n"))
	// admin exists in triple
	if err := ValidateForDeploy(m, &GritJSON{Architecture: ModeTriple}); err != nil {
		t.Errorf("triple admin override should be valid: %v", err)
	}
	// admin does NOT exist in double
	if err := ValidateForDeploy(m, &GritJSON{Architecture: ModeDouble}); err == nil {
		t.Error("admin override in a double app should fail")
	}
	// mobile rejected
	if err := ValidateForDeploy(m, &GritJSON{Architecture: ModeMobile}); err == nil {
		t.Error("mobile should be rejected for deploy")
	}
}
