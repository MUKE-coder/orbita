package orbita

import (
	"context"
	"net/http"
)

// GritReconcileRequest mirrors the server's handler payload.
type GritReconcileRequest struct {
	GritYAML        string            `json:"grit_yaml"`
	GritJSON        string            `json:"grit_json"`
	EnvValues       map[string]string `json:"env_values,omitempty"`
	GitConnectionID string            `json:"git_connection_id,omitempty"`
}

// ServiceResult is one service in a plan/reconcile response.
type ServiceResult struct {
	Role    string `json:"role"`
	AppName string `json:"app_name"`
	AppID   string `json:"app_id"`
	Domain  string `json:"domain"`
	Created bool   `json:"created"`
}

// ReconcileResult is the plan/reconcile response.
type ReconcileResult struct {
	GritApp       string            `json:"grit_app"`
	Mode          string            `json:"mode"`
	ProjectID     string            `json:"project_id"`
	EnvironmentID string            `json:"environment_id"`
	Services      []ServiceResult   `json:"services"`
	Addons        []string          `json:"addons"`
	Migrate       bool              `json:"migrate"`
	DashboardURLs map[string]string `json:"dashboard_urls"`
}

// ServiceLink is one service's status/URL in a deploy response.
type ServiceLink struct {
	Role   string `json:"role"`
	AppID  string `json:"app_id"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

// DeployResult is the deploy/status/rollback response.
type DeployResult struct {
	GritApp      string            `json:"grit_app"`
	Migrated     bool              `json:"migrated"`
	Services     []ServiceLink     `json:"services"`
	LiveURL      string            `json:"live_url"`
	APIURL       string            `json:"api_url"`
	DashboardURL map[string]string `json:"dashboard_urls"`
}

// GritPlan calls POST /grit/plan (dry run — no mutation).
func (c *Client) GritPlan(ctx context.Context, org string, req GritReconcileRequest) (*ReconcileResult, error) {
	var out ReconcileResult
	err := c.do(ctx, http.MethodPost, "/api/v1/orgs/"+org+"/grit/plan", req, &out)
	return &out, err
}

// GritReconcile calls POST /grit/reconcile (idempotent create-or-update).
func (c *Client) GritReconcile(ctx context.Context, org string, req GritReconcileRequest) (*ReconcileResult, error) {
	var out ReconcileResult
	err := c.do(ctx, http.MethodPost, "/api/v1/orgs/"+org+"/grit/reconcile", req, &out)
	return &out, err
}

// GritDeploy calls POST /grit/deploy (build → migrate → cut over).
func (c *Client) GritDeploy(ctx context.Context, org, gritApp, envID string) (*DeployResult, error) {
	var out DeployResult
	body := map[string]string{"grit_app": gritApp}
	if envID != "" {
		body["environment_id"] = envID
	}
	err := c.do(ctx, http.MethodPost, "/api/v1/orgs/"+org+"/grit/deploy", body, &out)
	return &out, err
}

// GritStatus calls GET /grit/:app/status.
func (c *Client) GritStatus(ctx context.Context, org, gritApp string) (*DeployResult, error) {
	var out DeployResult
	err := c.do(ctx, http.MethodGet, "/api/v1/orgs/"+org+"/grit/"+gritApp+"/status", nil, &out)
	return &out, err
}

// GritRollback calls POST /grit/:app/rollback.
func (c *Client) GritRollback(ctx context.Context, org, gritApp string) (*DeployResult, error) {
	var out DeployResult
	err := c.do(ctx, http.MethodPost, "/api/v1/orgs/"+org+"/grit/"+gritApp+"/rollback", nil, &out)
	return &out, err
}

// EnsureOrg creates the org if it doesn't exist (idempotent). Returns the slug.
func (c *Client) EnsureOrg(ctx context.Context, name, slug string) (string, error) {
	// List existing orgs; create if the slug isn't present.
	var orgs []struct {
		Slug string `json:"slug"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/v1/orgs", nil, &orgs); err == nil {
		for _, o := range orgs {
			if o.Slug == slug {
				return slug, nil
			}
		}
	}
	var created struct {
		Slug string `json:"slug"`
	}
	err := c.do(ctx, http.MethodPost, "/api/v1/orgs", map[string]string{"name": name, "slug": slug}, &created)
	if err != nil {
		return "", err
	}
	if created.Slug != "" {
		return created.Slug, nil
	}
	return slug, nil
}
