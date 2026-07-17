package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/orbita-sh/orbita/cmd/orbita/internal/github"
	"github.com/orbita-sh/orbita/cmd/orbita/internal/hosts"
	"github.com/orbita-sh/orbita/cmd/orbita/internal/orbita"
	"github.com/orbita-sh/orbita/cmd/orbita/internal/project"
	"github.com/orbita-sh/orbita/cmd/orbita/internal/ui"
)

func deployCmd() *cobra.Command {
	var host, org, dir string
	var plan, skipPush bool
	c := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy an app to an Orbita host (build → migrate → route → live)",
		Long: "orbita deploy reads orbita.yaml and drives the deploy: reconcile\n" +
			"(org/project/env/app/addons/env/domains) → build → cut over. Idempotent;\n" +
			"re-run to redeploy.\n\n" +
			"The build is chosen by what's in the repo: grit.json (reuse the Dockerfiles\n" +
			"Grit ships) → Dockerfile → an explicit build in orbita.yaml → Nixpacks. Grit\n" +
			"apps additionally run migrations under an advisory lock before cutover.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dir == "" {
				dir, _ = os.Getwd()
			}
			return runDeploy(cmd.Context(), deployOpts{host: host, org: org, dir: dir, plan: plan, skipPush: skipPush})
		},
	}
	f := c.Flags()
	f.StringVar(&host, "host", "prod", "registered Orbita host name")
	f.StringVar(&org, "org", "", "org slug (defaults to the app name)")
	f.StringVar(&dir, "dir", "", "project directory (defaults to cwd)")
	f.BoolVar(&plan, "plan", false, "dry run: show what would be created/changed, don't execute")
	f.BoolVar(&skipPush, "skip-push", false, "skip the GitHub ensure/push step")
	return c
}

type deployOpts struct {
	host, org, dir string
	plan, skipPush bool
}

func runDeploy(ctx context.Context, o deployOpts) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// 1. Resolve the host (Orbita API URL + orb_ token).
	h, err := hosts.Resolve(o.host)
	if err != nil {
		ui.ErrorLine(err.Error(), "run `orbita init` to register a host")
		return err
	}

	// 2. Load orbita.yaml (+ grit.json + env.from), or run the first-run wizard.
	proj, err := project.Load(o.dir)
	if errors.Is(err, project.ErrNoManifest) {
		m, werr := project.Wizard(o.dir)
		if werr != nil {
			ui.ErrorLine(werr.Error(), "")
			return werr
		}
		if werr := project.WriteManifest(o.dir, m); werr != nil {
			return werr
		}
		ui.Step("Wrote orbita.yaml")
		proj, err = project.Load(o.dir)
	}
	if err != nil {
		ui.ErrorLine(err.Error(), "")
		return err
	}

	m := proj.Manifest
	org := o.org
	if org == "" {
		org = slugify(m.App)
	}

	ui.Header(fmt.Sprintf("Deploying %s → host %q", ui.Value(m.App), o.host))
	ui.Field("Repo", ui.Value(m.Repo+"@"+m.BranchOrDefault()))
	ui.Field("Mode", proj.GritJSON.Architecture)

	client := orbita.New(h.APIURL, h.Token)

	// 3. Ensure the org (tenant boundary) exists — needed for both plan and
	//    reconcile since the Grit endpoints are org-scoped.
	orgSlug, err := client.EnsureOrg(ctx, m.App, org)
	if err != nil {
		ui.ErrorLine("could not ensure org: "+err.Error(), "check the host token has deploy scope")
		return err
	}

	// 4. --plan: dry run and stop.
	if o.plan {
		return runPlan(ctx, client, orgSlug, proj)
	}

	// 5. Ensure the GitHub repo exists + push (so Orbita can build from it).
	if !o.skipPush {
		if err := repoStep(ctx, proj); err != nil {
			ui.ErrorLine("repo step failed: "+err.Error(),
				"run `orbita github-auth` to store a token, or pass --skip-push if already pushed")
			return err
		}
	}

	// 6. Reconcile (idempotent): org/project/env/app/addons/env/domains.
	ui.StepActive("Reconciling with Orbita…")
	rec, err := client.GritReconcile(ctx, orgSlug, orbita.GritReconcileRequest{
		GritYAML:  proj.ManifestYAML,
		GritJSON:  proj.GritJSONText,
		EnvValues: proj.EnvValues,
	})
	if err != nil {
		ui.StepFail("Reconcile failed")
		ui.ErrorLine(err.Error(), "")
		return err
	}
	for _, s := range rec.Services {
		verb := "updated"
		if s.Created {
			verb = "created"
		}
		ui.Step(fmt.Sprintf("%s service %s", verb, ui.Value(s.AppName)))
	}
	if len(rec.Addons) > 0 {
		ui.Step("Addons: " + strings.Join(rec.Addons, ", "))
	}

	// 7. Deploy: build → migrate → cut over.
	ui.Header("Building & deploying")
	ui.StepActive("Build → migrate → cutover (this can take a few minutes)…")
	dep, err := client.GritDeploy(ctx, orgSlug, m.App, rec.EnvironmentID)
	if err != nil {
		ui.StepFail("Deploy failed")
		ui.ErrorLine(err.Error(), "run `orbita logs -f --host "+o.host+"` to see build/runtime logs")
		return err
	}
	if dep.Migrated {
		ui.Step("Migrations applied (under advisory lock)")
	}
	for _, s := range dep.Services {
		ui.Field(s.Role, ui.Status(orZero(s.Status, "deployed"))+"  "+ui.URL(s.URL))
	}

	// 8. Success — print the live URL + dashboard links.
	printLiveLinks(dep)
	ui.Info("Auto-deploy is on: future `git push` to " + m.BranchOrDefault() + " redeploys via webhook.")
	return nil
}

func runPlan(ctx context.Context, client *orbita.Client, org string, proj *project.Project) error {
	ui.Header("Plan (dry run — nothing will be changed)")
	res, err := client.GritPlan(ctx, org, orbita.GritReconcileRequest{
		GritYAML: proj.ManifestYAML,
		GritJSON: proj.GritJSONText,
	})
	if err != nil {
		ui.ErrorLine(err.Error(), "")
		return err
	}
	ui.Field("App", ui.Value(res.GritApp))
	ui.Field("Mode", res.Mode)
	ui.Field("Migrate", fmt.Sprintf("%v", res.Migrate))
	if len(res.Addons) > 0 {
		ui.Field("Addons", strings.Join(res.Addons, ", "))
	}
	fmt.Println()
	for _, s := range res.Services {
		verb := "create"
		if !s.Created {
			verb = "update"
		}
		line := fmt.Sprintf("%-7s %s", verb, ui.Value(s.AppName))
		if s.Domain != "" {
			line += "  → " + ui.URL(s.Domain)
		}
		ui.Info(line)
	}
	fmt.Println()
	ui.Info("Run without --plan to apply.")
	return nil
}

func repoStep(ctx context.Context, proj *project.Project) error {
	token := LoadGitHubToken()
	if token == "" {
		return fmt.Errorf("no GitHub token stored")
	}
	owner, name, ok := proj.Manifest.RepoOwnerName()
	if !ok {
		return fmt.Errorf("repo must be owner/name")
	}
	ui.StepActive("Ensuring GitHub repo + pushing…")
	gh := github.New(token)
	if err := gh.EnsureAndPush(ctx, proj.Dir, owner, name, proj.Manifest.BranchOrDefault()); err != nil {
		return err
	}
	ui.Step("Pushed to " + ui.Value(owner+"/"+name))
	return nil
}

func printLiveLinks(dep *orbita.DeployResult) {
	ui.Success("Live")
	if dep.LiveURL != "" {
		ui.Field("App", ui.URL(dep.LiveURL))
	}
	if dep.APIURL != "" {
		ui.Field("API", ui.URL(dep.APIURL))
	}
	for _, k := range []string{"pulse", "sentinel", "studio"} {
		if u, ok := dep.DashboardURL[k]; ok {
			ui.Field(strings.Title(k), ui.URL(u))
		}
	}
}

func orZero(s, def string) string {
	if s == "" || s == "unknown" {
		return def
	}
	return s
}

// slugify turns an app name into an org/url-safe slug.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
