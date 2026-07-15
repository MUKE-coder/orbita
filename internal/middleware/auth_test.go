package middleware

import "testing"

func TestApiKeyScopeAllows(t *testing.T) {
	tests := []struct {
		name   string
		scopes []string
		method string
		want   bool
	}{
		{"read scope allows GET", []string{"read"}, "GET", true},
		{"read scope allows HEAD", []string{"read"}, "HEAD", true},
		{"read scope blocks POST", []string{"read"}, "POST", false},
		{"read scope blocks DELETE", []string{"read"}, "DELETE", false},
		{"deploy scope allows GET", []string{"deploy"}, "GET", true},
		{"deploy scope allows POST", []string{"deploy"}, "POST", true},
		{"admin scope allows DELETE", []string{"admin"}, "DELETE", true},
		{"multiple scopes union", []string{"read", "deploy"}, "PUT", true},
		{"no scopes allows nothing", []string{}, "GET", false},
		{"unknown scope allows nothing", []string{"banana"}, "GET", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := apiKeyScopeAllows(tt.scopes, tt.method); got != tt.want {
				t.Errorf("apiKeyScopeAllows(%v, %s) = %v, want %v", tt.scopes, tt.method, got, tt.want)
			}
		})
	}
}
