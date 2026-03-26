package traefik

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	configDir string
}

func NewManager(configDir string) *Manager {
	return &Manager{configDir: configDir}
}

type TraefikResource struct {
	ResourceID   uuid.UUID
	Domain       string
	ServiceName  string
	ServicePort  int
	SSLEnabled   bool
	HTTPSRedirect bool
}

type traefikConfig struct {
	HTTP *httpConfig `json:"http,omitempty"`
}

type httpConfig struct {
	Routers     map[string]router     `json:"routers,omitempty"`
	Services    map[string]svcConfig  `json:"services,omitempty"`
	Middlewares map[string]middleware `json:"middlewares,omitempty"`
}

type router struct {
	Rule        string   `json:"rule"`
	Service     string   `json:"service"`
	EntryPoints []string `json:"entryPoints"`
	TLS         *tlsConf `json:"tls,omitempty"`
	Middlewares []string `json:"middlewares,omitempty"`
}

type tlsConf struct {
	CertResolver string `json:"certResolver,omitempty"`
}

type svcConfig struct {
	LoadBalancer *loadBalancer `json:"loadBalancer,omitempty"`
}

type loadBalancer struct {
	Servers []server `json:"servers"`
}

type server struct {
	URL string `json:"url"`
}

type middleware struct {
	RedirectScheme *redirectScheme `json:"redirectScheme,omitempty"`
	Headers        *headers       `json:"headers,omitempty"`
}

type redirectScheme struct {
	Scheme    string `json:"scheme"`
	Permanent bool   `json:"permanent"`
}

type headers struct {
	SSLRedirect         bool   `json:"sslRedirect,omitempty"`
	STSSeconds          int    `json:"stsSeconds,omitempty"`
	STSIncludeSubdomains bool  `json:"stsIncludeSubdomains,omitempty"`
	ContentTypeNosniff  bool   `json:"contentTypeNosniff,omitempty"`
	FrameDeny           bool   `json:"frameDeny,omitempty"`
}

func (m *Manager) UpsertRoute(resource TraefikResource) error {
	dynamicDir := filepath.Join(m.configDir, "dynamic")
	if err := os.MkdirAll(dynamicDir, 0755); err != nil {
		return fmt.Errorf("UpsertRoute: create dir: %w", err)
	}

	routerName := fmt.Sprintf("orbita-%s", resource.ResourceID.String()[:8])
	serviceName := fmt.Sprintf("orbita-svc-%s", resource.ResourceID.String()[:8])

	cfg := traefikConfig{
		HTTP: &httpConfig{
			Routers: map[string]router{
				routerName: {
					Rule:        fmt.Sprintf("Host(`%s`)", resource.Domain),
					Service:     serviceName,
					EntryPoints: []string{"websecure"},
				},
			},
			Services: map[string]svcConfig{
				serviceName: {
					LoadBalancer: &loadBalancer{
						Servers: []server{
							{URL: fmt.Sprintf("http://%s:%d", resource.ServiceName, resource.ServicePort)},
						},
					},
				},
			},
		},
	}

	if resource.SSLEnabled {
		cfg.HTTP.Routers[routerName] = router{
			Rule:        fmt.Sprintf("Host(`%s`)", resource.Domain),
			Service:     serviceName,
			EntryPoints: []string{"websecure"},
			TLS:         &tlsConf{CertResolver: "letsencrypt"},
		}

		// Add HTTP to HTTPS redirect router
		redirectName := fmt.Sprintf("%s-redirect", routerName)
		cfg.HTTP.Routers[redirectName] = router{
			Rule:        fmt.Sprintf("Host(`%s`)", resource.Domain),
			Service:     serviceName,
			EntryPoints: []string{"web"},
			Middlewares: []string{fmt.Sprintf("%s-https", routerName)},
		}

		cfg.HTTP.Middlewares = map[string]middleware{
			fmt.Sprintf("%s-https", routerName): {
				RedirectScheme: &redirectScheme{
					Scheme:    "https",
					Permanent: true,
				},
			},
		}
	}

	configFile := filepath.Join(dynamicDir, fmt.Sprintf("%s.json", resource.ResourceID.String()))
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("UpsertRoute: marshal: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("UpsertRoute: write file: %w", err)
	}

	log.Info().
		Str("domain", resource.Domain).
		Str("resource_id", resource.ResourceID.String()).
		Bool("ssl", resource.SSLEnabled).
		Msg("Traefik route upserted")

	return nil
}

func (m *Manager) RemoveRoute(resourceID uuid.UUID) error {
	dynamicDir := filepath.Join(m.configDir, "dynamic")
	configFile := filepath.Join(dynamicDir, fmt.Sprintf("%s.json", resourceID.String()))

	if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("RemoveRoute: %w", err)
	}

	log.Info().Str("resource_id", resourceID.String()).Msg("Traefik route removed")
	return nil
}
