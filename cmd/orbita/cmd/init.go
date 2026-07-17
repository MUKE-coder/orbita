package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/orbita-sh/orbita/cmd/orbita/internal/assets"
	"github.com/orbita-sh/orbita/cmd/orbita/internal/hosts"
	"github.com/orbita-sh/orbita/cmd/orbita/internal/orbita"
	"github.com/orbita-sh/orbita/cmd/orbita/internal/sshx"
	"github.com/orbita-sh/orbita/cmd/orbita/internal/ui"
)

type initOpts struct {
	// Connection to the fresh server (root by default).
	server   string // user@ip[:port]; wizard builds root@<ip>
	sshKey   string // private key for the initial connect
	password string // root password for the initial connect (wizard/env only)

	// What to build.
	name       string // friendly host name to register (e.g. "prod")
	domain     string // dashboard domain; empty → install on the IP
	acmeEmail  string // Let's Encrypt email (domain mode)
	adminEmail string // super-admin email for the first-run account

	// Hardening / deploy user.
	deployUser   string // new sudo user (default "deploy")
	genKey       bool   // generate a new deploy key locally
	deployPubKey string // pasted deploy public key (when not generating)
	skipHarden   bool

	// forgetHostKey removes this server's entry from the local ~/.ssh/known_hosts
	// (ssh-keygen -R) when it's stale — e.g. the VPS was rebuilt on the same IP.
	forgetHostKey bool

	yes bool // non-interactive

	// computed
	useIP         bool
	deployKeyPath string
}

func initCmd() *cobra.Command {
	o := &initOpts{}
	c := &cobra.Command{
		Use:   "init",
		Short: "Set up Orbita on your server — interactive by default",
		Long: "orbita init walks you through turning a fresh Ubuntu/Debian VPS into a\n" +
			"hardened Orbita host: it secures the server (new sudo user + SSH\n" +
			"keys, firewall, Fail2ban), installs Docker + Orbita + Traefik on an HTTPS\n" +
			"subdomain (or the server IP if DNS isn't ready), bootstraps your admin login\n" +
			"and an orb_ deploy token, and registers the host in ~/.orbita/hosts.yaml.\n\n" +
			"Run it with no flags for the guided wizard, or pass flags + --yes to automate.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd.Context(), o)
		},
	}
	f := c.Flags()
	f.StringVar(&o.server, "server", "", "target VPS as user@ip[:port] (wizard asks if omitted)")
	f.StringVar(&o.name, "name", "prod", "friendly host name to register")
	f.StringVar(&o.domain, "domain", "", "Orbita dashboard domain (empty = use the server IP)")
	f.StringVar(&o.acmeEmail, "acme-email", "", "Let's Encrypt email (domain mode)")
	f.StringVar(&o.adminEmail, "admin-email", "", "super-admin email for the first-run account")
	f.StringVar(&o.deployUser, "deploy-user", "deploy", "sudo user to create on the server")
	f.StringVar(&o.sshKey, "ssh-key", "", "SSH private key for the initial connect (defaults to agent / ~/.ssh)")
	f.StringVar(&o.deployPubKey, "deploy-pubkey", "", "deploy user's public key (else one is generated)")
	f.BoolVarP(&o.yes, "yes", "y", false, "non-interactive; use flags + defaults, don't prompt")
	f.BoolVar(&o.skipHarden, "skip-harden", false, "skip hardening (server already hardened)")
	f.BoolVar(&o.forgetHostKey, "forget-host-key", false,
		"if this IP has a stale key in ~/.ssh/known_hosts (rebuilt VPS), remove it (ssh-keygen -R)")
	return c
}

func runInit(ctx context.Context, o *initOpts) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Interactive wizard unless --yes (or all required flags supplied).
	if !o.yes {
		if err := wizard(o); err != nil {
			if err == errAborted {
				return nil // user chose to stop — not an error
			}
			return err
		}
	}
	if err := validateInit(o); err != nil {
		ui.ErrorLine(err.Error(), "run `orbita init` with no flags for the guided setup")
		return err
	}

	target, err := sshx.ParseTarget(o.server)
	if err != nil {
		return err
	}
	o.useIP = o.domain == ""
	apiURL := "https://" + o.domain
	if o.useIP {
		apiURL = fmt.Sprintf("http://%s:8080", target.Host)
	}

	ui.Header(fmt.Sprintf("Setting up Orbita on %s", target.Host))
	if o.useIP {
		ui.Field("Dashboard", ui.URL(apiURL)+"  "+ui.Value("(IP mode — add a domain later for HTTPS)"))
	} else {
		ui.Field("Dashboard", ui.URL(apiURL))
		ui.Field("ACME email", o.acmeEmail)
	}
	ui.Field("Deploy user", ui.Value(o.deployUser))

	// 1. Connect (root, via key or password).
	ui.StepActive("Connecting over SSH…")
	client, err := sshx.Connect(target, sshx.Auth{KeyFile: expandHome(o.sshKey), Password: o.password})
	if err != nil {
		ui.StepFail("SSH connection failed")
		ui.ErrorLine(err.Error(), "check the IP, that your key/password is right, and that port "+target.Port+" is open")
		return err
	}
	defer client.Close()
	ui.Step("Connected to " + ui.Value(target.Host))

	// 1b. We connect trust-on-first-use, but the operator will `ssh` into this
	//     box themselves later — and OpenSSH refuses a changed host key. Offer
	//     to clear a stale entry now, while we know the server's real key.
	hostKeyStep(client, target, o)

	// 2. Update the system.
	ui.StepActive("Updating the server (apt update && upgrade)…")
	if err := client.Run("export DEBIAN_FRONTEND=noninteractive; apt-get update -y && apt-get upgrade -y", indent(), indent()); err != nil {
		ui.StepFail("System update failed")
		return err
	}
	ui.Step("Server updated")

	// 3. Harden FIRST (create the deploy user + key, firewall, Fail2ban, lock
	//    down root). The current root session persists after lockdown, so we keep
	//    installing on it.
	if !o.skipHarden {
		if err := hardenStep(client, o); err != nil {
			return err
		}
		// Passwordless sudo for the deploy user (so re-runs / later ops work),
		// and open the ports Orbita needs (harden's UFW only opened SSH).
		post := fmt.Sprintf(
			"echo '%s ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/90-orbita && chmod 440 /etc/sudoers.d/90-orbita; "+
				"ufw allow 80/tcp >/dev/null 2>&1; ufw allow 443/tcp >/dev/null 2>&1; ufw allow 8080/tcp >/dev/null 2>&1; ufw reload >/dev/null 2>&1 || true",
			o.deployUser)
		_ = client.Run(post, indent(), indent())
	}

	// 4. Install Docker + Swarm + Orbita.
	if err := installStep(client, o); err != nil {
		return err
	}

	// 5. Wait for Orbita, then bootstrap the super-admin + orb_ token.
	token, adminPass, err := bootstrapStep(ctx, apiURL, o)
	if err != nil {
		return err
	}

	// 6. Register the host locally (connect as the deploy user going forward).
	sshTarget := fmt.Sprintf("%s@%s:%s", o.deployUser, target.Host, target.Port)
	if err := hosts.Set(o.name, hosts.Host{APIURL: apiURL, Token: token, SSH: sshTarget}); err != nil {
		return err
	}
	ui.Step("Registered host " + ui.Value(o.name) + " in ~/.orbita/hosts.yaml")

	// 7. Done — print how to log in.
	ui.Success("Orbita is live")
	ui.Field("Dashboard", ui.URL(apiURL))
	ui.Field("Login email", ui.Value(o.adminEmail))
	if adminPass != "" {
		ui.Field("Password", ui.Value(adminPass)+"  "+ui.Value("(save this — shown once)"))
	}
	if o.deployKeyPath != "" {
		ui.Field("Deploy key", ui.Value(o.deployKeyPath))
	}
	fmt.Println()
	ui.Info("Next: open the dashboard and log in, or deploy an app from its directory:")
	ui.Info("  " + ui.Value("orbita deploy --host "+o.name))
	return nil
}

// wizard fills the options interactively (only asks for what's missing).
func wizard(o *initOpts) error {
	ui.Header("Grit Cloud — set up Orbita on your server")

	if o.server == "" {
		if !ui.Confirm("Do you already have a server (a fresh Ubuntu 24.04 VPS)?", true) {
			ui.Info("Get a fresh Ubuntu 24.04 VPS from Hetzner, DigitalOcean, or Vultr, then run `orbita init` again.")
			return errAborted
		}
		ip := ui.AskRequired("Server IP address")
		o.server = "root@" + ip
	}
	target, err := sshx.ParseTarget(o.server)
	if err != nil {
		return err
	}

	// Reachability check.
	ui.StepActive("Checking the server is reachable…")
	if err := sshx.TestReachable(target, 8*time.Second); err != nil {
		ui.StepFail("Couldn't reach " + target.Addr())
		if !ui.Confirm("Continue anyway?", false) {
			return errAborted
		}
	} else {
		ui.Step("Server is reachable")
	}

	// Initial auth (how they log in as root right now).
	if o.sshKey == "" && o.password == "" {
		switch ui.Select("How do you log into the server right now?", []string{
			"I have a root password (from my hosting provider)",
			"I added an SSH key when I created the VPS",
		}, 0) {
		case 0:
			o.password = ui.Secret("Root password:")
		case 1:
			o.sshKey = ui.Ask("Path to your SSH private key", "~/.ssh/id_ed25519")
		}
	}

	// Deploy user + key.
	o.deployUser = ui.Ask("Create a secure deploy user named", firstNonEmpty(o.deployUser, "deploy"))
	if o.deployPubKey == "" {
		switch ui.Select("SSH key for the deploy user:", []string{
			"Generate a new key for me (recommended)",
			"Paste an existing public key",
		}, 0) {
		case 0:
			o.genKey = true
		case 1:
			o.deployPubKey = ui.AskRequired("Paste the public key (ssh-ed25519 …)")
		}
	}

	// Domain (auto-detect DNS; fall back to IP).
	dom := ui.Ask("Domain for the Orbita dashboard (blank = use the server IP)", "")
	if dom != "" {
		ui.StepActive("Checking DNS for " + dom + "…")
		if dnsPointsTo(dom, target.Host) {
			ui.Step("DNS points here — we'll set up HTTPS at https://" + dom)
			o.domain = dom
		} else {
			ui.StepFail("DNS for " + dom + " doesn't point to " + target.Host + " yet")
			ui.Info("We can install on the IP now and you can add the domain later from the dashboard.")
			if ui.Confirm("Install on http://"+target.Host+":8080 for now?", true) {
				o.domain = ""
			} else {
				ui.Info("Point an A-record for " + dom + " at " + target.Host + ", wait for DNS, then re-run.")
				return errAborted
			}
		}
	}

	// Emails.
	if o.domain != "" && o.acmeEmail == "" {
		o.acmeEmail = ui.Ask("Email for Let's Encrypt (TLS certificates)", "admin@"+o.domain)
	}
	if o.adminEmail == "" {
		o.adminEmail = ui.AskRequired("Your email for the Orbita admin login")
	}

	// Summary.
	ui.Header("Ready to go")
	ui.Field("Server", ui.Value(o.server))
	if o.domain != "" {
		ui.Field("Dashboard", ui.URL("https://"+o.domain))
	} else {
		ui.Field("Dashboard", ui.URL("http://"+target.Host+":8080")+" "+ui.Value("(IP mode)"))
	}
	ui.Field("Secure server", "yes — deploy user, SSH keys, firewall, Fail2ban")
	if !ui.Confirm("Proceed?", true) {
		return errAborted
	}
	return nil
}

var errAborted = fmt.Errorf("aborted")

func validateInit(o *initOpts) error {
	var missing []string
	if o.server == "" {
		missing = append(missing, "server")
	}
	if o.adminEmail == "" {
		missing = append(missing, "admin-email")
	}
	if o.domain != "" && o.acmeEmail == "" {
		o.acmeEmail = "admin@" + o.domain
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required values: %s", strings.Join(missing, ", "))
	}
	return nil
}

// hostKeyStep warns when ~/.ssh/known_hosts on this machine holds a different
// key for the server, and offers to drop it (ssh-keygen -R).
//
// grit itself never trips over this — we connect trust-on-first-use — but the
// operator's own `ssh deploy@host` will hard-fail with "REMOTE HOST
// IDENTIFICATION HAS CHANGED" until the stale entry goes. Rebuilding a VPS on
// the same IP is the usual cause. Nothing here is fatal, so no error return.
func hostKeyStep(client *sshx.Client, target sshx.Target, o *initOpts) {
	state, err := client.KnownHostState()
	if err != nil || state != sshx.HostKeyMismatch {
		return // absent/match, or we couldn't read known_hosts — nothing to do
	}

	entry := sshx.KnownHostEntry(target)
	ui.WarnLine(
		"Your machine has a different saved SSH key for "+entry,
		"the VPS was probably rebuilt on this IP — your own `ssh` will refuse to connect until the old key is removed",
	)

	// --forget-host-key removes it outright; otherwise ask. In non-interactive
	// runs without the flag we only tell them the command — never edit their
	// known_hosts behind their back.
	remove := o.forgetHostKey
	if !remove && !o.yes {
		remove = ui.Confirm("Remove the old key from ~/.ssh/known_hosts?", true)
	}
	if !remove {
		ui.Field("Fix it later with", ui.Value(sshx.ForgetHostCommand(target)))
		return
	}

	if _, err := sshx.ForgetHost(target); err != nil {
		ui.StepFail("Couldn't update known_hosts: " + err.Error())
		ui.Field("Run this yourself", ui.Value(sshx.ForgetHostCommand(target)))
		return
	}
	ui.Step("Removed the stale host key for " + ui.Value(entry))
}

// hardenStep uploads and runs the vendored vps-harden.sh with the deploy user +
// public key, streaming its output (which ends with the 0–100 security score).
func hardenStep(client *sshx.Client, o *initOpts) error {
	ui.Header("Securing the server")

	// Resolve the deploy public key: generate one locally (private key stays on
	// the laptop) or use the pasted one.
	pubKey := o.deployPubKey
	if o.genKey {
		kp, err := sshx.GenerateEd25519(o.deployUser + "@orbita")
		if err != nil {
			return err
		}
		path, err := kp.SaveTo("orbita_" + o.deployUser)
		if err != nil {
			return err
		}
		pubKey = kp.PublicAuthorized
		o.deployKeyPath = path
		ui.Step("Generated a deploy key → " + ui.Value(path))
	}

	remote := "/tmp/vps-harden.sh"
	if err := client.Upload(assets.VPSHarden, remote, "0755"); err != nil {
		ui.StepFail("Failed to upload the hardening script")
		return err
	}
	// --no-dokploy: Orbita replaces Dokploy. --yes: non-interactive.
	env := fmt.Sprintf("INSTALL_DOKPLOY=no NEW_USER=%q SSH_PORT=22 SSH_PUBKEY=%q", o.deployUser, strings.TrimSpace(pubKey))
	cmd := fmt.Sprintf("%s bash %s --no-dokploy --yes", env, remote)
	if err := client.Run(cmd, indent(), indent()); err != nil {
		ui.StepFail("Hardening reported an error")
		ui.ErrorLine(err.Error(), "review the output above; you can re-run with --skip-harden once resolved")
		return err
	}
	ui.Step("Server hardened (security score above) — deploy user " + ui.Value(o.deployUser) + " with key-only SSH")
	return nil
}

// installStep installs Docker (if absent), inits Swarm, and runs Orbita's
// install.sh non-interactively. Domain mode uses HTTPS via Traefik; IP mode
// installs on :8080 without TLS.
func installStep(client *sshx.Client, o *initOpts) error {
	ui.Header("Installing Docker + Orbita")

	ui.StepActive("Ensuring Docker + Swarm…")
	dockerCmd := "command -v docker >/dev/null 2>&1 || (curl -fsSL https://get.docker.com | sh); docker swarm init >/dev/null 2>&1 || true"
	if err := client.Run(dockerCmd, indent(), indent()); err != nil {
		ui.StepFail("Docker install failed")
		return err
	}
	ui.Step("Docker + Swarm ready")

	ui.StepActive("Running Orbita install.sh…")
	// Domain mode passes the real domain + ACME email; IP mode passes the IP as
	// the domain, which install.sh detects and installs without TLS on :8080.
	target, _ := sshx.ParseTarget(o.server)
	domainArg := o.domain
	acme := o.acmeEmail
	if o.useIP {
		domainArg = target.Host
		acme = "admin@localhost"
	}
	install := fmt.Sprintf(
		"curl -sSL https://raw.githubusercontent.com/MUKE-coder/orbita/main/install.sh | ORBITA_DOMAIN=%q ORBITA_ACME_EMAIL=%q bash -s -- --yes",
		domainArg, acme)
	if err := client.Run(install, indent(), indent()); err != nil {
		ui.StepFail("Orbita install failed")
		ui.ErrorLine(err.Error(), "check the output above; common causes: ports 80/443 in use, DNS not pointing at the server")
		return err
	}
	ui.Step("Orbita installed")
	return nil
}

// bootstrapStep waits for Orbita, registers the first user (super admin), and
// creates an orb_ deploy key. Returns (token, generatedPassword).
func bootstrapStep(ctx context.Context, apiURL string, o *initOpts) (string, string, error) {
	ui.Header("Bootstrapping your login + deploy token")
	client := orbita.New(apiURL, "")

	ui.StepActive("Waiting for Orbita to become healthy…")
	if err := client.WaitHealthy(ctx, 3*time.Minute); err != nil {
		ui.StepFail("Orbita did not become healthy")
		hint := "check the install output above."
		if !o.useIP {
			hint = "DNS for " + o.domain + " must point at the server for the TLS cert. Once it resolves, re-run with --skip-harden to finish."
		}
		ui.ErrorLine(err.Error(), hint)
		return "", "", err
	}
	ui.Step("Orbita is healthy")

	password := os.Getenv("GRIT_ADMIN_PASSWORD")
	generated := ""
	if password == "" {
		b := make([]byte, 12)
		_, _ = rand.Read(b)
		password = "Orb-" + hex.EncodeToString(b)
		generated = password
	}

	access, err := client.Register(ctx, o.adminEmail, password, "Admin")
	if err != nil {
		access, err = client.Login(ctx, o.adminEmail, password)
		if err != nil {
			ui.StepFail("Could not create or log into the admin account")
			ui.ErrorLine(err.Error(), "if the account exists with a different password, set GRIT_ADMIN_PASSWORD and re-run with --skip-harden")
			return "", "", err
		}
		ui.Step("Logged into the existing admin account")
		generated = "" // existing account — don't claim a new password
	} else {
		ui.Step("Created your admin account " + ui.Value(o.adminEmail))
	}

	authed := orbita.New(apiURL, access)
	token, err := authed.CreateAPIKey(ctx, "grit-cloud", []string{"deploy"})
	if err != nil {
		ui.StepFail("Could not create the orb_ deploy key")
		return "", "", err
	}
	ui.Step("Created an orb_ deploy key")
	return token, generated, nil
}

// --- helpers ---

func dnsPointsTo(domain, serverIP string) bool {
	ips, err := net.LookupHost(domain)
	if err != nil {
		return false
	}
	for _, ip := range ips {
		if ip == serverIP {
			return true
		}
	}
	return false
}

func expandHome(p string) string {
	if p == "" {
		return ""
	}
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

// indent streams remote command output indented under the current step.
func indent() *linePrefixer { return &linePrefixer{prefix: "  │ ", atBOL: true} }

// linePrefixer writes each line of streamed remote output with a prefix so it
// nests under the current step.
type linePrefixer struct {
	prefix string
	atBOL  bool
}

func (w *linePrefixer) Write(p []byte) (int, error) {
	for _, b := range p {
		if w.atBOL {
			fmt.Print(w.prefix)
			w.atBOL = false
		}
		fmt.Printf("%c", b)
		if b == '\n' {
			w.atBOL = true
		}
	}
	return len(p), nil
}
