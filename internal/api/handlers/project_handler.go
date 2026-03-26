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

type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

type CreateProjectRequest struct {
	Name        string  `json:"name" binding:"required,min=2"`
	Description *string `json:"description"`
	Emoji       string  `json:"emoji"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Emoji       *string `json:"emoji"`
}

type CreateEnvRequest struct {
	Name string `json:"name" binding:"required,min=1"`
	Type string `json:"type" binding:"omitempty,oneof=production staging custom"`
}

type UpdateEnvRequest struct {
	Name *string `json:"name"`
}

// Projects

func (h *ProjectHandler) ListProjects(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	projects, err := h.projectService.ListProjects(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, "Failed to list projects")
		return
	}

	response.Success(c, http.StatusOK, projects)
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), orgID, req.Name, req.Description, req.Emoji)
	if err != nil {
		response.InternalError(c, "Failed to create project")
		return
	}

	response.Success(c, http.StatusCreated, project)
}

func (h *ProjectHandler) GetProject(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	project, err := h.projectService.GetProject(c.Request.Context(), projectID, orgID)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		response.InternalError(c, "Failed to get project")
		return
	}

	response.Success(c, http.StatusOK, project)
}

func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	project, err := h.projectService.UpdateProject(c.Request.Context(), projectID, orgID, req.Name, req.Description, req.Emoji)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		response.InternalError(c, "Failed to update project")
		return
	}

	response.Success(c, http.StatusOK, project)
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	if err := h.projectService.DeleteProject(c.Request.Context(), projectID, orgID); err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		response.InternalError(c, "Failed to delete project")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Project deleted"})
}

// Environments

func (h *ProjectHandler) ListEnvironments(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	envs, err := h.projectService.ListEnvironments(c.Request.Context(), projectID, orgID)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		response.InternalError(c, "Failed to list environments")
		return
	}

	response.Success(c, http.StatusOK, envs)
}

func (h *ProjectHandler) CreateEnvironment(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	var req CreateEnvRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	env, err := h.projectService.CreateEnvironment(c.Request.Context(), projectID, orgID, req.Name, req.Type)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			response.NotFound(c, "Project not found")
			return
		}
		response.InternalError(c, "Failed to create environment")
		return
	}

	response.Success(c, http.StatusCreated, env)
}

func (h *ProjectHandler) UpdateEnvironment(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	envID, err := uuid.Parse(c.Param("envId"))
	if err != nil {
		response.BadRequest(c, "Invalid environment ID")
		return
	}

	var req UpdateEnvRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	env, err := h.projectService.UpdateEnvironment(c.Request.Context(), envID, projectID, orgID, req.Name)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) || errors.Is(err, service.ErrEnvNotFound) {
			response.NotFound(c, "Environment not found")
			return
		}
		response.InternalError(c, "Failed to update environment")
		return
	}

	response.Success(c, http.StatusOK, env)
}

func (h *ProjectHandler) DeleteEnvironment(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	envID, err := uuid.Parse(c.Param("envId"))
	if err != nil {
		response.BadRequest(c, "Invalid environment ID")
		return
	}

	if err := h.projectService.DeleteEnvironment(c.Request.Context(), envID, projectID, orgID); err != nil {
		if errors.Is(err, service.ErrProjectNotFound) || errors.Is(err, service.ErrEnvNotFound) {
			response.NotFound(c, "Environment not found")
			return
		}
		response.InternalError(c, "Failed to delete environment")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Environment deleted"})
}
