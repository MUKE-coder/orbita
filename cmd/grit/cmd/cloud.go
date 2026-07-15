package cmd

import (
	"github.com/spf13/cobra"
)

// cloudCmd is the `grit cloud` command group: server provisioning + host
// management for the Orbita control plane.
func cloudCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "cloud",
		Short: "Provision and manage Orbita control-plane hosts",
		Long: "grit cloud turns a fresh VPS into a hardened, Grit-aware Orbita host and\n" +
			"registers it locally so `grit deploy --host <name>` can ship to it.",
	}
	c.AddCommand(cloudInitCmd())
	c.AddCommand(cloudStatusCmd())
	c.AddCommand(cloudDashboardCmd())
	c.AddCommand(cloudHostsCmd())
	c.AddCommand(cloudGithubAuthCmd())
	return c
}
