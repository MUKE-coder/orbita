package docker

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

func GetOrgNetworkName(orgSlug string) string {
	return fmt.Sprintf("orbita-org-%s", orgSlug)
}

func CreateOrgNetwork(orgSlug string) error {
	networkName := GetOrgNetworkName(orgSlug)
	// TODO: real impl — create Docker overlay/bridge network via Docker SDK
	log.Info().Str("network", networkName).Msg("Created org network (stub)")
	return nil
}

func DeleteOrgNetwork(orgSlug string) error {
	networkName := GetOrgNetworkName(orgSlug)
	// TODO: real impl — remove Docker network via Docker SDK
	log.Info().Str("network", networkName).Msg("Deleted org network (stub)")
	return nil
}
