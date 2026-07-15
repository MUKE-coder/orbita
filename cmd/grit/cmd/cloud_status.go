package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/orbita-sh/orbita/cmd/grit/internal/hosts"
	"github.com/orbita-sh/orbita/cmd/grit/internal/orbita"
	"github.com/orbita-sh/orbita/cmd/grit/internal/ui"
)

func cloudStatusCmd() *cobra.Command {
	var host string
	c := &cobra.Command{
		Use:   "status",
		Short: "Show an Orbita host's health and platform metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			h, err := hosts.Resolve(host)
			if err != nil {
				ui.ErrorLine(err.Error(), "run `grit cloud init` to register a host")
				return err
			}

			ui.Header(fmt.Sprintf("Orbita host %q", host))
			ui.Field("API URL", ui.URL(h.APIURL))

			client := orbita.New(h.APIURL, h.Token)
			ctx := context.Background()

			health, err := client.Health(ctx)
			if err != nil {
				ui.Field("Health", ui.Status("unreachable"))
				ui.ErrorLine("could not reach Orbita: "+err.Error(),
					"check the server is up and the API URL is correct")
				return err
			}
			status, _ := health["status"].(string)
			ui.Field("Health", ui.Status(status))
			if v, ok := health["version"].(string); ok {
				ui.Field("Version", ui.Value(v))
			}

			// Platform metrics require super-admin; the orb_ key may be scoped
			// narrower, so treat failure as non-fatal.
			if m, err := client.PlatformMetrics(ctx); err == nil {
				printMetric(m, "organizations", "Organizations")
				printMetric(m, "total_apps", "Apps")
				printMetric(m, "total_databases", "Databases")
				printMetric(m, "nodes", "Nodes")
			}
			return nil
		},
	}
	c.Flags().StringVar(&host, "host", "prod", "registered host name")
	return c
}

func printMetric(m map[string]interface{}, key, label string) {
	if v, ok := m[key]; ok {
		ui.Field(label, ui.Value(fmt.Sprintf("%v", v)))
	}
}
