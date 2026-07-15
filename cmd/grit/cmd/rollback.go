package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/orbita-sh/orbita/cmd/grit/internal/hosts"
	"github.com/orbita-sh/orbita/cmd/grit/internal/orbita"
	"github.com/orbita-sh/orbita/cmd/grit/internal/project"
	"github.com/orbita-sh/orbita/cmd/grit/internal/ui"
)

func rollbackCmd() *cobra.Command {
	var host, org, app string
	c := &cobra.Command{
		Use:   "rollback",
		Short: "Revert a Grit app to its previous deploy",
		RunE: func(cmd *cobra.Command, args []string) error {
			h, err := hosts.Resolve(host)
			if err != nil {
				ui.ErrorLine(err.Error(), "run `grit cloud init` first")
				return err
			}
			appName, orgSlug := app, org
			if appName == "" || orgSlug == "" {
				if dir, _ := os.Getwd(); dir != "" {
					if p, err := project.Load(dir); err == nil {
						if appName == "" {
							appName = p.Manifest.App
						}
						if orgSlug == "" {
							orgSlug = slugify(p.Manifest.App)
						}
					}
				}
			}
			if appName == "" || orgSlug == "" {
				return fmt.Errorf("could not determine app/org — pass --app and --org, or run from the project directory")
			}

			ui.Header(fmt.Sprintf("Rolling back %s on %q", ui.Value(appName), host))
			client := orbita.New(h.APIURL, h.Token)
			res, err := client.GritRollback(context.Background(), orgSlug, appName)
			if err != nil {
				ui.ErrorLine(err.Error(), "")
				return err
			}
			for _, s := range res.Services {
				ui.Field(s.Role, ui.Status(s.Status))
			}
			ui.Success("Rolled back to the previous deploy")
			return nil
		},
	}
	f := c.Flags()
	f.StringVar(&host, "host", "prod", "registered Orbita host name")
	f.StringVar(&org, "org", "", "org slug (defaults to the app name)")
	f.StringVar(&app, "app", "", "grit app name (defaults to grit.yaml app)")
	return c
}
