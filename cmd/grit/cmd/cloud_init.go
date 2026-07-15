package cmd

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/orbita-sh/orbita/cmd/grit/internal/assets"
	"github.com/orbita-sh/orbita/cmd/grit/internal/hosts"
	"github.com/orbita-sh/orbita/cmd/grit/internal/orbita"
	"github.com/orbita-sh/orbita/cmd/grit/internal/sshx"
	"github.com/orbita-sh/orbita/cmd/grit/internal/ui"
)

type initOpts struct {
	server     string // user@ip[:port] of the fresh VPS
	name       string // friendly host name to register (e.g. "prod")
	domain     string // Orbita dashboard domain (HTTPS)
	acmeEmail  string // Let's Encrypt email
	adminEmail string // super-admin email for the first-run account
	sshKey     string // explicit SSH private key file
	yes        bool   // non-interactive
	skipHarden bool   // skip the vps-harden step (already hardened)
}

func cloudInitCmd() *cobra.Command {
	o := &initOpts{}
	c := &cobra.Command{
		Use:   "init",
		Short: "Harden a fresh VPS, install Orbita, and register it locally",
		Long: "grit cloud init turns a fresh Ubuntu/Debian VPS into a hardened, Grit-aware\n" +
			"Orbita host: it hardens the server (vps-harden), installs Docker + Orbita +\n" +
			"Traefik, bootstraps an orb_ deploy token, and registers the host in\n" +
			"~/.grit/hosts.yaml so `grit deploy --host <name>` can ship to it.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd.Context(), o)
		},
	}
	f := c.Flags()
	f.StringVar(&o.server, "server", "", "target VPS as user@ip[:port] (required)")
	f.StringVar(&o.name, "name", "prod", "friendly host name to register")
	f.StringVar(&o.domain, "domain", "", "Orbita dashboard domain, e.g. orbita.example.com (required)")
	f.StringVar(&o.acmeEmail, "acme-email", "", "Let's Encrypt email (required)")
	f.StringVar(&o.adminEmail, "admin-email", "", "super-admin email for the first-run account (required)")
	f.StringVar(&o.sshKey, "ssh-key", "", "SSH private key file (defaults to agent / ~/.ssh)")
	f.BoolVarP(&o.yes, "yes", "y", false, "non-interactive; accept defaults and don't prompt")
	f.BoolVar(&o.skipHarden, "skip-harden", false, "skip vps-harden (server already hardened)")
	return c
}

func runInit(ctx context.Context, o *initOpts) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := validateInit(o); err != nil {
		ui.ErrorLine(err.Error(), "see `grit cloud init --help`")
		return err
	}

	target, err := sshx.ParseTarget(o.server)
	if err != nil {
		return err
	}

	ui.Header(fmt.Sprintf("Provisioning %s → Orbita host %q", o.server, o.name))
	ui.Field("Domain", ui.URL("https://"+o.domain))
	ui.Field("ACME email", o.acmeEmail)

	// Connect
	ui.StepActive("Connecting over SSH…")
	client, err := sshx.Connect(target, o.sshKey)
	if err != nil {
		ui.StepFail("SSH connection failed")
		ui.ErrorLine(err.Error(), "check the server IP, that your SSH key is authorized, and port "+target.Port)
		return err
	}
	defer client.Close()
	ui.Step("Connected to " + ui.Value(target.Host))

	// 1. Harden (vps-harden.sh --no-dokploy --yes), preserving the 0-100 score.
	if !o.skipHarden {
		if err := hardenStep(client, o); err != nil {
			return err
		}
	} else {
		ui.Info("Skipping harden step (--skip-harden)")
	}

	// 2. Install Docker + Swarm + Orbita (reuse Orbita's install.sh).
	if err := installStep(client, o); err != nil {
		return err
	}

	// 3. Wait for Orbita to answer, then bootstrap the super admin + orb_ token.
	apiURL := "https://" + o.domain
	token, err := bootstrapStep(ctx, apiURL, o)
	if err != nil {
		return err
	}

	// 4. Register the host locally.
	if err := hosts.Set(o.name, hosts.Host{
		APIURL: apiURL,
		Token:  token,
		SSH:    o.server,
	}); err != nil {
		return err
	}
	ui.Step("Registered host " + ui.Value(o.name) + " in ~/.grit/hosts.yaml")

	ui.Success("Orbita is live and ready")
	ui.Field("Dashboard", ui.URL(apiURL))
	ui.Field("Deploy with", ui.Value(fmt.Sprintf("grit deploy --host %s", o.name)))
	return nil
}

func validateInit(o *initOpts) error {
	var missing []string
	if o.server == "" {
		missing = append(missing, "--server")
	}
	if o.domain == "" {
		missing = append(missing, "--domain")
	}
	if o.acmeEmail == "" {
		missing = append(missing, "--acme-email")
	}
	if o.adminEmail == "" {
		missing = append(missing, "--admin-email")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required flags: %s", strings.Join(missing, ", "))
	}
	return nil
}

// hardenStep uploads and runs the vendored vps-harden.sh, streaming its output
// (which ends with the security score out of 100).
func hardenStep(client *sshx.Client, o *initOpts) error {
	ui.Header("Hardening the server")
	remote := "/tmp/vps-harden.sh"
	if err := client.Upload(assets.VPSHarden, remote, "0755"); err != nil {
		ui.StepFail("Failed to upload hardening script")
		return err
	}
	// --no-dokploy: Orbita replaces Dokploy. --yes: non-interactive.
	cmd := fmt.Sprintf("sudo INSTALL_DOKPLOY=no bash %s --no-dokploy --yes", remote)
	if err := client.Run(cmd, prefixWriter("  │ "), prefixWriter("  │ ")); err != nil {
		ui.StepFail("Hardening reported an error")
		ui.ErrorLine(err.Error(), "review the output above; you can re-run with --skip-harden once resolved")
		return err
	}
	ui.Step("Server hardened (see the security score above)")
	return nil
}

// installStep installs Docker (if absent), initializes Swarm, and runs Orbita's
// install.sh non-interactively with the domain + ACME email.
func installStep(client *sshx.Client, o *initOpts) error {
	ui.Header("Installing Docker + Orbita")

	ui.StepActive("Ensuring Docker + Swarm…")
	dockerCmd := "command -v docker >/dev/null 2>&1 || (curl -fsSL https://get.docker.com | sudo sh); " +
		"sudo docker swarm init >/dev/null 2>&1 || true"
	if err := client.Run(dockerCmd, prefixWriter("  │ "), prefixWriter("  │ ")); err != nil {
		ui.StepFail("Docker install failed")
		return err
	}
	ui.Step("Docker + Swarm ready")

	ui.StepActive("Running Orbita install.sh…")
	install := fmt.Sprintf(
		"curl -sSL https://raw.githubusercontent.com/MUKE-coder/orbita/main/install.sh | "+
			"sudo ORBITA_DOMAIN=%q ORBITA_ACME_EMAIL=%q bash -s -- --yes",
		o.domain, o.acmeEmail)
	if err := client.Run(install, prefixWriter("  │ "), prefixWriter("  │ ")); err != nil {
		ui.StepFail("Orbita install failed")
		ui.ErrorLine(err.Error(), "check the output above; common causes: ports 80/443 in use, DNS not pointing at the server")
		return err
	}
	ui.Step("Orbita installed")
	return nil
}

// bootstrapStep waits for Orbita to answer over HTTPS, registers the first user
// (super admin), and creates an orb_ deploy key.
func bootstrapStep(ctx context.Context, apiURL string, o *initOpts) (string, error) {
	ui.Header("Bootstrapping the deploy token")
	client := orbita.New(apiURL, "")

	ui.StepActive("Waiting for Orbita to become healthy…")
	if err := client.WaitHealthy(ctx, 3*time.Minute); err != nil {
		ui.StepFail("Orbita did not become healthy")
		ui.ErrorLine(err.Error(),
			"DNS may not have propagated yet (the Let's Encrypt cert needs "+o.domain+" → this server). "+
				"Once it resolves, re-run with --skip-harden to finish.")
		return "", err
	}
	ui.Step("Orbita is healthy")

	password := o.adminPassword()
	access, err := client.Register(ctx, o.adminEmail, password, "Admin")
	if err != nil {
		// The account may already exist from a previous run — try logging in.
		access, err = client.Login(ctx, o.adminEmail, password)
		if err != nil {
			ui.StepFail("Could not create or log into the super-admin account")
			ui.ErrorLine(err.Error(), "if the admin account exists with a different password, pass GRIT_ADMIN_PASSWORD")
			return "", err
		}
		ui.Step("Logged into existing super-admin account")
	} else {
		ui.Step("Created super-admin account " + ui.Value(o.adminEmail))
		if os.Getenv("GRIT_ADMIN_PASSWORD") == "" {
			ui.Info("Generated admin password: " + ui.Value(password) + "  (save this)")
		}
	}

	authed := orbita.New(apiURL, access)
	token, err := authed.CreateAPIKey(ctx, "grit-cloud", []string{"deploy"})
	if err != nil {
		ui.StepFail("Could not create the orb_ deploy key")
		return "", err
	}
	ui.Step("Created orb_ deploy key")
	return token, nil
}

// adminPassword returns GRIT_ADMIN_PASSWORD or a generated strong one.
func (o *initOpts) adminPassword() string {
	if p := os.Getenv("GRIT_ADMIN_PASSWORD"); p != "" {
		return p
	}
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "Orb-" + hex.EncodeToString(b)
}

// prefixWriter returns a writer that prefixes each line — used to indent
// streamed remote output under the step.
func prefixWriter(prefix string) *linePrefixer {
	return &linePrefixer{prefix: prefix, out: bufio.NewWriter(os.Stdout)}
}

type linePrefixer struct {
	prefix string
	out    *bufio.Writer
	atBOL  bool
	inited bool
}

func (w *linePrefixer) Write(p []byte) (int, error) {
	if !w.inited {
		w.atBOL = true
		w.inited = true
	}
	for _, b := range p {
		if w.atBOL {
			_, _ = w.out.WriteString(w.prefix)
			w.atBOL = false
		}
		_ = w.out.WriteByte(b)
		if b == '\n' {
			w.atBOL = true
			_ = w.out.Flush()
		}
	}
	_ = w.out.Flush()
	return len(p), nil
}
