package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/response"
	"github.com/orbita-sh/orbita/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=2"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, tokens, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			response.Conflict(c, "Email already registered")
			return
		}
		response.InternalError(c, "Failed to register")
		return
	}

	setRefreshTokenCookie(c, tokens.RefreshToken)

	response.Success(c, http.StatusCreated, gin.H{
		"user":         user,
		"access_token": tokens.AccessToken,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	deviceInfo := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	user, tokens, err := h.authService.Login(c.Request.Context(), req.Email, req.Password, deviceInfo, ip)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(c, "Invalid email or password")
			return
		}
		response.InternalError(c, "Failed to login")
		return
	}

	setRefreshTokenCookie(c, tokens.RefreshToken)

	response.Success(c, http.StatusOK, gin.H{
		"user":         user,
		"access_token": tokens.AccessToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		response.Success(c, http.StatusOK, gin.H{"message": "logged out"})
		return
	}

	claims, err := auth.ValidateRefreshToken(refreshToken, h.authService.GetConfig().JWTRefreshSecret)
	if err == nil {
		_ = h.authService.Logout(c.Request.Context(), claims.SessionID)
	}

	clearRefreshTokenCookie(c)

	response.Success(c, http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		response.Unauthorized(c, "No refresh token")
		return
	}

	user, tokens, err := h.authService.RefreshTokens(c.Request.Context(), refreshToken)
	if err != nil {
		clearRefreshTokenCookie(c)
		if errors.Is(err, service.ErrInvalidToken) || errors.Is(err, service.ErrSessionNotFound) || errors.Is(err, service.ErrSessionExpired) {
			response.Unauthorized(c, "Invalid or expired refresh token")
			return
		}
		response.InternalError(c, "Failed to refresh tokens")
		return
	}

	setRefreshTokenCookie(c, tokens.RefreshToken)

	response.Success(c, http.StatusOK, gin.H{
		"user":         user,
		"access_token": tokens.AccessToken,
	})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	_ = h.authService.ForgotPassword(c.Request.Context(), req.Email)

	response.Success(c, http.StatusOK, gin.H{
		"message": "If that email exists, a reset code has been sent",
	})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), req.Email, req.OTP, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidOTP) || errors.Is(err, service.ErrUserNotFound) {
			response.BadRequest(c, "Invalid or expired reset code")
			return
		}
		response.InternalError(c, "Failed to reset password")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		token := c.Query("token")
		if token == "" {
			response.BadRequest(c, "Token is required")
			return
		}
		req.Token = token
	}

	if err := h.authService.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			response.BadRequest(c, "Invalid or expired verification link")
			return
		}
		response.InternalError(c, "Failed to verify email")
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

func setRefreshTokenCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		token,
		int((30 * 24 * time.Hour).Seconds()),
		"/",
		"",
		false,
		true,
	)
}

func clearRefreshTokenCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)
}
