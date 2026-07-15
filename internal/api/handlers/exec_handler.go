package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/middleware"
	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

type ExecHandler struct {
	appService *service.AppService
}

func NewExecHandler(appService *service.AppService) *ExecHandler {
	return &ExecHandler{appService: appService}
}

type ExecRequest struct {
	Command string `json:"command" binding:"required"`
}

func (h *ExecHandler) ExecInApp(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		response.BadRequest(c, "Invalid app ID")
		return
	}

	var req ExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	output, exitCode, err := h.appService.ExecCommand(c.Request.Context(), appID, orgID, req.Command)
	if err != nil {
		if errors.Is(err, service.ErrAppNotFound) {
			response.NotFound(c, "App not found")
			return
		}
		log.Error().Err(err).Str("app_id", appID.String()).Msg("Exec failed")
		response.InternalError(c, "Exec failed: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"output":    output,
		"exit_code": exitCode,
	})
}
