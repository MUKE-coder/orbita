package orbita

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TestBootstrapAgainstLiveOrbita exercises the exact client path `grit cloud
// init` uses (health → login → create orb_ key) against a running Orbita.
// Skipped unless ORBITA_LIVE_URL + ORBITA_LIVE_EMAIL/PASSWORD are set.
func TestBootstrapAgainstLiveOrbita(t *testing.T) {
	base := os.Getenv("ORBITA_LIVE_URL")
	email := os.Getenv("ORBITA_LIVE_EMAIL")
	pass := os.Getenv("ORBITA_LIVE_PASSWORD")
	if base == "" || email == "" || pass == "" {
		t.Skip("set ORBITA_LIVE_URL/EMAIL/PASSWORD to run the live bootstrap test")
	}
	ctx := context.Background()
	c := New(base, "")

	if err := c.WaitHealthy(ctx, 10*time.Second); err != nil {
		t.Fatalf("health: %v", err)
	}

	access, err := c.Login(ctx, email, pass)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if access == "" {
		t.Fatal("empty access token")
	}

	authed := New(base, access)
	key, err := authed.CreateAPIKey(ctx, "grit-cloud-livetest", []string{"deploy"})
	if err != nil {
		t.Fatalf("create key: %v", err)
	}
	if !strings.HasPrefix(key, "orb_") {
		t.Fatalf("expected orb_ key, got %q", key)
	}

	// The key must actually authenticate (status path).
	keyed := New(base, key)
	if _, err := keyed.Health(ctx); err != nil {
		t.Fatalf("health with key: %v", err)
	}
}
