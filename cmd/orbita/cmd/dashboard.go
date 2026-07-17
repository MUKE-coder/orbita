package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/orbita-sh/orbita/cmd/orbita/internal/hosts"
	"github.com/orbita-sh/orbita/cmd/orbita/internal/sshx"
	"github.com/orbita-sh/orbita/cmd/orbita/internal/ui"
)

// dashboardCmd opens an SSH tunnel to the Orbita dashboard so it can be
// reached privately (matching the existing blog workflow) without exposing the
// admin panel publicly.
func dashboardCmd() *cobra.Command {
	var host string
	var localPort, remotePort int
	c := &cobra.Command{
		Use:   "dashboard",
		Short: "Open an SSH tunnel to the Orbita dashboard (private access)",
		RunE: func(cmd *cobra.Command, args []string) error {
			h, err := hosts.Resolve(host)
			if err != nil {
				ui.ErrorLine(err.Error(), "run `orbita init` first")
				return err
			}
			if h.SSH == "" {
				ui.ErrorLine("no SSH target stored for this host",
					"re-run `orbita init` or add `ssh: user@ip` under this host in ~/.orbita/hosts.yaml")
				return fmt.Errorf("no ssh target")
			}
			t, err := sshx.ParseTarget(h.SSH)
			if err != nil {
				return err
			}

			ui.Header(fmt.Sprintf("Tunneling to %q dashboard", host))
			ui.Field("Local", ui.URL(fmt.Sprintf("http://localhost:%d", localPort)))
			ui.Field("Remote", fmt.Sprintf("%s:%d", t.Host, remotePort))
			ui.Info("Press Ctrl-C to close the tunnel.")

			// Delegate to the system ssh client for a robust, interactive tunnel.
			fwd := fmt.Sprintf("%d:localhost:%d", localPort, remotePort)
			sshArgs := []string{"-N", "-L", fwd, "-p", t.Port, t.User + "@" + t.Host}
			sc := exec.Command("ssh", sshArgs...)
			sc.Stdout = os.Stdout
			sc.Stderr = os.Stderr
			sc.Stdin = os.Stdin
			return sc.Run()
		},
	}
	c.Flags().StringVar(&host, "host", "prod", "registered host name")
	c.Flags().IntVar(&localPort, "local-port", 8080, "local port to forward from")
	c.Flags().IntVar(&remotePort, "remote-port", 8080, "remote Orbita port")
	return c
}
