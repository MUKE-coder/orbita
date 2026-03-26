package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

type AdminHandler struct {
	orgService *service.OrgService
}

func NewAdminHandler(orgService *service.OrgService) *AdminHandler {
	return &AdminHandler{orgService: orgService}
}

type CreatePlanRequest struct {
	Name         string `json:"name" binding:"required"`
	MaxCPUCores  int    `json:"max_cpu_cores" binding:"required,min=1"`
	MaxRAMMB     int    `json:"max_ram_mb" binding:"required,min=128"`
	MaxDiskGB    int    `json:"max_disk_gb" binding:"required,min=1"`
	MaxApps      int    `json:"max_apps" binding:"required,min=1"`
	MaxDatabases int    `json:"max_databases" binding:"required,min=0"`
}

type UpdatePlanRequest struct {
	Name         *string `json:"name"`
	MaxCPUCores  *int    `json:"max_cpu_cores"`
	MaxRAMMB     *int    `json:"max_ram_mb"`
	MaxDiskGB    *int    `json:"max_disk_gb"`
	MaxApps      *int    `json:"max_apps"`
	MaxDatabases *int    `json:"max_databases"`
}

type AssignPlanRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
}

// List all plans
func (h *AdminHandler) ListPlans(c *gin.Context) {
	plans, err := h.orgService.ListPlans(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to list plans")
		return
	}
	response.Success(c, http.StatusOK, plans)
}

// Create plan
func (h *AdminHandler) CreatePlan(c *gin.Context) {
	var req CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	plan := &models.ResourcePlan{
		Name:         req.Name,
		MaxCPUCores:  req.MaxCPUCores,
		MaxRAMMB:     req.MaxRAMMB,
		MaxDiskGB:    req.MaxDiskGB,
		MaxApps:      req.MaxApps,
		MaxDatabases: req.MaxDatabases,
	}

	if err := h.orgService.CreatePlan(c.Request.Context(), plan); err != nil {
		response.InternalError(c, "Failed to create plan")
		return
	}

	response.Success(c, http.StatusCreated, plan)
}

// Update plan
func (h *AdminHandler) UpdatePlan(c *gin.Context) {
	planID, err := uuid.Parse(c.Param("planId"))
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	plans, _ := h.orgService.ListPlans(c.Request.Context())
	var plan *models.ResourcePlan
	for i := range plans {
		if plans[i].ID == planID {
			plan = &plans[i]
			break
		}
	}
	if plan == nil {
		response.NotFound(c, "Plan not found")
		return
	}

	if req.Name != nil {
		plan.Name = *req.Name
	}
	if req.MaxCPUCores != nil {
		plan.MaxCPUCores = *req.MaxCPUCores
	}
	if req.MaxRAMMB != nil {
		plan.MaxRAMMB = *req.MaxRAMMB
	}
	if req.MaxDiskGB != nil {
		plan.MaxDiskGB = *req.MaxDiskGB
	}
	if req.MaxApps != nil {
		plan.MaxApps = *req.MaxApps
	}
	if req.MaxDatabases != nil {
		plan.MaxDatabases = *req.MaxDatabases
	}

	if err := h.orgService.UpdatePlan(c.Request.Context(), plan); err != nil {
		response.InternalError(c, "Failed to update plan")
		return
	}

	response.Success(c, http.StatusOK, plan)
}

// Delete plan
func (h *AdminHandler) DeletePlan(c *gin.Context) {
	planID, err := uuid.Parse(c.Param("planId"))
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	if err := h.orgService.DeletePlan(c.Request.Context(), planID); err != nil {
		response.InternalError(c, "Failed to delete plan")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Plan deleted"})
}

// List all orgs (super admin)
func (h *AdminHandler) ListAllOrgs(c *gin.Context) {
	orgs, err := h.orgService.ListAllOrgs(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to list organizations")
		return
	}
	response.Success(c, http.StatusOK, orgs)
}

// Assign plan to org
func (h *AdminHandler) AssignPlanToOrg(c *gin.Context) {
	orgSlug := c.Param("orgSlug")

	var req AssignPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	if err := h.orgService.AssignPlanToOrg(c.Request.Context(), orgSlug, planID); err != nil {
		if errors.Is(err, service.ErrOrgNotFound) {
			response.NotFound(c, "Organization not found")
			return
		}
		if errors.Is(err, service.ErrPlanNotFound) {
			response.NotFound(c, "Plan not found")
			return
		}
		response.InternalError(c, "Failed to assign plan")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Plan assigned"})
}
