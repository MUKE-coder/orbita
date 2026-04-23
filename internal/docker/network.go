package docker

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// GetOrgNetworkName returns the canonical Docker network name for an org.
func GetOrgNetworkName(orgSlug string) string {
	return fmt.Sprintf("orbita-org-%s", orgSlug)
}

// --- package-level client (lazy) for package-scope network helpers ---
//
// OrgService calls these as package functions (not methods on an injected
// client). We honor that contract by initializing a shared client on first
// use from DOCKER_SOCKET (default /var/run/docker.sock).
var (
	defaultClientOnce sync.Once
	defaultClient     *Client
)

func getDefaultClient() *Client {
	defaultClientOnce.Do(func() {
		socket := os.Getenv("DOCKER_SOCKET")
		if socket == "" {
			socket = "/var/run/docker.sock"
		}
		defaultClient = NewClient(socket)
	})
	return defaultClient
}

// CreateOrgNetwork creates an overlay network dedicated to an organization.
// Overlay is chosen so deployed services (which run on Swarm) can share it.
// Idempotent: if the network already exists, returns success.
func CreateOrgNetwork(orgSlug string) error {
	name := GetOrgNetworkName(orgSlug)
	cli := getDefaultClient()
	if cli == nil || cli.cli == nil {
		log.Warn().Str("network", name).Msg("Docker client unavailable; skipping network create")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	labels := map[string]string{
		"orbita.managed": "true",
		"orbita.org":     orgSlug,
	}

	id, err := cli.ensureNetwork(ctx, name, "overlay", labels)
	if err != nil {
		// Overlay requires swarm. If we're not in swarm mode, fall back to bridge.
		log.Warn().Err(err).Str("network", name).Msg("Overlay failed; falling back to bridge")
		id, err = cli.ensureNetwork(ctx, name, "bridge", labels)
		if err != nil {
			return fmt.Errorf("CreateOrgNetwork: %w", err)
		}
	}
	log.Info().Str("network", name).Str("id", id).Msg("Org network ready")
	return nil
}

// DeleteOrgNetwork removes the org's network. No-op if it doesn't exist.
func DeleteOrgNetwork(orgSlug string) error {
	name := GetOrgNetworkName(orgSlug)
	cli := getDefaultClient()
	if cli == nil || cli.cli == nil {
		log.Warn().Str("network", name).Msg("Docker client unavailable; skipping network delete")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := cli.removeNetwork(ctx, name); err != nil {
		return fmt.Errorf("DeleteOrgNetwork: %w", err)
	}
	log.Info().Str("network", name).Msg("Removed org network")
	return nil
}
