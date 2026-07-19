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

type ProvisioningHandler struct {
	provisioning *service.ProvisioningService
}

func NewProvisioningHandler(p *service.ProvisioningService) *ProvisioningHandler {
	return &ProvisioningHandler{provisioning: p}
}

// ProvisionOrgRequest onboards a tenant: the organisation, its size, and the
// account its admin signs in with.
type ProvisionOrgRequest struct {
	Name        string  `json:"name" binding:"required,min=2"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	PlanID      string  `json:"plan_id"`

	// Resource overrides. Omit to inherit the plan's value.
	CustomCPUCores     *int `json:"custom_cpu_cores"`
	CustomRAMMB        *int `json:"custom_ram_mb"`
	CustomDiskGB       *int `json:"custom_disk_gb"`
	CustomMaxApps      *int `json:"custom_max_apps"`
	CustomMaxDatabases *int `json:"custom_max_databases"`

	// Billing
	BillingType       string `json:"billing_type"`
	PriceMonthlyCents *int   `json:"price_monthly_cents"`
	Currency          string `json:"currency"`
	BillingCycle      string `json:"billing_cycle"`

	// The org's admin account
	AdminEmail    string `json:"admin_email" binding:"required,email"`
	AdminName     string `json:"admin_name"`
	AdminPassword string `json:"admin_password"` // empty = generate one
	AdminRole     string `json:"admin_role"`     // default: owner
}

// CreateOrgUserRequest adds a person to an existing org without email.
type CreateOrgUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name"`
	Password string `json:"password"` // empty = generate one
	Role     string `json:"role"`
}

func (h *ProvisioningHandler) ProvisionOrg(c *gin.Context) {
	actor := middleware.GetUserIDFromContext(c)

	var req ProvisionOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	in := service.ProvisionOrgInput{
		Name:               req.Name,
		Slug:               req.Slug,
		Description:        req.Description,
		CustomCPUCores:     req.CustomCPUCores,
		CustomRAMMB:        req.CustomRAMMB,
		CustomDiskGB:       req.CustomDiskGB,
		CustomMaxApps:      req.CustomMaxApps,
		CustomMaxDatabases: req.CustomMaxDatabases,
		BillingType:        req.BillingType,
		PriceMonthlyCents:  req.PriceMonthlyCents,
		Currency:           req.Currency,
		BillingCycle:       req.BillingCycle,
		AdminEmail:         req.AdminEmail,
		AdminName:          req.AdminName,
		AdminPassword:      req.AdminPassword,
		AdminRole:          req.AdminRole,
	}
	if req.PlanID != "" {
		planID, err := uuid.Parse(req.PlanID)
		if err != nil {
			response.BadRequest(c, "Invalid plan_id")
			return
		}
		in.PlanID = &planID
	}

	result, err := h.provisioning.ProvisionOrgWithAdmin(c.Request.Context(), in, actor)
	if err != nil {
		writeProvisioningError(c, err)
		return
	}

	// The password appears here and nowhere else — it is not stored in
	// plaintext and cannot be retrieved again.
	response.Success(c, http.StatusCreated, result)
}

func (h *ProvisioningHandler) CreateOrgUser(c *gin.Context) {
	actor := middleware.GetUserIDFromContext(c)
	orgID := middleware.GetOrgIDFromContext(c)

	var req CreateOrgUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.provisioning.CreateUserInOrg(c.Request.Context(), orgID, service.CreateUserInOrgInput{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
		Role:     req.Role,
	}, actor)
	if err != nil {
		writeProvisioningError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, result)
}

func writeProvisioningError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserEmailTaken):
		response.Conflict(c, err.Error())
	case errors.Is(err, service.ErrInvalidOrgRole), errors.Is(err, service.ErrWeakPassword):
		response.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrOrgSlugTaken):
		response.Conflict(c, "That organisation slug is already taken")
	default:
		response.InternalError(c, "Failed to provision: "+err.Error())
	}
}
