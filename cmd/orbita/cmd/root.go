package cmd

import (
	"github.com/spf13/cobra"
)

// Root builds the top-level `orbita` command.
//
// Provisioning commands sit at the top level rather than under a `cloud` group:
// this is Orbita's own binary, so `orbita cloud init` would just be noise.
func Root() *cobra.Command {
	root := &cobra.Command{
		Use:   "orbita",
		Short: "Orbita — a self-hosted PaaS for any containerised app",
		Long: "Orbita turns a fresh VPS into a hardened, multi-tenant PaaS, then deploys any\n" +
			"containerisable app to it — a Dockerfile repo, a Compose file, a raw Docker\n" +
			"image, or anything Nixpacks can build (Laravel, Django, Rails, Node, static).\n\n" +
			"Grit apps are a zero-config fast path: Orbita detects grit.json and reuses the\n" +
			"Dockerfiles Grit ships. Grit is not required — Orbita stands alone.",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	// Provision + manage a host.
	root.AddCommand(initCmd())
	root.AddCommand(statusCmd())
	root.AddCommand(dashboardCmd())
	root.AddCommand(hostsCmd())
	root.AddCommand(githubAuthCmd())
	// Ship + operate an app.
	root.AddCommand(deployCmd())
	root.AddCommand(logsCmd())
	root.AddCommand(rollbackCmd())
	return root
}
