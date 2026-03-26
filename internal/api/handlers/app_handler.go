package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/middleware"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

type AppHandler struct {
	appService *service.AppService
}

func NewAppHandler(appService *service.AppService) *AppHandler {
	return &AppHandler{appService: appService}
}

type CreateAppRequest struct {
	Name          string `json:"name" binding:"required,min=2"`
	EnvironmentID string `json:"environment_id" binding:"required"`
	SourceType    string `json:"source_type" binding:"required,oneof=docker-image"`
	Image         string `json:"image" binding:"required"`
	Port          *int   `json:"port"`
	Replicas      int    `json:"replicas"`
}

type UpdateAppRequest struct {
	Name     string  `json:"name"`
	Port     *int    `json:"port"`
	Replicas *int    `json:"replicas"`
}

func (h *AppHandler) ListApps(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	apps, err := h.appService.ListApps(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, "Failed to list apps")
		return
	}
	response.Success(c, http.StatusOK, apps)
}

func (h *AppHandler) CreateApp(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	var req CreateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	envID, err := uuid.Parse(req.EnvironmentID)
	if err != nil {
		response.BadRequest(c, "Invalid environment ID")
		return
	}

	app, err := h.appService.CreateApp(c.Request.Context(), orgID, service.CreateAppInput{
		Name:          req.Name,
		EnvironmentID: envID,
		SourceType:    req.SourceType,
		Image:         req.Image,
		Port:          req.Port,
		Replicas:      req.Replicas,
	})
	if err != nil {
		response.InternalError(c, "Failed to create app")
		return
	}

	response.Success(c, http.StatusCreated, app)
}

func (h *AppHandler) GetApp(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		response.BadRequest(c, "Invalid app ID")
		return
	}

	app, err := h.appService.GetApp(c.Request.Context(), appID, orgID)
	if err != nil {
		if errors.Is(err, service.ErrAppNotFound) {
			response.NotFound(c, "App not found")
			return
		}
		response.InternalError(c, "Failed to get app")
		return
	}
	response.Success(c, http.StatusOK, app)
}

func (h *AppHandler) UpdateApp(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		response.BadRequest(c, "Invalid app ID")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	app, err := h.appService.UpdateApp(c.Request.Context(), appID, orgID, updates)
	if err != nil {
		if errors.Is(err, service.ErrAppNotFound) {
			response.NotFound(c, "App not found")
			return
		}
		response.InternalError(c, "Failed to update app")
		return
	}
	response.Success(c, http.StatusOK, app)
}

func (h *AppHandler) DeleteApp(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	org := middleware.GetOrgFromContext(c)
	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		response.BadRequest(c, "Invalid app ID")
		return
	}

	if err := h.appService.DeleteApp(c.Request.Context(), appID, orgID, org.Slug); err != nil {
		if errors.Is(err, service.ErrAppNotFound) {
			response.NotFound(c, "App not found")
			return
		}
		response.InternalError(c, "Failed to delete app")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "App deleted"})
}

func (h *AppHandler) Deploy(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	org := middleware.GetOrgFromContext(c)
	userID := middleware.GetUserIDFromContext(c)
	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		response.BadRequest(c, "Invalid app ID")
		return
	}

	deployment, err := h.appService.Deploy(c.Request.Context(), appID, orgID, org.Slug, &userID)
	if err != nil {
		if errors.Is(err, service.ErrAppNotFound) {
			response.NotFound(c, "App not found")
			return
		}
		// Deploy may partially succeed — return the deployment with error status
		if deployment != nil {
			response.Success(c, http.StatusOK, gin.H{
				"deployment": deployment,
				"error":      err.Error(),
			})
			return
		}
		response.InternalError(c, "Failed to deploy")
		return
	}

	response.Success(c, http.StatusOK, deployment)
}

func (h *AppHandler) Rollback(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	org := middleware.GetOrgFromContext(c)
	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		response.BadRequest(c, "Invalid app ID")
		return
	}
	deploymentID, err := uuid.Parse(c.Param("deploymentId"))
	if err != nil {
		response.BadRequest(c, "Invalid deployment ID")
		return
	}

	deployment, err := h.appService.Rollback(c.Request.Context(), appID, deploymentID, orgID, org.Slug)
	if err != nil {
		if errors.Is(err, service.ErrAppNotFound) || errors.Is(err, service.ErrDeploymentNotFound) {
			response.NotFound(c, "Not found")
			return
		}
		response.InternalError(c, "Failed to rollback")
		return
	}
	response.Success(c, http.StatusOK, deployment)
}

func (h *AppHandler) Stop(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, _ := uuid.Parse(c.Param("appId"))
	if err := h.appService.Stop(c.Request.Context(), appID, orgID); err != nil {
		response.InternalError(c, "Failed to stop app")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "App stopped"})
}

func (h *AppHandler) Start(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, _ := uuid.Parse(c.Param("appId"))
	if err := h.appService.Start(c.Request.Context(), appID, orgID); err != nil {
		response.InternalError(c, "Failed to start app")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "App started"})
}

func (h *AppHandler) Restart(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, _ := uuid.Parse(c.Param("appId"))
	if err := h.appService.Restart(c.Request.Context(), appID, orgID); err != nil {
		response.InternalError(c, "Failed to restart app")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "App restarted"})
}

func (h *AppHandler) ListDeployments(c *gin.Context) {
	appID, _ := uuid.Parse(c.Param("appId"))
	deployments, err := h.appService.ListDeployments(c.Request.Context(), appID, 20)
	if err != nil {
		response.InternalError(c, "Failed to list deployments")
		return
	}
	response.Success(c, http.StatusOK, deployments)
}

func (h *AppHandler) GetStatus(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, _ := uuid.Parse(c.Param("appId"))
	status, err := h.appService.GetStatus(c.Request.Context(), appID, orgID)
	if err != nil {
		response.InternalError(c, "Failed to get status")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"status": status})
}

func (h *AppHandler) GetLogs(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, _ := uuid.Parse(c.Param("appId"))
	logs, err := h.appService.GetLogs(c.Request.Context(), appID, orgID, 100)
	if err != nil {
		response.InternalError(c, "Failed to get logs")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"logs": logs})
}

func (h *AppHandler) GetMetrics(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, _ := uuid.Parse(c.Param("appId"))
	metrics, err := h.appService.GetMetrics(c.Request.Context(), appID, orgID)
	if err != nil {
		if errors.Is(err, service.ErrAppNotFound) {
			response.NotFound(c, "App not found")
			return
		}
		response.InternalError(c, "Failed to get metrics")
		return
	}
	response.Success(c, http.StatusOK, metrics)
}

// Suppress unused import
var _ = models.AppStatusCreated
