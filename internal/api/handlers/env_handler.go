package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/middleware"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

type EnvHandler struct {
	envService *service.EnvService
}

func NewEnvHandler(envService *service.EnvService) *EnvHandler {
	return &EnvHandler{envService: envService}
}

type SetEnvVarRequest struct {
	Key      string `json:"key" binding:"required"`
	Value    string `json:"value" binding:"required"`
	IsSecret bool   `json:"is_secret"`
}

type BulkSetEnvVarsRequest struct {
	Variables []SetEnvVarRequest `json:"variables" binding:"required"`
}

type ImportDotenvRequest struct {
	Content string `json:"content" binding:"required"`
}

// List env vars for an app
func (h *EnvHandler) ListAppEnvVars(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		response.BadRequest(c, "Invalid app ID")
		return
	}

	vars, err := h.envService.GetEnvVars(c.Request.Context(), appID, models.ResourceTypeApp, orgID)
	if err != nil {
		response.InternalError(c, "Failed to list env vars")
		return
	}
	response.Success(c, http.StatusOK, vars)
}

// Set a single env var
func (h *EnvHandler) SetAppEnvVar(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		response.BadRequest(c, "Invalid app ID")
		return
	}

	var req SetEnvVarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.envService.SetEnvVar(c.Request.Context(), appID, models.ResourceTypeApp, req.Key, req.Value, req.IsSecret, orgID); err != nil {
		response.InternalError(c, "Failed to set env var")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Environment variable set"})
}

// Bulk set env vars
func (h *EnvHandler) BulkSetAppEnvVars(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		response.BadRequest(c, "Invalid app ID")
		return
	}

	var req BulkSetEnvVarsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	for _, v := range req.Variables {
		if err := h.envService.SetEnvVar(c.Request.Context(), appID, models.ResourceTypeApp, v.Key, v.Value, v.IsSecret, orgID); err != nil {
			response.InternalError(c, "Failed to set env vars")
			return
		}
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Environment variables updated"})
}

// Delete an env var
func (h *EnvHandler) DeleteAppEnvVar(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	envVarID, err := uuid.Parse(c.Param("envId"))
	if err != nil {
		response.BadRequest(c, "Invalid env var ID")
		return
	}

	if err := h.envService.DeleteEnvVar(c.Request.Context(), envVarID, orgID); err != nil {
		response.InternalError(c, "Failed to delete env var")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Environment variable deleted"})
}

// Import from .env content
func (h *EnvHandler) ImportDotenv(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		response.BadRequest(c, "Invalid app ID")
		return
	}

	var req ImportDotenvRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	parsed, err := h.envService.ImportFromDotenv(req.Content)
	if err != nil {
		response.BadRequest(c, "Failed to parse .env content")
		return
	}

	if err := h.envService.BulkSetEnvVars(c.Request.Context(), appID, models.ResourceTypeApp, parsed, orgID); err != nil {
		response.InternalError(c, "Failed to import env vars")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message":  "Imported successfully",
		"imported": len(parsed),
	})
}
