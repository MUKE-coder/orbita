package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/orbita-sh/orbita/internal/mailer"
	"github.com/orbita-sh/orbita/internal/middleware"
	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

type SettingsHandler struct {
	settings *service.SettingsService
	mailer   *mailer.Mailer
}

func NewSettingsHandler(settings *service.SettingsService, mail *mailer.Mailer) *SettingsHandler {
	return &SettingsHandler{settings: settings, mailer: mail}
}

// UpdateEmailSettingsRequest updates the instance's email provider.
//
// APIKey is a pointer so three cases stay distinct: absent (leave the stored
// key alone), empty string (clear it), and a value (replace it). The stored key
// is never returned, so the UI can't round-trip it back.
type UpdateEmailSettingsRequest struct {
	APIKey        *string `json:"api_key"`
	EmailFrom     string  `json:"email_from"`
	EmailFromName string  `json:"email_from_name"`
}

type TestEmailRequest struct {
	To string `json:"to" binding:"required,email"`
}

func (h *SettingsHandler) GetEmailSettings(c *gin.Context) {
	s, err := h.settings.GetEmailSettings(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to load settings")
		return
	}
	response.Success(c, http.StatusOK, s)
}

func (h *SettingsHandler) UpdateEmailSettings(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	var req UpdateEmailSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.settings.UpdateEmailSettings(c.Request.Context(), service.UpdateEmailSettingsInput{
		APIKey:        req.APIKey,
		EmailFrom:     req.EmailFrom,
		EmailFromName: req.EmailFromName,
	}, userID); err != nil {
		response.InternalError(c, "Failed to save settings")
		return
	}

	s, err := h.settings.GetEmailSettings(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Saved, but could not reload settings")
		return
	}
	response.Success(c, http.StatusOK, s)
}

// SendTestEmail proves the credentials work, so a misconfiguration surfaces here
// rather than when a real invite silently fails to arrive.
func (h *SettingsHandler) SendTestEmail(c *gin.Context) {
	var req TestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.mailer.SendTest(c.Request.Context(), req.To); err != nil {
		if errors.Is(err, mailer.ErrNotConfigured) {
			response.BadRequest(c, "Email is not configured yet — add an API key and from-address first.")
			return
		}
		// The provider's message is the useful part (bad key, unverified domain).
		response.BadRequest(c, "Send failed: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Test email sent to " + req.To})
}
