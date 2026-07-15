package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

type WebhookHandler struct {
	appService *service.AppService
	gitService *service.GitService
	orgService *service.OrgService
}

func NewWebhookHandler(appService *service.AppService, gitService *service.GitService, orgService *service.OrgService) *WebhookHandler {
	return &WebhookHandler{
		appService: appService,
		gitService: gitService,
		orgService: orgService,
	}
}

type GitHubPushEvent struct {
	Ref        string `json:"ref"`
	Repository struct {
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
	HeadCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	} `json:"head_commit"`
}

func (h *WebhookHandler) HandleGitHub(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "Failed to read body")
		return
	}

	eventType := c.GetHeader("X-GitHub-Event")
	if eventType != "push" {
		response.Success(c, http.StatusOK, gin.H{"message": "event ignored"})
		return
	}

	var event GitHubPushEvent
	if err := json.Unmarshal(body, &event); err != nil {
		response.BadRequest(c, "Invalid payload")
		return
	}

	// Extract branch from ref (refs/heads/main -> main)
	branch := ""
	if len(event.Ref) > 11 {
		branch = event.Ref[11:]
	}

	log.Info().
		Str("repo", event.Repository.FullName).
		Str("branch", branch).
		Str("commit", event.HeadCommit.ID).
		Msg("Received GitHub push webhook")

	// Find app configured for auto-deploy on this repo+branch
	app, err := h.gitService.FindAppByRepoAndBranch(c.Request.Context(), event.Repository.CloneURL, event.Repository.FullName, branch)
	if err != nil {
		// No app found for this repo+branch, ignore
		response.Success(c, http.StatusOK, gin.H{"message": "no matching app"})
		return
	}

	// Signature is mandatory: unsigned deliveries never trigger a deploy. Apps
	// created before webhook secrets became automatic must regenerate one via
	// POST /apps/:appId/webhook-secret.
	if app.WebhookSecret == nil || *app.WebhookSecret == "" {
		log.Warn().Str("app_id", app.ID.String()).Msg("Webhook rejected: app has no webhook secret configured")
		response.Unauthorized(c, "App has no webhook secret; regenerate one and configure it on the repository webhook")
		return
	}
	signature := c.GetHeader("X-Hub-Signature-256")
	if !verifyGitHubSignature(body, signature, *app.WebhookSecret) {
		response.Unauthorized(c, "Invalid webhook signature")
		return
	}

	// Resolve the org slug — it names the image tag, network, and cgroup.
	org, err := h.orgService.GetOrganizationByID(c.Request.Context(), app.OrganizationID)
	if err != nil {
		log.Error().Err(err).Str("app_id", app.ID.String()).Msg("Webhook deploy: org lookup failed")
		response.InternalError(c, "Deploy failed")
		return
	}

	// Trigger deploy
	deployment, err := h.appService.DeployWithTrigger(c.Request.Context(), app.ID, app.OrganizationID, org.Slug, nil, models.TriggerWebhook)
	if err != nil {
		log.Error().Err(err).Str("app_id", app.ID.String()).Msg("Webhook deploy failed")
		response.InternalError(c, "Deploy failed")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message":    "Deploy triggered",
		"deployment": deployment,
	})
}

func (h *WebhookHandler) HandleGitLab(c *gin.Context) {
	// TODO: parse GitLab webhook payload and trigger deploy
	log.Info().Msg("Received GitLab webhook")
	response.Success(c, http.StatusOK, gin.H{"message": "GitLab webhook received"})
}

func (h *WebhookHandler) HandleGitea(c *gin.Context) {
	// TODO: parse Gitea webhook payload and trigger deploy
	log.Info().Msg("Received Gitea webhook")
	response.Success(c, http.StatusOK, gin.H{"message": "Gitea webhook received"})
}

func verifyGitHubSignature(payload []byte, signature, secret string) bool {
	if len(signature) < 7 {
		return false
	}
	sig := signature[7:] // Remove "sha256=" prefix
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}
