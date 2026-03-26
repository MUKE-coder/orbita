package api

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/api/handlers"
	"github.com/orbita-sh/orbita/internal/config"
	mw "github.com/orbita-sh/orbita/internal/middleware"
	"github.com/orbita-sh/orbita/internal/repository"
	"github.com/orbita-sh/orbita/internal/service"
)

type Router struct {
	Engine *gin.Engine
}

type RouterDeps struct {
	Config      *config.Config
	AuthService *service.AuthService
	UserRepo    *repository.UserRepository
	Redis       *redis.Client
	StaticFS    fs.FS
}

func NewRouter(deps *RouterDeps) *Router {
	if deps.Config.IsProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	engine.Use(gin.Recovery())
	engine.Use(requestid.New())
	engine.Use(zerologMiddleware())

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{deps.Config.CORSOrigins},
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

	// Handlers
	authHandler := handlers.NewAuthHandler(deps.AuthService)
	meHandler := handlers.NewMeHandler(deps.AuthService)

	// Rate limiter for auth endpoints
	authRateLimit := mw.RateLimit(deps.Redis, 5, 15*time.Minute)

	v1 := engine.Group("/api/v1")
	{
		// Auth routes (public)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authRateLimit, authHandler.Register)
			authGroup.POST("/login", authRateLimit, authHandler.Login)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.POST("/refresh", authHandler.Refresh)
			authGroup.POST("/forgot-password", authRateLimit, authHandler.ForgotPassword)
			authGroup.POST("/reset-password", authRateLimit, authHandler.ResetPassword)
			authGroup.POST("/verify-email", authHandler.VerifyEmail)
		}

		// Me routes (authenticated)
		meGroup := v1.Group("/me", mw.RequireAuth(deps.Config.JWTSecret))
		{
			meGroup.GET("", meHandler.GetProfile)
			meGroup.PUT("", meHandler.UpdateProfile)
			meGroup.POST("/change-password", meHandler.ChangePassword)
			meGroup.GET("/sessions", meHandler.GetSessions)
			meGroup.DELETE("/sessions/:id", meHandler.RevokeSession)
			meGroup.GET("/api-keys", meHandler.ListAPIKeys)
			meGroup.POST("/api-keys", meHandler.CreateAPIKey)
			meGroup.DELETE("/api-keys/:id", meHandler.DeleteAPIKey)
		}

		// Placeholder groups for future phases
		v1.Group("/orgs")
		v1.Group("/admin")
		v1.Group("/webhooks")
	}

	// Serve React SPA for all non-API routes
	if deps.StaticFS != nil {
		engine.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if path == "/" {
				path = "index.html"
			} else {
				path = path[1:]
			}

			file, err := deps.StaticFS.Open(path)
			if err != nil {
				c.FileFromFS("index.html", http.FS(deps.StaticFS))
				return
			}
			file.Close()

			c.FileFromFS(path, http.FS(deps.StaticFS))
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
