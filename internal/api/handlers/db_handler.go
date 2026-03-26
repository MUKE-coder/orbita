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

type DBHandler struct {
	dbService *service.DBService
}

func NewDBHandler(dbService *service.DBService) *DBHandler {
	return &DBHandler{dbService: dbService}
}

type CreateDBRequest struct {
	Name          string `json:"name" binding:"required,min=2"`
	Engine        string `json:"engine" binding:"required,oneof=postgres mysql mariadb mongodb redis"`
	Version       string `json:"version" binding:"required"`
	EnvironmentID string `json:"environment_id" binding:"required"`
	CPULimit      int    `json:"cpu_limit"`
	MemoryLimit   int    `json:"memory_limit"`
}

type SetBackupScheduleRequest struct {
	Frequency      string `json:"frequency" binding:"required,oneof=hourly daily weekly"`
	RetentionCount int    `json:"retention_count" binding:"required,min=1,max=100"`
}

func (h *DBHandler) ListDatabases(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	dbs, err := h.dbService.ListDatabases(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, "Failed to list databases")
		return
	}
	response.Success(c, http.StatusOK, dbs)
}

func (h *DBHandler) CreateDatabase(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	org := middleware.GetOrgFromContext(c)

	var req CreateDBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	envID, err := uuid.Parse(req.EnvironmentID)
	if err != nil {
		response.BadRequest(c, "Invalid environment ID")
		return
	}

	mdb, err := h.dbService.CreateDatabase(c.Request.Context(), orgID, org.Slug, service.CreateDBInput{
		Name:          req.Name,
		Engine:        req.Engine,
		Version:       req.Version,
		EnvironmentID: envID,
		CPULimit:      req.CPULimit,
		MemoryLimit:   req.MemoryLimit,
	})
	if err != nil {
		if mdb != nil {
			response.Success(c, http.StatusCreated, gin.H{"database": mdb, "warning": err.Error()})
			return
		}
		response.InternalError(c, "Failed to create database")
		return
	}
	response.Success(c, http.StatusCreated, mdb)
}

func (h *DBHandler) GetDatabase(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	dbID, err := uuid.Parse(c.Param("dbId"))
	if err != nil {
		response.BadRequest(c, "Invalid database ID")
		return
	}

	mdb, err := h.dbService.GetDatabase(c.Request.Context(), dbID, orgID)
	if err != nil {
		if errors.Is(err, service.ErrDBNotFound) {
			response.NotFound(c, "Database not found")
			return
		}
		response.InternalError(c, "Failed to get database")
		return
	}

	// Get connection string (masked by default)
	connStr, _ := h.dbService.GetConnectionString(c.Request.Context(), dbID, orgID)
	showConn := c.Query("show_connection") == "true"

	result := gin.H{"database": mdb}
	if showConn && connStr != "" {
		result["connection_string"] = connStr
	}

	response.Success(c, http.StatusOK, result)
}

func (h *DBHandler) DeleteDatabase(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	dbID, _ := uuid.Parse(c.Param("dbId"))

	if err := h.dbService.DeleteDatabase(c.Request.Context(), dbID, orgID); err != nil {
		if errors.Is(err, service.ErrDBNotFound) {
			response.NotFound(c, "Database not found")
			return
		}
		response.InternalError(c, "Failed to delete database")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Database deleted"})
}

func (h *DBHandler) RestartDatabase(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	dbID, _ := uuid.Parse(c.Param("dbId"))
	if err := h.dbService.RestartDatabase(c.Request.Context(), dbID, orgID); err != nil {
		response.InternalError(c, "Failed to restart database")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Database restarted"})
}

func (h *DBHandler) StopDatabase(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	dbID, _ := uuid.Parse(c.Param("dbId"))
	if err := h.dbService.StopDatabase(c.Request.Context(), dbID, orgID); err != nil {
		response.InternalError(c, "Failed to stop database")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Database stopped"})
}

func (h *DBHandler) StartDatabase(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	dbID, _ := uuid.Parse(c.Param("dbId"))
	if err := h.dbService.StartDatabase(c.Request.Context(), dbID, orgID); err != nil {
		response.InternalError(c, "Failed to start database")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Database started"})
}

func (h *DBHandler) CreateBackup(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	dbID, _ := uuid.Parse(c.Param("dbId"))
	backup, err := h.dbService.CreateBackup(c.Request.Context(), dbID, orgID)
	if err != nil {
		response.InternalError(c, "Failed to create backup")
		return
	}
	response.Success(c, http.StatusCreated, backup)
}

func (h *DBHandler) ListBackups(c *gin.Context) {
	dbID, _ := uuid.Parse(c.Param("dbId"))
	backups, err := h.dbService.ListBackups(c.Request.Context(), dbID, 20)
	if err != nil {
		response.InternalError(c, "Failed to list backups")
		return
	}
	response.Success(c, http.StatusOK, backups)
}

func (h *DBHandler) RestoreBackup(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	dbID, _ := uuid.Parse(c.Param("dbId"))
	backupID, _ := uuid.Parse(c.Param("backupId"))
	if err := h.dbService.RestoreBackup(c.Request.Context(), dbID, backupID, orgID); err != nil {
		response.InternalError(c, "Failed to restore backup")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"message": "Restore initiated"})
}

func (h *DBHandler) GetBackupSchedule(c *gin.Context) {
	dbID, _ := uuid.Parse(c.Param("dbId"))
	schedule, err := h.dbService.GetBackupSchedule(c.Request.Context(), dbID)
	if err != nil {
		response.Success(c, http.StatusOK, nil)
		return
	}
	response.Success(c, http.StatusOK, schedule)
}

func (h *DBHandler) SetBackupSchedule(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	dbID, _ := uuid.Parse(c.Param("dbId"))

	var req SetBackupScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	schedule, err := h.dbService.SetBackupSchedule(c.Request.Context(), dbID, orgID, req.Frequency, req.RetentionCount)
	if err != nil {
		response.InternalError(c, "Failed to set backup schedule")
		return
	}
	response.Success(c, http.StatusOK, schedule)
}
