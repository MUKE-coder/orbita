package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateInit(t *testing.T) {
	// Valid: server + admin-email, IP mode (no domain).
	o := &initOpts{server: "root@1.2.3.4", adminEmail: "a@b.com"}
	if err := validateInit(o); err != nil {
		t.Errorf("valid IP-mode init rejected: %v", err)
	}

	// Domain mode auto-fills acme-email.
	o = &initOpts{server: "root@1.2.3.4", adminEmail: "a@b.com", domain: "orbita.x.com"}
	if err := validateInit(o); err != nil {
		t.Fatalf("valid domain init rejected: %v", err)
	}
	if o.acmeEmail != "admin@orbita.x.com" {
		t.Errorf("acme email should default to admin@<domain>, got %q", o.acmeEmail)
	}

	// Missing server / admin-email.
	if err := validateInit(&initOpts{adminEmail: "a@b.com"}); err == nil {
		t.Error("missing server should fail")
	}
	if err := validateInit(&initOpts{server: "root@1.2.3.4"}); err == nil {
		t.Error("missing admin-email should fail")
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	if got := expandHome("~/.ssh/id_ed25519"); got != filepath.Join(home, ".ssh/id_ed25519") {
		t.Errorf("expandHome = %q", got)
	}
	if got := expandHome("/abs/path"); got != "/abs/path" {
		t.Errorf("abs path changed: %q", got)
	}
	if got := expandHome(""); got != "" {
		t.Errorf("empty changed: %q", got)
	}
}

func TestDNSPointsTo(t *testing.T) {
	// localhost resolves to 127.0.0.1 — a stable check that doesn't hit the network.
	if !dnsPointsTo("localhost", "127.0.0.1") {
		t.Error("localhost should point to 127.0.0.1")
	}
	if dnsPointsTo("localhost", "8.8.8.8") {
		t.Error("localhost should not point to 8.8.8.8")
	}
	if dnsPointsTo("this-domain-does-not-exist.invalid", "1.2.3.4") {
		t.Error("nonexistent domain should not match")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if firstNonEmpty("a", "b") != "a" {
		t.Error("should return a")
	}
	if firstNonEmpty("", "b") != "b" {
		t.Error("should return b")
	}
	if firstNonEmpty("  ", "b") != "b" {
		t.Error("whitespace should fall through to b")
	}
}
