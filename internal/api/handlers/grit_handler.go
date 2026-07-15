package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/grit"
	"github.com/orbita-sh/orbita/internal/middleware"
	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

// GritHandler exposes the Grit-aware endpoints the `grit deploy` CLI calls.
type GritHandler struct {
	gritService *service.GritService
}

func NewGritHandler(gritService *service.GritService) *GritHandler {
	return &GritHandler{gritService: gritService}
}

// GritReconcileRequest is the CLI payload: the grit.yaml text, the detected
// grit.json text, the env-file values, and the git connection (optional).
type GritReconcileRequest struct {
	GritYAML        string            `json:"grit_yaml" binding:"required"`
	GritJSON        string            `json:"grit_json" binding:"required"`
	EnvValues       map[string]string `json:"env_values"`
	GitConnectionID string            `json:"git_connection_id"`
	Plan            bool              `json:"plan"` // dry-run: validate + derive, don't mutate
}

// parseReconcile parses and validates the request into a ReconcileInput.
func (h *GritHandler) parseReconcile(c *gin.Context) (*service.ReconcileInput, bool) {
	orgID := middleware.GetOrgIDFromContext(c)
	org := middleware.GetOrgFromContext(c)

	var req GritReconcileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return nil, false
	}

	m, err := grit.ParseManifest([]byte(req.GritYAML))
	if err != nil {
		response.BadRequest(c, "invalid grit.yaml: "+err.Error())
		return nil, false
	}
	gj, err := grit.ParseGritJSON([]byte(req.GritJSON))
	if err != nil {
		response.BadRequest(c, "invalid grit.json: "+err.Error())
		return nil, false
	}
	if err := grit.ValidateForDeploy(m, gj); err != nil {
		response.BadRequest(c, err.Error())
		return nil, false
	}

	in := &service.ReconcileInput{
		OrgID:     orgID,
		OrgSlug:   org.Slug,
		Manifest:  m,
		GritJSON:  gj,
		EnvValues: req.EnvValues,
	}
	if req.GitConnectionID != "" {
		if id, err := uuid.Parse(req.GitConnectionID); err == nil {
			in.GitConnID = &id
		}
	}
	return in, true
}

// Reconcile creates-or-updates all Orbita resources for a Grit app (idempotent).
// POST /orgs/:orgSlug/grit/reconcile
func (h *GritHandler) Reconcile(c *gin.Context) {
	in, ok := h.parseReconcile(c)
	if !ok {
		return
	}
	res, err := h.gritService.Reconcile(c.Request.Context(), *in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, http.StatusOK, res)
}

// Plan validates + derives the full plan without mutating anything (dry run).
// POST /orgs/:orgSlug/grit/plan
func (h *GritHandler) Plan(c *gin.Context) {
	in, ok := h.parseReconcile(c)
	if !ok {
		return
	}
	res, err := h.gritService.PlanReconcile(c.Request.Context(), *in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, http.StatusOK, res)
}

// GritDeployRequest triggers a deploy of an already-reconciled Grit app.
type GritDeployRequest struct {
	GritApp       string `json:"grit_app" binding:"required"`
	EnvironmentID string `json:"environment_id"`
}

// Deploy builds+deploys every service, runs migrations, and cuts over.
// POST /orgs/:orgSlug/grit/deploy
func (h *GritHandler) Deploy(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	org := middleware.GetOrgFromContext(c)
	userID := middleware.GetUserIDFromContext(c)

	var req GritDeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	envID, err := h.gritService.ResolveEnvID(c.Request.Context(), orgID, req.GritApp, req.EnvironmentID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.gritService.Deploy(c.Request.Context(), orgID, org.Slug, req.GritApp, envID, &userID)
	if err != nil {
		response.InternalError(c, "Grit deploy failed: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, res)
}

// Status returns the current status + links for a Grit app's services.
// GET /orgs/:orgSlug/grit/:gritApp/status
func (h *GritHandler) Status(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	gritApp := c.Param("gritApp")

	res, err := h.gritService.Status(c.Request.Context(), orgID, gritApp, c.Query("environment_id"))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, http.StatusOK, res)
}

// Rollback reverts every service of a Grit app to its previous deploy.
// POST /orgs/:orgSlug/grit/:gritApp/rollback
func (h *GritHandler) Rollback(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	org := middleware.GetOrgFromContext(c)
	userID := middleware.GetUserIDFromContext(c)
	gritApp := c.Param("gritApp")

	res, err := h.gritService.Rollback(c.Request.Context(), orgID, org.Slug, gritApp, c.Query("environment_id"), &userID)
	if err != nil {
		response.InternalError(c, "Grit rollback failed: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, res)
}
