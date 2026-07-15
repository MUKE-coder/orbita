// Package project reads a Grit project from the working directory: the grit.yaml
// deploy manifest, the grit.json architecture marker, and the env.from file. It
// also generates a grit.yaml interactively when one is absent (the first-run
// wizard). It reuses Orbita's internal/grit for parsing/detection/validation.
package project

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/orbita-sh/orbita/internal/grit"
)

// Project is a loaded Grit project ready to deploy.
type Project struct {
	Dir          string
	Manifest     *grit.Manifest
	ManifestYAML string // raw grit.yaml text (submitted to Orbita)
	GritJSON     *grit.GritJSON
	GritJSONText string // raw grit.json text
	EnvValues    map[string]string
}

// Load reads grit.yaml + grit.json (+ env.from) from dir. Returns
// ErrNoManifest if grit.yaml is absent (caller may run the wizard).
func Load(dir string) (*Project, error) {
	manifestPath := filepath.Join(dir, "grit.yaml")
	rawYAML, err := os.ReadFile(manifestPath)
	if os.IsNotExist(err) {
		return nil, ErrNoManifest
	}
	if err != nil {
		return nil, fmt.Errorf("read grit.yaml: %w", err)
	}
	m, err := grit.ParseManifest(rawYAML)
	if err != nil {
		return nil, err
	}

	gjPath := filepath.Join(dir, "grit.json")
	rawJSON, err := os.ReadFile(gjPath)
	if err != nil {
		return nil, fmt.Errorf("read grit.json (is this a Grit app?): %w", err)
	}
	gj, err := grit.ParseGritJSON(rawJSON)
	if err != nil {
		return nil, err
	}

	if err := grit.ValidateForDeploy(m, gj); err != nil {
		return nil, err
	}

	p := &Project{
		Dir:          dir,
		Manifest:     m,
		ManifestYAML: string(rawYAML),
		GritJSON:     gj,
		GritJSONText: string(rawJSON),
		EnvValues:    map[string]string{},
	}

	// Read env.from (never logged).
	if m.Env.From != "" {
		envPath := filepath.Join(dir, m.Env.From)
		if vals, err := parseDotenv(envPath); err == nil {
			p.EnvValues = vals
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("read %s: %w", m.Env.From, err)
		}
	}
	return p, nil
}

// ErrNoManifest signals grit.yaml is absent.
var ErrNoManifest = fmt.Errorf("no grit.yaml found")

// DetectGritJSON reads grit.json from dir (used by the wizard to derive
// defaults). Returns nil if absent.
func DetectGritJSON(dir string) *grit.GritJSON {
	raw, err := os.ReadFile(filepath.Join(dir, "grit.json"))
	if err != nil {
		return nil
	}
	gj, err := grit.ParseGritJSON(raw)
	if err != nil {
		return nil
	}
	return gj
}

// WriteManifest writes a grit.yaml to dir.
func WriteManifest(dir string, m *grit.Manifest) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "grit.yaml"), data, 0o644)
}

// parseDotenv reads a .env file into a map (KEY=value; # comments; quotes trimmed).
func parseDotenv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		val = strings.Trim(val, `"'`)
		out[key] = val
	}
	return out, sc.Err()
}
