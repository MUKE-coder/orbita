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

type MeHandler struct {
	authService *service.AuthService
}

func NewMeHandler(authService *service.AuthService) *MeHandler {
	return &MeHandler{authService: authService}
}

type UpdateProfileRequest struct {
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type CreateAPIKeyRequest struct {
	Name   string   `json:"name" binding:"required"`
	Scopes []string `json:"scopes" binding:"required"`
}

func (h *MeHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	user, err := h.authService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to get profile")
		return
	}

	response.Success(c, http.StatusOK, user)
}

func (h *MeHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), userID, req.Name, req.AvatarURL)
	if err != nil {
		response.InternalError(c, "Failed to update profile")
		return
	}

	response.Success(c, http.StatusOK, user)
}

func (h *MeHandler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.BadRequest(c, "Current password is incorrect")
			return
		}
		response.InternalError(c, "Failed to change password")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (h *MeHandler) GetSessions(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	sessions, err := h.authService.GetSessions(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to get sessions")
		return
	}

	response.Success(c, http.StatusOK, sessions)
}

func (h *MeHandler) RevokeSession(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid session ID")
		return
	}

	if err := h.authService.RevokeSession(c.Request.Context(), sessionID, userID); err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			response.NotFound(c, "Session not found")
			return
		}
		response.InternalError(c, "Failed to revoke session")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Session revoked"})
}

func (h *MeHandler) ListAPIKeys(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	keys, err := h.authService.ListAPIKeys(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, "Failed to list API keys")
		return
	}

	response.Success(c, http.StatusOK, keys)
}

func (h *MeHandler) CreateAPIKey(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	apiKey, rawKey, err := h.authService.CreateAPIKey(c.Request.Context(), userID, req.Name, req.Scopes)
	if err != nil {
		response.InternalError(c, "Failed to create API key")
		return
	}

	response.Success(c, http.StatusCreated, gin.H{
		"api_key": apiKey,
		"key":     rawKey,
	})
}

func (h *MeHandler) DeleteAPIKey(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid API key ID")
		return
	}

	if err := h.authService.DeleteAPIKey(c.Request.Context(), keyID, userID); err != nil {
		response.InternalError(c, "Failed to delete API key")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "API key deleted"})
}
