package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/repository"
	"github.com/orbita-sh/orbita/internal/response"
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

// ApiKeyAuth authenticates either an `orb_` API key or a JWT bearer token.
// API keys act as the user who created them (RBAC applies as that user), can
// be bound to a single org, and carry scopes: "read" (GET only), "deploy"
// (read + resource writes), "admin" (everything).
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

			if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
				response.Unauthorized(c, "API key expired")
				c.Abort()
				return
			}

			if !apiKeyScopeAllows(apiKey.Scopes, c.Request.Method) {
				response.Forbidden(c, "API key scope does not permit this operation")
				c.Abort()
				return
			}

			c.Set("user_id", apiKey.UserID)
			c.Set("api_key_id", apiKey.ID.String())
			c.Set("api_key_scopes", []string(apiKey.Scopes))
			if apiKey.OrgID != nil {
				// Org-bound key: RequireOrgMember rejects use on other orgs
				c.Set("api_key_org_id", *apiKey.OrgID)
			}

			keyID := apiKey.ID
			go func() {
				_ = userRepo.UpdateAPIKeyLastUsed(context.Background(), keyID)
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
		if claims.OrgID != nil {
			c.Set("org_id", *claims.OrgID)
		}

		c.Next()
	}
}

// apiKeyScopeAllows maps scopes to permitted HTTP methods. "admin" allows
// everything; "deploy" allows reads and writes; "read" allows reads only.
// A key with no recognized scope can do nothing.
func apiKeyScopeAllows(scopes []string, method string) bool {
	isRead := method == "GET" || method == "HEAD" || method == "OPTIONS"
	for _, s := range scopes {
		switch s {
		case "admin", "deploy":
			return true
		case "read":
			if isRead {
				return true
			}
		}
	}
	return false
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
