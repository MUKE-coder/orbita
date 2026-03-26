package api

import (
	"io/fs"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Router struct {
	Engine *gin.Engine
}

func NewRouter(corsOrigins string, isProduction bool, staticFS fs.FS) *Router {
	if isProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Recovery middleware
	engine.Use(gin.Recovery())

	// Request ID middleware
	engine.Use(requestid.New())

	// Zerolog request logger
	engine.Use(zerologMiddleware())

	// CORS middleware
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": "0.1.0",
		})
	})

	// API route groups
	v1 := engine.Group("/api/v1")
	{
		v1.Group("/auth")
		v1.Group("/me")
		v1.Group("/orgs")
		v1.Group("/admin")
		v1.Group("/webhooks")
	}

	// Serve React SPA for all non-API routes
	if staticFS != nil {
		engine.NoRoute(func(c *gin.Context) {
			// Try to serve the file from the embedded FS
			path := c.Request.URL.Path
			if path == "/" {
				path = "index.html"
			} else {
				path = path[1:] // strip leading slash
			}

			file, err := staticFS.Open(path)
			if err != nil {
				// Fallback to index.html for SPA routing
				c.FileFromFS("index.html", http.FS(staticFS))
				return
			}
			file.Close()

			c.FileFromFS(path, http.FS(staticFS))
		})
	}

	return &Router{Engine: engine}
}

func zerologMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := requestid.Get(c)
		c.Next()

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
			Msg("request")
	}
}
