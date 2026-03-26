package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/repository"
)

func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearerToken(c)
		if tokenString == "" {
			response.Unauthorized(c, "Missing or invalid authorization header")
			c.Abort()
			return
		}

		claims, err := auth.ValidateAccessToken(tokenString, jwtSecret)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("is_super_admin", claims.IsSuperAdmin)
		if claims.OrgID != nil {
			c.Set("org_id", *claims.OrgID)
		}

		c.Next()
	}
}

func RequireSuperAdmin(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearerToken(c)
		if tokenString == "" {
			response.Unauthorized(c, "Missing or invalid authorization header")
			c.Abort()
			return
		}

		claims, err := auth.ValidateAccessToken(tokenString, jwtSecret)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		if !claims.IsSuperAdmin {
			response.Forbidden(c, "Super admin access required")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("is_super_admin", claims.IsSuperAdmin)

		c.Next()
	}
}

func OptionalAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearerToken(c)
		if tokenString == "" {
			c.Next()
			return
		}

		claims, err := auth.ValidateAccessToken(tokenString, jwtSecret)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("is_super_admin", claims.IsSuperAdmin)
		if claims.OrgID != nil {
			c.Set("org_id", *claims.OrgID)
		}

		c.Next()
	}
}

func ApiKeyAuth(userRepo *repository.UserRepository, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearerToken(c)
		if tokenString == "" {
			response.Unauthorized(c, "Missing authorization")
			c.Abort()
			return
		}

		// Try API key first (starts with "orb_")
		if strings.HasPrefix(tokenString, "orb_") {
			hash := auth.HashToken(tokenString)
			apiKey, err := userRepo.FindAPIKeyByHash(c.Request.Context(), hash)
			if err != nil {
				response.Unauthorized(c, "Invalid API key")
				c.Abort()
				return
			}

			c.Set("user_id", apiKey.UserID)
			c.Set("api_key_id", apiKey.ID.String())
			c.Set("api_key_scopes", apiKey.Scopes)
			if apiKey.OrgID != nil {
				c.Set("org_id", *apiKey.OrgID)
			}

			go func() {
				_ = userRepo.UpdateAPIKeyLastUsed(c.Request.Context(), apiKey.ID)
			}()

			c.Next()
			return
		}

		// Fall back to JWT
		claims, err := auth.ValidateAccessToken(tokenString, jwtSecret)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("is_super_admin", claims.IsSuperAdmin)

		c.Next()
	}
}

func GetUserIDFromContext(c *gin.Context) uuid.UUID {
	id, _ := c.Get("user_id")
	return id.(uuid.UUID)
}

func extractBearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}
