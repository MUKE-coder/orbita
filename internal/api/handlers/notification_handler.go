package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/middleware"
	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

type NotificationHandler struct {
	notifService *service.NotificationService
}

func NewNotificationHandler(notifService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

type UpdateSettingRequest struct {
	EventType      string  `json:"event_type" binding:"required"`
	EmailEnabled   bool    `json:"email_enabled"`
	WebhookEnabled bool    `json:"webhook_enabled"`
	WebhookURL     *string `json:"webhook_url"`
}

func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	notifs, unread, err := h.notifService.ListNotifications(c.Request.Context(), userID, 50)
	if err != nil {
		response.InternalError(c, "Failed to list notifications")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"notifications": notifs,
		"unread_count":  unread,
	})
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)
	notifID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid notification ID")
		return
	}

	if err := h.notifService.MarkRead(c.Request.Context(), notifID, userID); err != nil {
		response.InternalError(c, "Failed to mark as read")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Marked as read"})
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	if err := h.notifService.MarkAllRead(c.Request.Context(), userID); err != nil {
		response.InternalError(c, "Failed to mark all as read")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "All marked as read"})
}

func (h *NotificationHandler) GetSettings(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	userID := middleware.GetUserIDFromContext(c)

	settings, err := h.notifService.GetSettings(c.Request.Context(), orgID, userID)
	if err != nil {
		response.InternalError(c, "Failed to get settings")
		return
	}
	response.Success(c, http.StatusOK, settings)
}

func (h *NotificationHandler) UpdateSettings(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	userID := middleware.GetUserIDFromContext(c)

	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.notifService.UpdateSettings(c.Request.Context(), orgID, userID, req.EventType, req.EmailEnabled, req.WebhookEnabled, req.WebhookURL); err != nil {
		response.InternalError(c, "Failed to update settings")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Settings updated"})
}

func (h *NotificationHandler) ListAuditLogs(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if pageSize > 100 {
		pageSize = 100
	}

	logs, total, err := h.notifService.ListAuditLogs(c.Request.Context(), orgID, page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to list audit logs")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"audit_logs": logs,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}
