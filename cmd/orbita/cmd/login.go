package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/orbita-sh/orbita/cmd/orbita/internal/hosts"
	"github.com/orbita-sh/orbita/cmd/orbita/internal/orbita"
	"github.com/orbita-sh/orbita/cmd/orbita/internal/ui"
)

// loginCmd registers a host the CLI can deploy to, for people who installed
// Orbita some other way than `orbita init` (the install.sh one-liner, the
// dashboard, an existing server). It logs in with the admin account, mints a
// deploy token, and writes it to ~/.orbita/hosts.yaml — the Vercel-style
// `login` a lot of people reach for first.
func loginCmd() *cobra.Command {
	var name, email, password string
	c := &cobra.Command{
		Use:   "login [https://orbita.example.com]",
		Short: "Register an Orbita host so the CLI can deploy to it",
		Long: "orbita login authenticates against a running Orbita, creates a deploy token,\n" +
			"and saves it as a named host in ~/.orbita/hosts.yaml. Use it when Orbita was\n" +
			"installed without `orbita init` (e.g. the install.sh one-liner or the dashboard).",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiURL := ""
			if len(args) == 1 {
				apiURL = args[0]
			}
			if apiURL == "" {
				apiURL = ui.AskRequired("Orbita URL (e.g. https://orbita.example.com)")
			}
			apiURL = strings.TrimRight(strings.TrimSpace(apiURL), "/")
			if !strings.HasPrefix(apiURL, "http://") && !strings.HasPrefix(apiURL, "https://") {
				apiURL = "https://" + apiURL
			}

			client := orbita.New(apiURL, "")
			ctx := context.Background()

			// Fail early with a clear message if the URL isn't a live Orbita.
			if _, err := client.Health(ctx); err != nil {
				ui.ErrorLine("couldn't reach Orbita at "+apiURL+": "+err.Error(),
					"check the URL, and that the server is up and reachable")
				return err
			}

			if email == "" {
				email = ui.AskRequired("Admin email")
			}
			if password == "" {
				password = ui.Secret("Password:")
			}

			token, err := client.Login(ctx, email, password)
			if err != nil {
				ui.ErrorLine("login failed: "+err.Error(), "check the email and password")
				return err
			}

			// Mint a scoped deploy token so the stored credential isn't the admin
			// session itself. `deploy` covers reconcile + build; fall back to the
			// login token if key creation isn't permitted.
			authed := orbita.New(apiURL, token)
			deployToken, err := authed.CreateAPIKey(ctx, "orbita-cli", []string{"deploy"})
			if err != nil || deployToken == "" {
				ui.WarnLine("could not mint a deploy token: "+errText(err),
					"storing the session token instead — re-run `orbita login` if deploys start failing")
				deployToken = token
			}

			if err := hosts.Set(name, hosts.Host{APIURL: apiURL, Token: deployToken}); err != nil {
				return err
			}
			ui.Success("Logged in")
			ui.Field("Host", ui.Value(name))
			ui.Field("API URL", ui.URL(apiURL))
			fmt.Println()
			ui.Info("Deploy from a project directory with: " + ui.Value("orbita deploy --host "+name))
			return nil
		},
	}
	f := c.Flags()
	f.StringVar(&name, "name", "prod", "name to register this host under")
	f.StringVar(&email, "email", "", "admin email (prompted if omitted)")
	f.StringVar(&password, "password", "", "admin password (prompted if omitted; env-safer to omit)")
	return c
}

func errText(err error) string {
	if err == nil {
		return "unknown error"
	}
	return err.Error()
}
