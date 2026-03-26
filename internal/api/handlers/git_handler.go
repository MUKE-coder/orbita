package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/middleware"
	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

type GitHandler struct {
	gitService *service.GitService
}

func NewGitHandler(gitService *service.GitService) *GitHandler {
	return &GitHandler{gitService: gitService}
}

type CreateGitConnectionRequest struct {
	Provider     string `json:"provider" binding:"required,oneof=github gitlab gitea"`
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token"`
	BaseURL      string `json:"base_url"`
}

func (h *GitHandler) ListConnections(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	conns, err := h.gitService.ListConnections(c.Request.Context(), orgID)
	if err != nil {
		response.InternalError(c, "Failed to list connections")
		return
	}
	response.Success(c, http.StatusOK, conns)
}

func (h *GitHandler) CreateConnection(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)

	var req CreateGitConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	conn, err := h.gitService.CreateConnection(c.Request.Context(), orgID, service.CreateGitConnectionInput{
		Provider:     req.Provider,
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
		BaseURL:      req.BaseURL,
	})
	if err != nil {
		response.InternalError(c, "Failed to create connection")
		return
	}

	response.Success(c, http.StatusCreated, conn)
}

func (h *GitHandler) DeleteConnection(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	connID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid connection ID")
		return
	}

	if err := h.gitService.DeleteConnection(c.Request.Context(), connID, orgID); err != nil {
		response.InternalError(c, "Failed to delete connection")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Connection deleted"})
}

func (h *GitHandler) ListRepos(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	connID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid connection ID")
		return
	}

	repos, err := h.gitService.ListRepos(c.Request.Context(), connID, orgID)
	if err != nil {
		response.InternalError(c, "Failed to list repos")
		return
	}

	response.Success(c, http.StatusOK, repos)
}

func (h *GitHandler) ListBranches(c *gin.Context) {
	orgID := middleware.GetOrgIDFromContext(c)
	connID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid connection ID")
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")

	branches, err := h.gitService.ListBranches(c.Request.Context(), connID, orgID, owner, repo)
	if err != nil {
		response.InternalError(c, "Failed to list branches")
		return
	}

	response.Success(c, http.StatusOK, branches)
}
