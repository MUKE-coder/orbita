package hosts

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestHostsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts.yaml")
	t.Setenv("ORBITA_HOSTS_FILE", path)

	// Empty registry when file is absent.
	f, err := Load()
	if err != nil {
		t.Fatalf("Load empty: %v", err)
	}
	if len(f.Hosts) != 0 {
		t.Fatalf("expected empty registry, got %d", len(f.Hosts))
	}

	// Set two hosts.
	if err := Set("prod", Host{APIURL: "https://orbita.example.com", Token: "orb_abc", SSH: "root@1.2.3.4"}); err != nil {
		t.Fatal(err)
	}
	if err := Set("staging", Host{APIURL: "https://staging.example.com", Token: "orb_def"}); err != nil {
		t.Fatal(err)
	}

	// Resolve.
	h, err := Resolve("prod")
	if err != nil {
		t.Fatal(err)
	}
	if h.APIURL != "https://orbita.example.com" || h.Token != "orb_abc" || h.SSH != "root@1.2.3.4" {
		t.Errorf("resolved wrong host: %+v", h)
	}

	// Names sorted.
	f, _ = Load()
	names := f.Names()
	if len(names) != 2 || names[0] != "prod" || names[1] != "staging" {
		t.Errorf("names = %v", names)
	}

	// File perms are 0600 (holds tokens). Windows doesn't enforce Unix perms,
	// so this only holds on the Linux/macOS operator machines that matter.
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("hosts file perms = %o, want 600", perm)
		}
	}
}

func TestResolveUnknownHost(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ORBITA_HOSTS_FILE", filepath.Join(dir, "hosts.yaml"))

	// No hosts registered.
	if _, err := Resolve("prod"); err == nil {
		t.Error("expected error resolving with empty registry")
	}

	_ = Set("prod", Host{APIURL: "https://x", Token: "orb_x"})
	if _, err := Resolve("nope"); err == nil {
		t.Error("expected error resolving unknown host")
	}
}
