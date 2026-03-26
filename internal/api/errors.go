package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

// RespondError maps domain errors to HTTP responses
func RespondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.NotFound(c, "Resource not found")
	case errors.Is(err, service.ErrInvalidCredentials):
		response.Unauthorized(c, "Invalid credentials")
	case errors.Is(err, service.ErrEmailAlreadyExists):
		response.Conflict(c, "Email already registered")
	case errors.Is(err, service.ErrUserNotFound):
		response.NotFound(c, "User not found")
	case errors.Is(err, service.ErrInvalidToken):
		response.Unauthorized(c, "Invalid or expired token")
	case errors.Is(err, service.ErrOrgNotFound):
		response.NotFound(c, "Organization not found")
	case errors.Is(err, service.ErrOrgSlugTaken):
		response.Conflict(c, "Organization slug already taken")
	case errors.Is(err, service.ErrMemberNotFound):
		response.NotFound(c, "Member not found")
	case errors.Is(err, service.ErrAlreadyMember):
		response.Conflict(c, "Already a member")
	case errors.Is(err, service.ErrInsufficientRole):
		response.Forbidden(c, "Insufficient permissions")
	case errors.Is(err, service.ErrAppNotFound):
		response.NotFound(c, "Application not found")
	case errors.Is(err, service.ErrDBNotFound):
		response.NotFound(c, "Database not found")
	case errors.Is(err, service.ErrProjectNotFound):
		response.NotFound(c, "Project not found")
	case errors.Is(err, service.ErrDomainTaken):
		response.Conflict(c, "Domain already in use")
	default:
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
	}
}
