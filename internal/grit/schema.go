// Package grit implements Orbita's Grit-awareness layer: parsing the grit.yaml
// deploy manifest, detecting a Grit app from grit.json, deriving the service
// map/build recipe from the architecture mode, and validating both. It is the
// ground truth for how Orbita builds, routes, and migrates a Grit app.
//
// The guiding principle (from grit-knowledge/08): the user writes grit.yaml
// once for the things only they know (repo, branch, domains, addons, env
// source, migrate toggle); everything about *what services exist and how to
// build them* is derived from grit.json + the repository, which Grit itself
// generates and keeps correct.
package grit

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest is the parsed grit.yaml deploy contract (project-description.md §5,
// reconciled with reality in grit-knowledge/08).
type Manifest struct {
	App    string `yaml:"app"`
	Repo   string `yaml:"repo"`   // GitHub "owner/name"
	Branch string `yaml:"branch"` // default "main"

	// Optional explicit service overrides. Usually omitted — derived from
	// grit.json instead. When present, paths are validated against the repo.
	Services map[string]ServiceOverride `yaml:"services,omitempty"`

	Addons  []string          `yaml:"addons,omitempty"` // subset of {postgres, redis, minio}
	Domains Domains           `yaml:"domains,omitempty"`
	Migrate *bool             `yaml:"migrate,omitempty"` // default true

	// Dashboard toggles (grit-knowledge/07). Defaults: observability/security
	// on, studio off.
	Observability *bool `yaml:"observability,omitempty"`
	Security      *bool `yaml:"security,omitempty"`
	Studio        *bool `yaml:"studio,omitempty"`

	Env EnvSource `yaml:"env,omitempty"`
}

// ServiceOverride is an optional per-service path/port override in grit.yaml.
type ServiceOverride struct {
	Path string `yaml:"path"`
	Port int    `yaml:"port,omitempty"`
}

// Domains maps Grit service roles to their public hostnames.
type Domains struct {
	Web   string `yaml:"web,omitempty"`
	Admin string `yaml:"admin,omitempty"`
	API   string `yaml:"api,omitempty"`
	Docs  string `yaml:"docs,omitempty"`
}

// EnvSource points at a local file whose values Orbita encrypts and injects.
type EnvSource struct {
	From string `yaml:"from,omitempty"` // e.g. ".env.production"
}

// Defaults applied when a field is omitted.
const (
	DefaultBranch = "main"
)

// MigrateEnabled reports whether migrations should run (default true).
func (m *Manifest) MigrateEnabled() bool {
	return m.Migrate == nil || *m.Migrate
}

// ObservabilityEnabled reports whether Pulse should be enabled (default true).
func (m *Manifest) ObservabilityEnabled() bool {
	return m.Observability == nil || *m.Observability
}

// SecurityEnabled reports whether Sentinel should be enabled (default true).
func (m *Manifest) SecurityEnabled() bool {
	return m.Security == nil || *m.Security
}

// StudioEnabled reports whether GORM Studio should be enabled (default false —
// it can edit production data).
func (m *Manifest) StudioEnabled() bool {
	return m.Studio != nil && *m.Studio
}

// BranchOrDefault returns the branch, defaulting to "main".
func (m *Manifest) BranchOrDefault() string {
	if strings.TrimSpace(m.Branch) == "" {
		return DefaultBranch
	}
	return m.Branch
}

// ParseManifest parses grit.yaml bytes into a Manifest (no validation).
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("ParseManifest: %w", err)
	}
	return &m, nil
}

// RepoOwnerName splits "owner/name" from the repo field.
func (m *Manifest) RepoOwnerName() (owner, name string, ok bool) {
	parts := strings.SplitN(strings.TrimSpace(m.Repo), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}
