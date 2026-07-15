package cmd

import (
	"github.com/spf13/cobra"

	"github.com/orbita-sh/orbita/cmd/grit/internal/hosts"
	"github.com/orbita-sh/orbita/cmd/grit/internal/ui"
)

// cloudHostsCmd lists and removes registered hosts.
func cloudHostsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "hosts",
		Short: "List registered Orbita hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := hosts.Load()
			if err != nil {
				return err
			}
			if len(f.Hosts) == 0 {
				ui.Info("No hosts registered. Run `grit cloud init` to add one.")
				return nil
			}
			ui.Header("Registered hosts")
			for _, name := range f.Names() {
				h := f.Hosts[name]
				ui.Field(name, ui.URL(h.APIURL))
			}
			return nil
		},
	}
	c.AddCommand(&cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a registered host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := hosts.Load()
			if err != nil {
				return err
			}
			delete(f.Hosts, args[0])
			if err := f.Save(); err != nil {
				return err
			}
			ui.Step("Removed host " + ui.Value(args[0]))
			return nil
		},
	})
	return c
}
