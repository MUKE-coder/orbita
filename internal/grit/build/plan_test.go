package build

import (
	"strings"
	"testing"

	"github.com/orbita-sh/orbita/internal/grit"
)

func triple(docs bool) *grit.GritJSON {
	g := &grit.GritJSON{Architecture: grit.ModeTriple, Frontend: "next"}
	g.Apps.Docs = docs
	return g
}

func TestBuildPlanTriple(t *testing.T) {
	m, _ := grit.ParseManifest([]byte(`
app: stoka
repo: MUKE-coder/stoka
domains:
  web: stoka.com
  admin: admin.stoka.com
  api: api.stoka.com
`))
	p, err := BuildPlan(m, triple(true))
	if err != nil {
		t.Fatal(err)
	}
	if p.GritApp != "stoka" || p.Mode != grit.ModeTriple || !p.Migrate {
		t.Errorf("plan meta wrong: %+v", *p)
	}
	if len(p.Services) != 4 {
		t.Fatalf("want 4 services, got %d", len(p.Services))
	}

	byRole := map[string]ServicePlan{}
	for _, s := range p.Services {
		byRole[s.Role] = s
	}

	api := byRole["api"]
	if api.AppName != "stoka-api" || api.BuildContext != "apps/api" || api.DockerfilePath != "Dockerfile" ||
		api.Port != 8080 || api.Domain != "api.stoka.com" || !api.IsAPI {
		t.Errorf("api plan wrong: %+v", api)
	}
	if api.BuildArgs != nil {
		t.Error("api should have no build args")
	}

	web := byRole["web"]
	if web.BuildContext != "" || web.DockerfilePath != "apps/web/Dockerfile" || web.Port != 3000 ||
		web.Domain != "stoka.com" {
		t.Errorf("web plan wrong: %+v", web)
	}
	if web.BuildArgs["NEXT_PUBLIC_API_URL"] != "https://api.stoka.com" {
		t.Errorf("web NEXT_PUBLIC_API_URL wrong: %q", web.BuildArgs["NEXT_PUBLIC_API_URL"])
	}

	admin := byRole["admin"]
	if admin.Domain != "admin.stoka.com" || admin.BuildArgs["NEXT_PUBLIC_API_URL"] != "https://api.stoka.com" {
		t.Errorf("admin plan wrong: %+v", admin)
	}

	docs := byRole["docs"]
	if docs.Port != 3002 || docs.DockerfilePath != "apps/docs/Dockerfile" {
		t.Errorf("docs plan wrong: %+v", docs)
	}

	if api2, ok := p.APIService(); !ok || api2.Role != "api" {
		t.Error("APIService should return the api service")
	}
}

func TestBuildPlanAPIURLFallback(t *testing.T) {
	// No domains.api but domains.web present → derive api.<web>
	m, _ := grit.ParseManifest([]byte("app: x\nrepo: o/x\ndomains:\n  web: www.example.com\n"))
	p, _ := BuildPlan(m, triple(false))
	for _, s := range p.Services {
		if s.Role == "web" {
			if s.BuildArgs["NEXT_PUBLIC_API_URL"] != "https://api.example.com" {
				t.Errorf("fallback api url wrong: %q", s.BuildArgs["NEXT_PUBLIC_API_URL"])
			}
		}
	}
}

func TestBuildPlanSingle(t *testing.T) {
	m, _ := grit.ParseManifest([]byte("app: mini\nrepo: o/mini\ndomains:\n  web: mini.com\n"))
	p, _ := BuildPlan(m, &grit.GritJSON{Architecture: grit.ModeSingle})
	if len(p.Services) != 1 {
		t.Fatalf("single should have 1 service, got %d", len(p.Services))
	}
	app := p.Services[0]
	if app.Role != "app" || app.Port != 8080 || app.DockerfilePath != "Dockerfile" ||
		app.BuildContext != "" || app.Domain != "mini.com" || !app.IsAPI {
		t.Errorf("single app plan wrong: %+v", app)
	}
}

func TestBuildPlanMigrateToggle(t *testing.T) {
	m, _ := grit.ParseManifest([]byte("app: x\nrepo: o/x\nmigrate: false\n"))
	p, _ := BuildPlan(m, triple(false))
	if p.Migrate {
		t.Error("migrate should be false")
	}
}

func TestGenerateDockerfileShapes(t *testing.T) {
	api, ok := GenerateDockerfile(grit.RoleAPI)
	if !ok || !strings.Contains(api, "CGO_ENABLED=0") || !strings.Contains(api, "chown -R app:app /app") || !strings.Contains(api, "EXPOSE 8080") {
		t.Errorf("api dockerfile missing key properties")
	}
	web, ok := GenerateDockerfile(grit.RoleWeb)
	if !ok || !strings.Contains(web, "pnpm@9.15.0") || !strings.Contains(web, "NEXT_PUBLIC_API_URL") || !strings.Contains(web, ".next/standalone") || !strings.Contains(web, "EXPOSE 3000") {
		t.Errorf("web dockerfile missing key properties")
	}
	single, ok := GenerateDockerfile(grit.RoleApp)
	if !ok || !strings.Contains(single, "frontend/dist") || !strings.Contains(single, "CGO_ENABLED=0") {
		t.Errorf("single dockerfile missing key properties")
	}
	if _, ok := GenerateDockerfile("bogus"); ok {
		t.Error("unknown role should return ok=false")
	}
}
