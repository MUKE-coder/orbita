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

type DomainHandler struct {
	domainService *service.DomainService
}

func NewDomainHandler(domainService *service.DomainService) *DomainHandler {
	return &DomainHandler{domainService: domainService}
}

type AddDomainRequest struct {
	Domain     string `json:"domain" binding:"required"`
	SSLEnabled *bool  `json:"ssl_enabled"`
	Port       int    `json:"port"`
}

// List all domains in org
func (h *DomainHandler) ListOrgDomains(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	domains, err := h.domainService.ListDomainsByOrg(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, "Failed to list domains")
		return
	}
	response.Success(c, http.StatusOK, domains)
}

// List domains for a specific app
func (h *DomainHandler) ListAppDomains(c *gin.Context) {
	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		response.BadRequest(c, "Invalid app ID")
		return
	}
	domains, err := h.domainService.ListDomainsByResource(c.Request.Context(), appID, models.ResourceTypeApp)
	if err != nil {
		response.InternalError(c, "Failed to list domains")
		return
	}
	response.Success(c, http.StatusOK, domains)
}

// Add domain to an app
func (h *DomainHandler) AddAppDomain(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		response.BadRequest(c, "Invalid app ID")
		return
	}

	var req AddDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ssl := true
	if req.SSLEnabled != nil {
		ssl = *req.SSLEnabled
	}
	port := req.Port
	if port == 0 {
		port = 80
	}

	domain, err := h.domainService.AddDomain(c.Request.Context(), appID, models.ResourceTypeApp, req.Domain, orgID, ssl, port)
	if err != nil {
		if errors.Is(err, service.ErrDomainTaken) {
			response.Conflict(c, "Domain already in use by another organization")
			return
		}
		response.InternalError(c, "Failed to add domain")
		return
	}
	response.Success(c, http.StatusCreated, domain)
}

// Remove domain
func (h *DomainHandler) RemoveDomain(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	domainID, err := uuid.Parse(c.Param("domainId"))
	if err != nil {
		response.BadRequest(c, "Invalid domain ID")
		return
	}

	if err := h.domainService.RemoveDomain(c.Request.Context(), domainID, orgID); err != nil {
		if errors.Is(err, service.ErrDomainNotFound) {
			response.NotFound(c, "Domain not found")
			return
		}
		response.InternalError(c, "Failed to remove domain")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Domain removed"})
}

// Verify domain DNS
func (h *DomainHandler) VerifyDomain(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		response.BadRequest(c, "Domain parameter required")
		return
	}

	verified, _ := h.domainService.VerifyDomain(c.Request.Context(), domain)
	response.Success(c, http.StatusOK, gin.H{
		"domain":   domain,
		"verified": verified,
	})
}
