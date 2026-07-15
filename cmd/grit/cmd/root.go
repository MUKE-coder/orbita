package cmd

import (
	"github.com/spf13/cobra"
)

// Root builds the top-level `grit` command with the Grit Cloud subcommands.
// In the real Grit CLI these are grafted onto the existing root; kept minimal
// here so they don't collide with the framework's own commands.
func Root() *cobra.Command {
	root := &cobra.Command{
		Use:           "grit",
		Short:         "Grit Cloud — deploy Grit apps to a self-hosted Orbita control plane",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.AddCommand(cloudCmd())
	root.AddCommand(deployCmd())
	root.AddCommand(logsCmd())
	root.AddCommand(rollbackCmd())
	return root
}
