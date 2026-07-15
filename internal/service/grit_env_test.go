package service

import (
	"strings"
	"testing"

	"github.com/orbita-sh/orbita/internal/grit"
)

func manifest(t *testing.T, y string) *grit.Manifest {
	t.Helper()
	m, err := grit.ParseManifest([]byte(y))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return m
}

func TestComputeGritAPIEnv_Precedence(t *testing.T) {
	m := manifest(t, `
app: stoka
repo: o/stoka
domains:
  web: stoka.com
  admin: admin.stoka.com
  api: api.stoka.com
`)
	userEnv := map[string]string{
		"JWT_SECRET":    "user-secret",
		"APP_ENV":       "development", // must be overridden to production
		"MAIL_FROM":     "hi@stoka.com",
		"POSTGRES_HOST": "localhost", // user's local value, addon should win via DATABASE_URL
	}
	addonEnv := map[string]string{
		"DATABASE_URL": "postgres://u:p@orbita-db-abc:5432/stoka-postgres?sslmode=disable",
		"REDIS_URL":    "redis://orbita-db-def:6379",
	}

	env := ComputeGritAPIEnv(m, userEnv, addonEnv)

	if env["APP_ENV"] != "production" {
		t.Errorf("APP_ENV = %q, want production (override must win)", env["APP_ENV"])
	}
	if env["JWT_SECRET"] != "user-secret" {
		t.Errorf("user JWT_SECRET should survive, got %q", env["JWT_SECRET"])
	}
	if !strings.HasPrefix(env["DATABASE_URL"], "postgres://") {
		t.Errorf("addon DATABASE_URL missing: %q", env["DATABASE_URL"])
	}
	if env["REDIS_URL"] != "redis://orbita-db-def:6379" {
		t.Errorf("addon REDIS_URL missing: %q", env["REDIS_URL"])
	}
	if env["API_URL"] != "https://api.stoka.com" {
		t.Errorf("API_URL = %q", env["API_URL"])
	}
	if env["CORS_ORIGINS"] != "https://stoka.com,https://admin.stoka.com" {
		t.Errorf("CORS_ORIGINS = %q", env["CORS_ORIGINS"])
	}
	if env["ORBITA_GRIT_MIGRATE"] != "true" {
		t.Errorf("migrate marker = %q", env["ORBITA_GRIT_MIGRATE"])
	}
}

func TestDashboardEnv_DefaultsAndGeneration(t *testing.T) {
	// Defaults: pulse+sentinel on, studio off; passwords generated when absent.
	m := manifest(t, "app: x\nrepo: o/x\n")
	env := DashboardEnv(m, map[string]string{})

	if env["PULSE_ENABLED"] != "true" || env["SENTINEL_ENABLED"] != "true" {
		t.Error("pulse/sentinel should default enabled")
	}
	if env["GORM_STUDIO_ENABLED"] != "false" {
		t.Error("studio should default disabled")
	}
	if len(env["PULSE_PASSWORD"]) < 16 {
		t.Errorf("pulse password should be strong+generated, got %q", env["PULSE_PASSWORD"])
	}
	if env["PULSE_USERNAME"] != "admin" {
		t.Errorf("pulse username default = %q", env["PULSE_USERNAME"])
	}
}

func TestDashboardEnv_RespectsUserCreds(t *testing.T) {
	m := manifest(t, "app: x\nrepo: o/x\nstudio: true\n")
	user := map[string]string{"PULSE_PASSWORD": "my-own-pass", "PULSE_USERNAME": "ops"}
	env := DashboardEnv(m, user)

	// Should NOT override user-supplied creds.
	if _, set := env["PULSE_PASSWORD"]; set {
		t.Error("should not generate a password when the user supplied one")
	}
	if _, set := env["PULSE_USERNAME"]; set {
		t.Error("should not set username when the user supplied one")
	}
	if env["GORM_STUDIO_ENABLED"] != "true" {
		t.Error("studio should be enabled when toggled on")
	}
	if len(env["GORM_STUDIO_PASSWORD"]) < 16 {
		t.Error("studio password should be generated when enabled + absent")
	}
}

func TestDashboardEnv_Disabled(t *testing.T) {
	m := manifest(t, "app: x\nrepo: o/x\nobservability: false\nsecurity: false\n")
	env := DashboardEnv(m, map[string]string{})
	if env["PULSE_ENABLED"] != "false" || env["SENTINEL_ENABLED"] != "false" {
		t.Error("disabled toggles should set *_ENABLED=false")
	}
	if _, set := env["PULSE_PASSWORD"]; set {
		t.Error("disabled dashboard should not generate a password")
	}
}

func TestDashboardURLs(t *testing.T) {
	m := manifest(t, "app: x\nrepo: o/x\nstudio: true\ndomains:\n  api: api.x.com\n")
	urls := DashboardURLs(m)
	if urls["pulse"] != "https://api.x.com/pulse/ui" {
		t.Errorf("pulse url = %q", urls["pulse"])
	}
	if urls["sentinel"] != "https://api.x.com/sentinel/ui" {
		t.Errorf("sentinel url = %q", urls["sentinel"])
	}
	if urls["studio"] != "https://api.x.com/studio" {
		t.Errorf("studio url = %q", urls["studio"])
	}
	if urls["health"] != "https://api.x.com/api/health" {
		t.Errorf("health url = %q", urls["health"])
	}

	// No api domain → no urls.
	m2 := manifest(t, "app: x\nrepo: o/x\n")
	if DashboardURLs(m2) != nil {
		t.Error("no api domain should yield nil urls")
	}
}

func TestIsSecretKey(t *testing.T) {
	secret := []string{"JWT_SECRET", "POSTGRES_PASSWORD", "DATABASE_URL", "REDIS_URL", "MINIO_SECRET_KEY", "GITHUB_CLIENT_SECRET", "RESEND_API_KEY"}
	for _, k := range secret {
		if !isSecretKey(k) {
			t.Errorf("%q should be secret", k)
		}
	}
	plain := []string{"APP_ENV", "PORT", "MAIL_FROM", "CORS_ORIGINS", "PULSE_ENABLED"}
	for _, k := range plain {
		if isSecretKey(k) {
			t.Errorf("%q should not be secret", k)
		}
	}
}
