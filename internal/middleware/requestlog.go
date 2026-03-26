package middleware

import (
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// StructuredLogger provides production-ready request logging with zerolog.
// Includes: request ID, org ID, user ID, method, path, status, latency, IP.
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := requestid.Get(c)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		event := log.Info()
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		}

		event.
			Str("request_id", reqID).
			Int("status", status).
			Str("method", method).
			Str("path", path).
			Str("ip", c.ClientIP()).
			Dur("latency", latency).
			Int("size", c.Writer.Size())

		// Add user context if available
		if userID, exists := c.Get("user_id"); exists {
			event.Str("user_id", userID.(interface{ String() string }).String())
		}
		if orgID, exists := c.Get("org_id"); exists {
			event.Str("org_id", orgID.(interface{ String() string }).String())
		}

		event.Msg("request")
	}
}
