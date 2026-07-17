// Package hosts manages ~/.orbita/hosts.yaml — the operator's registry mapping a
// friendly host name (e.g. "prod") to an Orbita API URL + orb_ deploy token.
// orbita init writes it; orbita deploy/logs/status/rollback read it.
package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Host is one registered Orbita control plane.
type Host struct {
	APIURL string `yaml:"api_url"`       // https://orbita.example.com
	Token  string `yaml:"token"`         // orb_...
	SSH    string `yaml:"ssh,omitempty"` // user@ip[:port] for the dashboard tunnel
	Org    string `yaml:"org,omitempty"` // default org slug for deploys
}

// File is the on-disk shape of ~/.orbita/hosts.yaml.
type File struct {
	Hosts map[string]Host `yaml:"hosts"`
}

// Path returns the hosts file path (honors ORBITA_HOSTS_FILE for tests).
func Path() (string, error) {
	if p := os.Getenv("ORBITA_HOSTS_FILE"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("hosts.Path: %w", err)
	}
	return filepath.Join(home, ".orbita", "hosts.yaml"), nil
}

// Load reads the hosts file, returning an empty registry if it doesn't exist.
func Load() (*File, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &File{Hosts: map[string]Host{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("hosts.Load: %w", err)
	}
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("hosts.Load: parse: %w", err)
	}
	if f.Hosts == nil {
		f.Hosts = map[string]Host{}
	}
	return &f, nil
}

// Save writes the hosts file with 0600 perms (it holds tokens), creating
// ~/.orbita if needed.
func (f *File) Save() error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return fmt.Errorf("hosts.Save: mkdir: %w", err)
	}
	data, err := yaml.Marshal(f)
	if err != nil {
		return fmt.Errorf("hosts.Save: marshal: %w", err)
	}
	if err := os.WriteFile(p, data, 0o600); err != nil {
		return fmt.Errorf("hosts.Save: write: %w", err)
	}
	return nil
}

// Get returns a host by name.
func (f *File) Get(name string) (Host, bool) {
	h, ok := f.Hosts[name]
	return h, ok
}

// Set adds or replaces a host and saves.
func Set(name string, h Host) error {
	f, err := Load()
	if err != nil {
		return err
	}
	f.Hosts[name] = h
	return f.Save()
}

// Names returns the registered host names, sorted.
func (f *File) Names() []string {
	names := make([]string, 0, len(f.Hosts))
	for n := range f.Hosts {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Resolve loads the registry and returns the named host, with a helpful error
// listing available hosts when the name is unknown.
func Resolve(name string) (Host, error) {
	f, err := Load()
	if err != nil {
		return Host{}, err
	}
	h, ok := f.Get(name)
	if !ok {
		if len(f.Hosts) == 0 {
			return Host{}, fmt.Errorf("no hosts registered — run `orbita init` first")
		}
		return Host{}, fmt.Errorf("unknown host %q — registered: %v", name, f.Names())
	}
	return h, nil
}
