package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/docker"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"

	"github.com/docker/docker/pkg/stdcopy"
)

// upgrader restricts origins to same-host plus the configured allowed origins.
func newUpgrader(allowedOrigins []string) websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true // non-browser client
			}
			if strings.Contains(origin, r.Host) {
				return true
			}
			for _, o := range allowedOrigins {
				if o != "" && strings.EqualFold(strings.TrimRight(o, "/"), strings.TrimRight(origin, "/")) {
					return true
				}
			}
			return false
		},
	}
}

type TerminalHandler struct {
	jwtSecret    string
	appRepo      *repository.AppRepository
	orgRepo      *repository.OrgRepository
	dockerClient *docker.Client
	upgrader     websocket.Upgrader
}

func NewTerminalHandler(jwtSecret string, appRepo *repository.AppRepository, orgRepo *repository.OrgRepository, dockerClient *docker.Client, allowedOrigins []string) *TerminalHandler {
	return &TerminalHandler{
		jwtSecret:    jwtSecret,
		appRepo:      appRepo,
		orgRepo:      orgRepo,
		dockerClient: dockerClient,
		upgrader:     newUpgrader(allowedOrigins),
	}
}

// authorizeWS validates the query token and the caller's role on the app's
// org, returning the app when authorized. Browsers cannot set headers on
// WebSocket handshakes, so auth happens here instead of middleware.
func (h *TerminalHandler) authorizeWS(c *gin.Context, minRole string) (*models.Application, bool) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return nil, false
	}
	claims, err := auth.ValidateAccessToken(token, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return nil, false
	}

	appID, err := uuid.Parse(c.Param("appId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid app ID"})
		return nil, false
	}

	orgSlug := c.Param("orgSlug")
	org, err := h.orgRepo.FindOrgBySlug(c.Request.Context(), orgSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return nil, false
	}

	if !claims.IsSuperAdmin {
		member, err := h.orgRepo.GetMember(c.Request.Context(), org.ID, claims.UserID)
		if err != nil || !models.HasMinRole(member.Role, minRole) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return nil, false
		}
	}

	app, err := h.appRepo.FindByID(c.Request.Context(), appID, org.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "App not found"})
		return nil, false
	}
	return app, true
}

type resizeMsg struct {
	Type string `json:"type"`
	Cols uint   `json:"cols"`
	Rows uint   `json:"rows"`
}

// HandleTerminal attaches an interactive shell in the app's container to the
// WebSocket. Requires developer+ role.
func (h *TerminalHandler) HandleTerminal(c *gin.Context) {
	app, ok := h.authorizeWS(c, models.RoleDeveloper)
	if !ok {
		return
	}
	if app.DockerServiceID == nil || *app.DockerServiceID == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "App has no running service"})
		return
	}

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	containerID, err := h.dockerClient.FindContainerIDForService(ctx, *app.DockerServiceID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "No running container for app"})
		return
	}

	attach, execID, err := h.dockerClient.ExecInteractiveShell(ctx, containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start shell"})
		return
	}
	defer attach.Close()

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket")
		return
	}
	defer conn.Close()

	log.Info().Str("app", app.Name).Str("container", containerID[:12]).Msg("Terminal session opened")

	// container -> websocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := attach.Reader.Read(buf)
			if n > 0 {
				if werr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					break
				}
			}
			if err != nil {
				break
			}
		}
		cancel()
		_ = conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shell exited"),
			time.Now().Add(time.Second))
	}()

	// websocket -> container (with resize control messages)
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if msgType == websocket.TextMessage && len(msg) > 0 && msg[0] == '{' {
			var rm resizeMsg
			if json.Unmarshal(msg, &rm) == nil && rm.Type == "resize" && rm.Cols > 0 && rm.Rows > 0 {
				_ = h.dockerClient.ResizeExec(ctx, execID, rm.Rows, rm.Cols)
				continue
			}
		}
		if _, err := attach.Conn.Write(msg); err != nil {
			break
		}
	}

	log.Info().Str("app", app.Name).Msg("Terminal session closed")
}

// HandleLogStream follows the app's service logs over the WebSocket. Requires
// viewer+ role.
func (h *TerminalHandler) HandleLogStream(c *gin.Context) {
	app, ok := h.authorizeWS(c, models.RoleViewer)
	if !ok {
		return
	}
	if app.DockerServiceID == nil || *app.DockerServiceID == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "App has no running service"})
		return
	}

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	reader, err := h.dockerClient.FollowServiceLogs(ctx, *app.DockerServiceID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open log stream"})
		return
	}
	defer reader.Close()

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket for logs")
		return
	}
	defer conn.Close()

	// Drain client messages so pings/close frames are processed
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	// Demux the docker log stream into the websocket as text lines
	w := &wsLogWriter{conn: conn}
	_, _ = stdcopy.StdCopy(w, w, reader)

	_ = conn.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "log stream ended"),
		time.Now().Add(time.Second))
}

// wsLogWriter adapts websocket writes to io.Writer for stdcopy demuxing.
type wsLogWriter struct {
	conn *websocket.Conn
}

func (w *wsLogWriter) Write(p []byte) (int, error) {
	if err := w.conn.WriteMessage(websocket.TextMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}
