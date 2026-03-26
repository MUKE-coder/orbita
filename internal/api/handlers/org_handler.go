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

type OrgHandler struct {
	orgService *service.OrgService
}

func NewOrgHandler(orgService *service.OrgService) *OrgHandler {
	return &OrgHandler{orgService: orgService}
}

type CreateOrgRequest struct {
	Name string `json:"name" binding:"required,min=2"`
	Slug string `json:"slug" binding:"required,min=2"`
}

type UpdateOrgRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin developer viewer"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin developer viewer"`
}

// List user's organizations
func (h *OrgHandler) ListOrgs(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	orgs, err := h.orgService.ListUserOrgs(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to list organizations")
		return
	}
	response.Success(c, http.StatusOK, orgs)
}

// Create organization
func (h *OrgHandler) CreateOrg(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	var req CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	org, err := h.orgService.CreateOrganization(c.Request.Context(), userID, req.Name, req.Slug)
	if err != nil {
		if errors.Is(err, service.ErrOrgSlugTaken) {
			response.Conflict(c, "Organization slug already taken")
			return
		}
		response.InternalError(c, "Failed to create organization")
		return
	}

	response.Success(c, http.StatusCreated, org)
}

// Get organization details
func (h *OrgHandler) GetOrg(c *gin.Context) {
	org := middleware.GetOrgFromContext(c)
	memberCount, _ := h.orgService.ListMembers(c.Request.Context(), org.ID)

	response.Success(c, http.StatusOK, gin.H{
		"organization": org,
		"member_count": len(memberCount),
	})
}

// Update organization
func (h *OrgHandler) UpdateOrg(c *gin.Context) {
	orgSlug := c.Param("orgSlug")

	var req UpdateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	org, err := h.orgService.UpdateOrganization(c.Request.Context(), orgSlug, req.Name, req.Description)
	if err != nil {
		response.InternalError(c, "Failed to update organization")
		return
	}

	response.Success(c, http.StatusOK, org)
}

// Delete organization
func (h *OrgHandler) DeleteOrg(c *gin.Context) {
	orgSlug := c.Param("orgSlug")
	role := middleware.GetMemberRoleFromContext(c)

	if role != models.RoleOwner {
		response.Forbidden(c, "Only the owner can delete the organization")
		return
	}

	if err := h.orgService.DeleteOrganization(c.Request.Context(), orgSlug); err != nil {
		response.InternalError(c, "Failed to delete organization")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Organization deleted"})
}

// List members
func (h *OrgHandler) ListMembers(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	members, err := h.orgService.ListMembers(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, "Failed to list members")
		return
	}
	response.Success(c, http.StatusOK, members)
}

// Invite member
func (h *OrgHandler) InviteMember(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	inviterID := middleware.GetUserIDFromContext(c)

	var req InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.orgService.InviteMember(c.Request.Context(), orgID, req.Email, req.Role, inviterID); err != nil {
		if errors.Is(err, service.ErrAlreadyMember) {
			response.Conflict(c, "User is already a member")
			return
		}
		response.InternalError(c, "Failed to send invite")
		return
	}

	response.Success(c, http.StatusCreated, gin.H{"message": "Invitation sent"})
}

// List pending invites
func (h *OrgHandler) ListInvites(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	invites, err := h.orgService.ListPendingInvites(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, "Failed to list invites")
		return
	}
	response.Success(c, http.StatusOK, invites)
}

// Revoke invite
func (h *OrgHandler) RevokeInvite(c *gin.Context) {
	inviteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid invite ID")
		return
	}

	if err := h.orgService.RevokeInvite(c.Request.Context(), inviteID); err != nil {
		response.InternalError(c, "Failed to revoke invite")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Invite revoked"})
}

// Update member role
func (h *OrgHandler) UpdateMemberRole(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	requesterID := middleware.GetUserIDFromContext(c)

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.orgService.UpdateMemberRole(c.Request.Context(), orgID, targetUserID, requesterID, req.Role); err != nil {
		if errors.Is(err, service.ErrSelfRoleChange) {
			response.BadRequest(c, "Cannot change your own role")
			return
		}
		if errors.Is(err, service.ErrMemberNotFound) {
			response.NotFound(c, "Member not found")
			return
		}
		if errors.Is(err, service.ErrInsufficientRole) {
			response.Forbidden(c, "Cannot change the owner's role")
			return
		}
		response.InternalError(c, "Failed to update role")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Role updated"})
}

// Remove member
func (h *OrgHandler) RemoveMember(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	requesterID := middleware.GetUserIDFromContext(c)

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.orgService.RemoveMember(c.Request.Context(), orgID, targetUserID, requesterID); err != nil {
		if errors.Is(err, service.ErrCannotLeave) {
			response.BadRequest(c, "Cannot remove the owner")
			return
		}
		if errors.Is(err, service.ErrMemberNotFound) {
			response.NotFound(c, "Member not found")
			return
		}
		response.InternalError(c, "Failed to remove member")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Member removed"})
}

// Leave org
func (h *OrgHandler) LeaveOrg(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	userID := middleware.GetUserIDFromContext(c)

	if err := h.orgService.LeaveOrg(c.Request.Context(), orgID, userID); err != nil {
		if errors.Is(err, service.ErrCannotLeave) {
			response.BadRequest(c, "Owner cannot leave the organization")
			return
		}
		response.InternalError(c, "Failed to leave organization")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Left organization"})
}

// Get invite info (public)
func (h *OrgHandler) GetInviteInfo(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.BadRequest(c, "Token is required")
		return
	}

	invite, org, err := h.orgService.GetInviteInfo(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, service.ErrInviteNotFound) {
			response.NotFound(c, "Invite not found or expired")
			return
		}
		response.InternalError(c, "Failed to get invite info")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"organization": org.Name,
		"org_slug":     org.Slug,
		"role":         invite.Role,
		"email":        invite.Email,
		"inviter":      invite.Inviter,
	})
}

// Accept invite (authenticated)
func (h *OrgHandler) AcceptInvite(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	token := c.Query("token")
	if token == "" {
		var body struct {
			Token string `json:"token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			response.BadRequest(c, "Token is required")
			return
		}
		token = body.Token
	}

	if err := h.orgService.AcceptInvite(c.Request.Context(), token, userID); err != nil {
		if errors.Is(err, service.ErrInviteNotFound) {
			response.NotFound(c, "Invite not found or expired")
			return
		}
		if errors.Is(err, service.ErrAlreadyMember) {
			response.Conflict(c, "You are already a member")
			return
		}
		response.InternalError(c, "Failed to accept invite")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Invitation accepted"})
}
