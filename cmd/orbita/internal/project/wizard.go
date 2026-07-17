package project

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/orbita-sh/orbita/internal/grit"
)

// Wizard generates an orbita.yaml interactively from the detected project shape
// (the first-run experience). It reads grit.json to seed defaults and prompts
// only for what Orbita can't infer: repo, domains, addons, env file.
func Wizard(dir string) (*grit.Manifest, error) {
	gj := DetectGritJSON(dir)
	if gj == nil {
		return nil, fmt.Errorf("no grit.json in %s — orbita deploy must run from a Grit app root", dir)
	}
	if !grit.IsVPSDeployable(gj.Architecture) {
		return nil, fmt.Errorf("architecture %q is not VPS-deployable (mobile apps ship to app stores)", gj.Architecture)
	}

	r := bufio.NewReader(os.Stdin)
	fmt.Println("No orbita.yaml found — let's create one.")
	fmt.Printf("Detected a %q Grit app.\n\n", gj.Architecture)

	appName := prompt(r, "App name", defaultAppName(dir))
	repo := prompt(r, "GitHub repo (owner/name)", "")
	for !validRepo(repo) {
		fmt.Println("  Please enter owner/name, e.g. MUKE-coder/rental-manager")
		repo = prompt(r, "GitHub repo (owner/name)", "")
	}
	branch := prompt(r, "Branch", "main")

	m := &grit.Manifest{App: appName, Repo: repo, Branch: branch}

	// Domains per the mode's services.
	services, _ := grit.DeriveServices(gj)
	for _, s := range services {
		switch s.Role {
		case grit.RoleApp:
			m.Domains.Web = prompt(r, "App domain (serves SPA + API)", "")
		case grit.RoleWeb:
			m.Domains.Web = prompt(r, "Web domain (root site)", "")
		case grit.RoleAdmin:
			m.Domains.Admin = prompt(r, "Admin domain", "")
		case grit.RoleAPI:
			m.Domains.API = prompt(r, "API domain", "")
		case grit.RoleDocs:
			m.Domains.Docs = prompt(r, "Docs domain (optional)", "")
		}
	}

	// Addons.
	addonAns := prompt(r, "Addons (comma-separated: postgres,redis,minio)", "postgres,redis")
	for _, a := range strings.Split(addonAns, ",") {
		a = strings.TrimSpace(a)
		if grit.ValidAddons[a] {
			m.Addons = append(m.Addons, a)
		}
	}

	// Env file.
	m.Env.From = prompt(r, "Env file (values encrypted into Orbita)", ".env.production")

	if err := grit.ValidateForDeploy(m, gj); err != nil {
		return nil, err
	}
	return m, nil
}

func prompt(r *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

func defaultAppName(dir string) string {
	parts := strings.Split(strings.TrimRight(dir, "/\\"), string(os.PathSeparator))
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "app"
}

func validRepo(s string) bool {
	parts := strings.SplitN(s, "/", 2)
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}
