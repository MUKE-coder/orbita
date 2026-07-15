package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"

	"github.com/orbita-sh/orbita/cmd/grit/internal/hosts"
	"github.com/orbita-sh/orbita/cmd/grit/internal/orbita"
	"github.com/orbita-sh/orbita/cmd/grit/internal/project"
	"github.com/orbita-sh/orbita/cmd/grit/internal/ui"
)

func logsCmd() *cobra.Command {
	var host, org, app, role string
	var follow bool
	c := &cobra.Command{
		Use:   "logs",
		Short: "Stream logs from a deployed Grit app service",
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

			// Resolve the target service (default: the API service).
			client := orbita.New(h.APIURL, h.Token)
			st, err := client.GritStatus(cmd.Context(), orgSlug, appName)
			if err != nil {
				ui.ErrorLine("could not resolve app services: "+err.Error(), "")
				return err
			}
			appID := pickServiceAppID(st, role)
			if appID == "" {
				return fmt.Errorf("no service %q found for app %q", role, appName)
			}

			return streamLogs(h.APIURL, h.Token, orgSlug, appID, follow)
		},
	}
	f := c.Flags()
	f.StringVar(&host, "host", "prod", "registered Orbita host name")
	f.StringVar(&org, "org", "", "org slug (defaults to the app name)")
	f.StringVar(&app, "app", "", "grit app name (defaults to grit.yaml app)")
	f.StringVar(&role, "role", "api", "which service (api|web|admin|docs|app)")
	f.BoolVarP(&follow, "follow", "f", true, "follow the log stream")
	return c
}

func pickServiceAppID(st *orbita.DeployResult, role string) string {
	for _, s := range st.Services {
		if s.Role == role {
			return s.AppID
		}
	}
	// Fall back to the first service.
	if len(st.Services) > 0 {
		return st.Services[0].AppID
	}
	return ""
}

// streamLogs connects to Orbita's WebSocket log endpoint and prints lines until
// interrupted.
func streamLogs(apiURL, token, org, appID string, follow bool) error {
	u, err := url.Parse(apiURL)
	if err != nil {
		return err
	}
	scheme := "wss"
	if u.Scheme == "http" {
		scheme = "ws"
	}
	wsURL := fmt.Sprintf("%s://%s/api/v1/orgs/%s/apps/%s/logs/stream?token=%s",
		scheme, u.Host, org, appID, url.QueryEscape(token))

	ui.Header("Streaming logs")
	ui.Live("connected — Ctrl-C to stop")

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		ui.ErrorLine("could not open log stream: "+err.Error(),
			"the WebSocket needs a JWT (not an orb_ key); log in on this host or use the dashboard")
		return err
	}
	defer conn.Close()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			fmt.Print(strings.TrimRight(string(msg), "\n") + "\n")
			if !follow {
				return
			}
		}
	}()

	select {
	case <-interrupt:
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		return nil
	case <-done:
		return nil
	}
}
