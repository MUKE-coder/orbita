package models

type HealthCheck struct {
	Enabled   bool   `json:"enabled"`
	Type      string `json:"type"` // http, tcp
	Path      string `json:"path,omitempty"`
	Port      int    `json:"port"`
	Interval  int    `json:"interval"`  // seconds
	Timeout   int    `json:"timeout"`   // seconds
	Retries   int    `json:"retries"`
	Status    int    `json:"status,omitempty"` // expected HTTP status
}

const (
	HealthCheckTypeHTTP = "http"
	HealthCheckTypeTCP  = "tcp"
)
