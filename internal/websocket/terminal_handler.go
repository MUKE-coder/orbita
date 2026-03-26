package websocket

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/auth"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: restrict to allowed origins in production
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type TerminalHandler struct {
	jwtSecret string
}

func NewTerminalHandler(jwtSecret string) *TerminalHandler {
	return &TerminalHandler{jwtSecret: jwtSecret}
}

func (h *TerminalHandler) HandleTerminal(c *gin.Context) {
	// Authenticate via query parameter token
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	claims, err := auth.ValidateAccessToken(token, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	appID := c.Param("appId")

	log.Info().
		Str("user_id", claims.UserID.String()).
		Str("app_id", appID).
		Msg("Terminal session opened")

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket")
		return
	}
	defer conn.Close()

	// TODO: real impl
	// 1. Find running container ID for app (via Docker API task lookup)
	// 2. docker exec -it {containerID} /bin/sh (or /bin/bash if available)
	// 3. Pipe stdin/stdout/stderr between WS and exec stream
	// 4. Handle resize messages: {type:"resize",cols:N,rows:M}
	// 5. Log in audit trail: "User X opened terminal in app Y"
	// 6. Force close if user loses membership

	// Stub: echo back messages
	_ = conn.WriteMessage(websocket.TextMessage, []byte(
		fmt.Sprintf("Connected to terminal for app %s\r\n", appID)+
			"$ ",
	))

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				break
			}
			log.Error().Err(err).Msg("Terminal read error")
			break
		}

		if messageType == websocket.TextMessage {
			input := string(msg)

			// Handle resize messages
			if len(input) > 0 && input[0] == '{' {
				// JSON message (resize, etc.) — ignore for stub
				continue
			}

			// Echo back with simulated response
			response := fmt.Sprintf("%s\r\n$ ", input)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
				break
			}
		}
	}

	log.Info().
		Str("user_id", claims.UserID.String()).
		Str("app_id", appID).
		Msg("Terminal session closed")
}

func (h *TerminalHandler) HandleLogStream(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	_, err := auth.ValidateAccessToken(token, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket for logs")
		return
	}
	defer conn.Close()

	appID := c.Param("appId")

	// TODO: real impl — subscribe to Docker service logs and fan out
	// Stub: send periodic mock log messages
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			msg := fmt.Sprintf("[%s] [app:%s] Application running normally\n", t.Format("15:04:05"), appID)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return
			}
		}
	}
}
