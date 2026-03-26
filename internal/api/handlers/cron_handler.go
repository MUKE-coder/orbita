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

type CronHandler struct {
	cronService *service.CronService
}

func NewCronHandler(cronService *service.CronService) *CronHandler {
	return &CronHandler{cronService: cronService}
}

type CreateCronRequest struct {
	Name              string  `json:"name" binding:"required,min=2"`
	Schedule          string  `json:"schedule" binding:"required"`
	Image             string  `json:"image" binding:"required"`
	Command           *string `json:"command"`
	EnvironmentID     string  `json:"environment_id" binding:"required"`
	Timeout           int     `json:"timeout"`
	ConcurrencyPolicy string  `json:"concurrency_policy" binding:"omitempty,oneof=allow forbid replace"`
	MaxRetries        int     `json:"max_retries" binding:"omitempty,min=0,max=5"`
	CPULimit          int     `json:"cpu_limit"`
	MemoryLimit       int     `json:"memory_limit"`
}

func (h *CronHandler) ListCronJobs(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	jobs, err := h.cronService.ListCronJobs(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, "Failed to list cron jobs")
		return
	}
	response.Success(c, http.StatusOK, jobs)
}

func (h *CronHandler) CreateCronJob(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	var req CreateCronRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	envID, err := uuid.Parse(req.EnvironmentID)
	if err != nil {
		response.BadRequest(c, "Invalid environment ID")
		return
	}

	job, err := h.cronService.CreateCronJob(c.Request.Context(), orgID, service.CreateCronInput{
		Name:              req.Name,
		Schedule:          req.Schedule,
		Image:             req.Image,
		Command:           req.Command,
		EnvironmentID:     envID,
		Timeout:           req.Timeout,
		ConcurrencyPolicy: req.ConcurrencyPolicy,
		MaxRetries:        req.MaxRetries,
		CPULimit:          req.CPULimit,
		MemoryLimit:       req.MemoryLimit,
	})
	if err != nil {
		response.InternalError(c, "Failed to create cron job")
		return
	}

	response.Success(c, http.StatusCreated, job)
}

func (h *CronHandler) GetCronJob(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	cronID, err := uuid.Parse(c.Param("cronId"))
	if err != nil {
		response.BadRequest(c, "Invalid cron job ID")
		return
	}

	job, err := h.cronService.GetCronJob(c.Request.Context(), cronID, orgID)
	if err != nil {
		if errors.Is(err, service.ErrCronJobNotFound) {
			response.NotFound(c, "Cron job not found")
			return
		}
		response.InternalError(c, "Failed to get cron job")
		return
	}
	response.Success(c, http.StatusOK, job)
}

func (h *CronHandler) UpdateCronJob(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	cronID, _ := uuid.Parse(c.Param("cronId"))

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	job, err := h.cronService.UpdateCronJob(c.Request.Context(), cronID, orgID, updates)
	if err != nil {
		if errors.Is(err, service.ErrCronJobNotFound) {
			response.NotFound(c, "Cron job not found")
			return
		}
		response.InternalError(c, "Failed to update cron job")
		return
	}
	response.Success(c, http.StatusOK, job)
}

func (h *CronHandler) DeleteCronJob(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	cronID, _ := uuid.Parse(c.Param("cronId"))

	if err := h.cronService.DeleteCronJob(c.Request.Context(), cronID, orgID); err != nil {
		if errors.Is(err, service.ErrCronJobNotFound) {
			response.NotFound(c, "Cron job not found")
			return
		}
		response.InternalError(c, "Failed to delete cron job")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Cron job deleted"})
}

func (h *CronHandler) ToggleCronJob(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	cronID, _ := uuid.Parse(c.Param("cronId"))

	job, err := h.cronService.ToggleCronJob(c.Request.Context(), cronID, orgID)
	if err != nil {
		response.InternalError(c, "Failed to toggle cron job")
		return
	}
	response.Success(c, http.StatusOK, job)
}

func (h *CronHandler) TriggerCronJob(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	cronID, _ := uuid.Parse(c.Param("cronId"))

	if err := h.cronService.TriggerCronJob(c.Request.Context(), cronID, orgID); err != nil {
		response.InternalError(c, "Failed to trigger cron job")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Cron job triggered"})
}

func (h *CronHandler) ListRuns(c *gin.Context) {
	cronID, _ := uuid.Parse(c.Param("cronId"))
	runs, err := h.cronService.ListRuns(c.Request.Context(), cronID, 50)
	if err != nil {
		response.InternalError(c, "Failed to list runs")
		return
	}
	response.Success(c, http.StatusOK, runs)
}

func (h *CronHandler) GetRunLogs(c *gin.Context) {
	runID, _ := uuid.Parse(c.Param("runId"))
	logs, err := h.cronService.GetRunLogs(c.Request.Context(), runID)
	if err != nil {
		response.InternalError(c, "Failed to get run logs")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"logs": logs})
}
