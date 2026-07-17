package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDotenv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	os.WriteFile(path, []byte(`
# a comment
JWT_SECRET=abc123
export API_URL="https://api.example.com"
QUOTED='single'
EMPTY=
SPACED = value with spaces
`), 0o644)

	env, err := parseDotenv(path)
	if err != nil {
		t.Fatal(err)
	}
	if env["JWT_SECRET"] != "abc123" {
		t.Errorf("JWT_SECRET = %q", env["JWT_SECRET"])
	}
	if env["API_URL"] != "https://api.example.com" {
		t.Errorf("API_URL = %q (export prefix + quotes should be stripped)", env["API_URL"])
	}
	if env["QUOTED"] != "single" {
		t.Errorf("QUOTED = %q", env["QUOTED"])
	}
	if _, ok := env["# a comment"]; ok {
		t.Error("comment line was parsed as a var")
	}
	if env["SPACED"] != "value with spaces" {
		t.Errorf("SPACED = %q", env["SPACED"])
	}
}

func TestLoadReadsManifestAndDetects(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "grit.json"),
		[]byte(`{"architecture":"api","version":"3.59.0","apps":{}}`), 0o644)
	os.WriteFile(filepath.Join(dir, "orbita.yaml"),
		[]byte("app: x\nrepo: o/x\naddons: [postgres]\ndomains:\n  api: api.x.com\nenv:\n  from: .env.production\n"), 0o644)
	os.WriteFile(filepath.Join(dir, ".env.production"), []byte("JWT_SECRET=s\n"), 0o644)

	p, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if p.Manifest.App != "x" || p.GritJSON.Architecture != "api" {
		t.Errorf("parsed wrong: %+v / %+v", p.Manifest, p.GritJSON)
	}
	if p.EnvValues["JWT_SECRET"] != "s" {
		t.Errorf("env.from not read: %v", p.EnvValues)
	}
}

func TestLoadNoManifest(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(dir); err != ErrNoManifest {
		t.Errorf("expected ErrNoManifest, got %v", err)
	}
}
