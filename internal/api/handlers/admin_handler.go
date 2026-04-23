package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/docker"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

type AdminHandler struct {
	orgService   *service.OrgService
	dockerClient *docker.Client
}

func NewAdminHandler(orgService *service.OrgService, dockerClient *docker.Client) *AdminHandler {
	return &AdminHandler{orgService: orgService, dockerClient: dockerClient}
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

// ---------- Host capacity ----------

// Capacity is the host's resource capacity + what's currently allocated across
// all orgs, so super admin can make informed sizing decisions.
type Capacity struct {
	Host      ResourcePool `json:"host"`
	Allocated ResourcePool `json:"allocated"`
	Available ResourcePool `json:"available"`
	Orgs      []OrgAlloc   `json:"orgs"`
}

type ResourcePool struct {
	CPUCores int `json:"cpu_cores"`
	RAMMB    int `json:"ram_mb"`
	DiskGB   int `json:"disk_gb"`
}

type OrgAlloc struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	CPUCores int    `json:"cpu_cores"`
	RAMMB    int    `json:"ram_mb"`
	DiskGB   int    `json:"disk_gb"`
}

// GetPlatformCapacity reports the host's total capacity (via Docker info +
// statfs on /) and the aggregate quotas allocated to existing orgs.
func (h *AdminHandler) GetPlatformCapacity(c *gin.Context) {
	var host ResourcePool
	info, err := h.dockerClient.HostInfo(c.Request.Context())
	if err == nil && info != nil {
		host.CPUCores = info.CPUCount
		host.RAMMB = info.MemoryMB
	}

	// Disk — platform-specific syscall. statfs on the root filesystem is the
	// closest we can get from inside a container. Implemented in
	// admin_handler_linux.go; other platforms return 0.
	host.DiskGB = diskGBRoot()

	// Sum allocations across all orgs
	orgs, err := h.orgService.ListAllOrgs(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to list organizations")
		return
	}

	var allocated ResourcePool
	allocList := make([]OrgAlloc, 0, len(orgs))
	for _, o := range orgs {
		cpu := o.EffectiveCPUCores()
		ram := o.EffectiveRAMMB()
		disk := o.EffectiveDiskGB()
		allocated.CPUCores += cpu
		allocated.RAMMB += ram
		allocated.DiskGB += disk
		allocList = append(allocList, OrgAlloc{
			Slug:     o.Slug,
			Name:     o.Name,
			CPUCores: cpu,
			RAMMB:    ram,
			DiskGB:   disk,
		})
	}

	available := ResourcePool{
		CPUCores: max(host.CPUCores-allocated.CPUCores, 0),
		RAMMB:    max(host.RAMMB-allocated.RAMMB, 0),
		DiskGB:   max(host.DiskGB-allocated.DiskGB, 0),
	}

	response.Success(c, http.StatusOK, Capacity{
		Host:      host,
		Allocated: allocated,
		Available: available,
		Orgs:      allocList,
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
