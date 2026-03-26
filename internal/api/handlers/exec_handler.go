package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/middleware"
	"github.com/orbita-sh/orbita/internal/response"
)

type ExecHandler struct{}

func NewExecHandler() *ExecHandler {
	return &ExecHandler{}
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

	// TODO: real impl
	// 1. Find running container for app
	// 2. Create temporary container with same image/env
	// 3. Run command, capture output
	// 4. Return output in response body

	log.Info().
		Str("app_id", appID.String()).
		Str("org_id", orgID.String()).
		Str("command", req.Command).
		Msg("Exec command in app (stub)")

	output := "$ " + req.Command + "\n" +
		"Command executed successfully (stub output)\n"

	response.Success(c, http.StatusOK, gin.H{
		"output":    output,
		"exit_code": 0,
	})
}
