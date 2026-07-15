package grit

import (
	"fmt"
	"strings"
)

// ValidAddons is the closed set of addons Orbita provisions.
var ValidAddons = map[string]bool{
	"postgres": true,
	"redis":    true,
	"minio":    true,
}

// ValidationError aggregates all problems found in a manifest so the CLI can
// show them at once.
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	return "invalid grit.yaml:\n  - " + strings.Join(e.Problems, "\n  - ")
}

// ValidateManifest checks a parsed grit.yaml for structural correctness. It does
// not touch the filesystem — repo-relative checks (service paths) happen during
// detection against grit.json. Returns a *ValidationError with all problems.
func ValidateManifest(m *Manifest) error {
	var problems []string

	if strings.TrimSpace(m.App) == "" {
		problems = append(problems, "`app` is required")
	}
	if strings.TrimSpace(m.Repo) == "" {
		problems = append(problems, "`repo` is required")
	} else if _, _, ok := m.RepoOwnerName(); !ok {
		problems = append(problems, fmt.Sprintf("`repo` must be in owner/name form, got %q", m.Repo))
	}

	for _, a := range m.Addons {
		if !ValidAddons[a] {
			problems = append(problems, fmt.Sprintf("unknown addon %q (allowed: postgres, redis, minio)", a))
		}
	}

	for role, host := range map[string]string{
		"web":   m.Domains.Web,
		"admin": m.Domains.Admin,
		"api":   m.Domains.API,
		"docs":  m.Domains.Docs,
	} {
		if host == "" {
			continue
		}
		if err := validateHostname(host); err != nil {
			problems = append(problems, fmt.Sprintf("domains.%s: %v", role, err))
		}
	}

	if len(problems) > 0 {
		return &ValidationError{Problems: problems}
	}
	return nil
}

// validateHostname rejects schemes, paths, ports, and obvious junk — grit.yaml
// domains must be bare hostnames (grit-knowledge/08).
func validateHostname(host string) error {
	if strings.Contains(host, "://") {
		return fmt.Errorf("must be a bare hostname without a scheme (got %q)", host)
	}
	if strings.ContainsAny(host, "/ ") {
		return fmt.Errorf("must not contain a path or spaces (got %q)", host)
	}
	if strings.Contains(host, ":") {
		return fmt.Errorf("must not contain a port (got %q)", host)
	}
	if !strings.Contains(host, ".") {
		return fmt.Errorf("must be a fully-qualified domain (got %q)", host)
	}
	return nil
}

// ValidateForDeploy checks a manifest + detected grit.json together: the mode
// must be VPS-deployable, and any explicit service overrides must reference a
// role that exists in the derived service set.
func ValidateForDeploy(m *Manifest, g *GritJSON) error {
	if err := ValidateManifest(m); err != nil {
		return err
	}
	if !IsVPSDeployable(g.Architecture) {
		return &ValidationError{Problems: []string{
			fmt.Sprintf("architecture %q is not VPS-deployable — mobile apps ship to app stores", g.Architecture),
		}}
	}

	services, err := DeriveServices(g)
	if err != nil {
		return &ValidationError{Problems: []string{err.Error()}}
	}
	roles := map[string]bool{}
	for _, s := range services {
		roles[s.Role] = true
	}
	var problems []string
	for role := range m.Services {
		if !roles[role] {
			problems = append(problems, fmt.Sprintf("services.%s overrides a service that does not exist in a %q app", role, g.Architecture))
		}
	}
	if len(problems) > 0 {
		return &ValidationError{Problems: problems}
	}
	return nil
}
