// Package orbita is a thin HTTP client for the Orbita control-plane API used by
// the grit CLI: health, register/login (first-run super admin bootstrap),
// orb_ API-key creation, platform metrics, and the Grit deploy endpoints.
package orbita

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to one Orbita instance. Token may be a JWT (from login) or an
// orb_ API key; both authenticate the same routes.
type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

// New creates a client for the given Orbita base URL (e.g. https://orbita.example.com).
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTP:    &http.Client{Timeout: 15 * time.Minute}, // deploys are synchronous
	}
}

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// do performs a request and decodes the `data` envelope into out (if non-nil).
func (c *Client) do(ctx context.Context, method, path string, body, out interface{}) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var ae apiError
		if json.Unmarshal(data, &ae) == nil && ae.Error.Message != "" {
			return fmt.Errorf("%s (%s)", ae.Error.Message, ae.Error.Code)
		}
		return fmt.Errorf("%s %s: HTTP %d: %s", method, path, resp.StatusCode, truncate(string(data), 300))
	}

	if out != nil {
		var env struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return json.Unmarshal(data, out)
	}
	return nil
}

// Health hits GET /health (no auth). Returns the raw status map.
func (c *Client) Health(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/health", nil)
	if err != nil {
		return nil, err
	}
	hc := &http.Client{Timeout: 10 * time.Second}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out map[string]interface{}
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("health: HTTP %d", resp.StatusCode)
	}
	_ = json.Unmarshal(data, &out)
	return out, nil
}

// WaitHealthy polls /health until it responds ok or the timeout elapses.
func (c *Client) WaitHealthy(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if _, err := c.Health(ctx); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("Orbita did not become healthy within %s", timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

// RegisterResult is returned by register/login.
type RegisterResult struct {
	AccessToken string `json:"access_token"`
}

// Register creates the first user (who becomes super admin on a fresh Orbita)
// and returns an access token.
func (c *Client) Register(ctx context.Context, email, password, name string) (string, error) {
	var out RegisterResult
	err := c.do(ctx, http.MethodPost, "/api/v1/auth/register",
		map[string]string{"email": email, "password": password, "name": name}, &out)
	if err != nil {
		return "", err
	}
	return out.AccessToken, nil
}

// Login authenticates and returns an access token.
func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
	var out RegisterResult
	err := c.do(ctx, http.MethodPost, "/api/v1/auth/login",
		map[string]string{"email": email, "password": password}, &out)
	if err != nil {
		return "", err
	}
	return out.AccessToken, nil
}

// APIKeyResult is returned when creating an orb_ key (the full key is shown once).
type APIKeyResult struct {
	Key string `json:"key"`
}

// CreateAPIKey creates an orb_ deploy key with the given scopes. Requires the
// client to be authenticated with a JWT (from Register/Login).
func (c *Client) CreateAPIKey(ctx context.Context, name string, scopes []string) (string, error) {
	// The API returns the raw key in one of a few field names depending on
	// version; capture all and pick the orb_-prefixed one.
	var raw map[string]interface{}
	err := c.do(ctx, http.MethodPost, "/api/v1/me/api-keys",
		map[string]interface{}{"name": name, "scopes": scopes}, &raw)
	if err != nil {
		return "", err
	}
	for _, f := range []string{"key", "api_key", "raw_key", "token"} {
		if v, ok := raw[f].(string); ok && strings.HasPrefix(v, "orb_") {
			return v, nil
		}
	}
	// Fall back to any orb_ string value.
	for _, v := range raw {
		if s, ok := v.(string); ok && strings.HasPrefix(s, "orb_") {
			return s, nil
		}
	}
	return "", fmt.Errorf("API key created but the key value was not returned")
}

// PlatformMetrics returns super-admin platform metrics (best-effort shape).
func (c *Client) PlatformMetrics(ctx context.Context) (map[string]interface{}, error) {
	var out map[string]interface{}
	err := c.do(ctx, http.MethodGet, "/api/v1/admin/platform/metrics", nil, &out)
	return out, err
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
