package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/middleware"
	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

type ServiceHandler struct {
	templateService *service.TemplateService
}

func NewServiceHandler(templateService *service.TemplateService) *ServiceHandler {
	return &ServiceHandler{templateService: templateService}
}

type DeployServiceRequest struct {
	TemplateID    string            `json:"template_id" binding:"required"`
	Name          string            `json:"name" binding:"required,min=2"`
	EnvironmentID string            `json:"environment_id" binding:"required"`
	Params        map[string]string `json:"params" binding:"required"`
}

// List templates (public within authenticated context)
func (h *ServiceHandler) ListTemplates(c *gin.Context) {
	templates, err := h.templateService.ListTemplates(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to list templates")
		return
	}
	response.Success(c, http.StatusOK, templates)
}

// Get template detail
func (h *ServiceHandler) GetTemplate(c *gin.Context) {
	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		response.BadRequest(c, "Invalid template ID")
		return
	}

	tmpl, err := h.templateService.GetTemplate(c.Request.Context(), templateID)
	if err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			response.NotFound(c, "Template not found")
			return
		}
		response.InternalError(c, "Failed to get template")
		return
	}
	response.Success(c, http.StatusOK, tmpl)
}

// List deployed services
func (h *ServiceHandler) ListServices(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	services, err := h.templateService.ListServices(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, "Failed to list services")
		return
	}
	response.Success(c, http.StatusOK, services)
}

// Deploy service from template
func (h *ServiceHandler) DeployService(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	org := middleware.GetOrgFromContext(c)

	var req DeployServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		response.BadRequest(c, "Invalid template ID")
		return
	}

	envID, err := uuid.Parse(req.EnvironmentID)
	if err != nil {
		response.BadRequest(c, "Invalid environment ID")
		return
	}

	svc, err := h.templateService.DeployService(c.Request.Context(), orgID, org.Slug, service.DeployServiceInput{
		TemplateID:    templateID,
		Name:          req.Name,
		EnvironmentID: envID,
		Params:        req.Params,
	})
	if err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			response.NotFound(c, "Template not found")
			return
		}
		if svc != nil {
			response.Success(c, http.StatusCreated, gin.H{"service": svc, "warning": err.Error()})
			return
		}
		response.InternalError(c, "Failed to deploy service")
		return
	}
	response.Success(c, http.StatusCreated, svc)
}

// Get deployed service detail
func (h *ServiceHandler) GetService(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	serviceID, err := uuid.Parse(c.Param("serviceId"))
	if err != nil {
		response.BadRequest(c, "Invalid service ID")
		return
	}

	svc, err := h.templateService.GetService(c.Request.Context(), serviceID, orgID)
	if err != nil {
		if errors.Is(err, service.ErrServiceNotFound) {
			response.NotFound(c, "Service not found")
			return
		}
		response.InternalError(c, "Failed to get service")
		return
	}
	response.Success(c, http.StatusOK, svc)
}

// Delete deployed service
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	serviceID, err := uuid.Parse(c.Param("serviceId"))
	if err != nil {
		response.BadRequest(c, "Invalid service ID")
		return
	}

	if err := h.templateService.DeleteService(c.Request.Context(), serviceID, orgID); err != nil {
		if errors.Is(err, service.ErrServiceNotFound) {
			response.NotFound(c, "Service not found")
			return
		}
		response.InternalError(c, "Failed to delete service")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Service deleted"})
}
