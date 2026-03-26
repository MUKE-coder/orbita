package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
	"github.com/orbita-sh/orbita/internal/response"
)

func RequireOrgMember(orgRepo *repository.OrgRepository, minRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgSlug := c.Param("orgSlug")
		if orgSlug == "" {
			response.BadRequest(c, "Organization slug is required")
			c.Abort()
			return
		}

		org, err := orgRepo.FindOrgBySlug(c.Request.Context(), orgSlug)
		if err != nil {
			response.NotFound(c, "Organization not found")
			c.Abort()
			return
		}

		userID := GetUserIDFromContext(c)

		member, err := orgRepo.GetMember(c.Request.Context(), org.ID, userID)
		if err != nil {
			// Check if user is super admin
			isSuperAdmin, _ := c.Get("is_super_admin")
			if isSuperAdmin == true {
				c.Set("org", org)
				c.Set("org_id", org.ID)
				c.Set("member_role", models.RoleOwner)
				c.Next()
				return
			}

			response.Forbidden(c, "You are not a member of this organization")
			c.Abort()
			return
		}

		if !models.HasMinRole(member.Role, minRole) {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Set("org", org)
		c.Set("org_id", org.ID)
		c.Set("member_role", member.Role)

		c.Next()
	}
}

func GetOrgFromContext(c *gin.Context) *models.Organization {
	org, exists := c.Get("org")
	if !exists {
		return nil
	}
	return org.(*models.Organization)
}

func GetOrgIDFromContext(c *gin.Context) uuid.UUID {
	id, _ := c.Get("org_id")
	return id.(uuid.UUID)
}

func GetMemberRoleFromContext(c *gin.Context) string {
	role, _ := c.Get("member_role")
	return role.(string)
}
