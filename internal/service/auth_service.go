package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/config"
	"github.com/orbita-sh/orbita/internal/mailer"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrInvalidOTP         = errors.New("invalid or expired OTP")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
)

type AuthService struct {
	userRepo *repository.UserRepository
	mailer   *mailer.Mailer
	cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, mailer *mailer.Mailer, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		mailer:   mailer,
		cfg:      cfg,
	}
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*models.User, *AuthTokens, error) {
	// Check if user already exists
	existing, err := s.userRepo.FindUserByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, nil, ErrEmailAlreadyExists
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, nil, fmt.Errorf("Register: %w", err)
	}

	// Check if this is the first user (make super admin)
	count, _ := s.userRepo.CountUsers(ctx)
	isSuperAdmin := count == 0

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		IsSuperAdmin: isSuperAdmin,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("Register: %w", err)
	}

	// Generate email verification token
	token, err := auth.GenerateRandomToken(32)
	if err != nil {
		return nil, nil, fmt.Errorf("Register: generate verify token: %w", err)
	}

	ev := &models.EmailVerification{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: auth.HashToken(token),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := s.userRepo.CreateEmailVerification(ctx, ev); err != nil {
		return nil, nil, fmt.Errorf("Register: save verification: %w", err)
	}

	// Send verification email (non-blocking)
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.cfg.AppBaseURL, token)
	go func() {
		_ = s.mailer.SendEmailVerification(context.Background(), email, name, verifyURL)
	}()

	// Generate tokens
	tokens, err := s.generateTokens(ctx, user, "", "")
	if err != nil {
		return nil, nil, fmt.Errorf("Register: %w", err)
	}

	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, email, password, deviceInfo, ip string) (*models.User, *AuthTokens, error) {
	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("Login: %w", err)
	}

	if !auth.CheckPassword(password, user.PasswordHash) {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.generateTokens(ctx, user, deviceInfo, ip)
	if err != nil {
		return nil, nil, fmt.Errorf("Login: %w", err)
	}

	return user, tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.userRepo.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("Logout: %w", err)
	}
	return nil
}

func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*models.User, *AuthTokens, error) {
	claims, err := auth.ValidateRefreshToken(refreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}

	session, err := s.userRepo.FindSessionByID(ctx, claims.SessionID)
	if err != nil {
		return nil, nil, ErrSessionNotFound
	}

	if time.Now().After(session.ExpiresAt) {
		_ = s.userRepo.DeleteSession(ctx, session.ID)
		return nil, nil, ErrSessionExpired
	}

	// Verify the refresh token hash matches
	tokenHash := auth.HashToken(refreshToken)
	if tokenHash != session.RefreshTokenHash {
		return nil, nil, ErrInvalidToken
	}

	user, err := s.userRepo.FindUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	// Delete old session
	_ = s.userRepo.DeleteSession(ctx, session.ID)

	// Generate new tokens with new session
	deviceInfo := ""
	ip := ""
	if session.DeviceInfo != nil {
		deviceInfo = *session.DeviceInfo
	}
	if session.IPAddress != nil {
		ip = *session.IPAddress
	}

	tokens, err := s.generateTokens(ctx, user, deviceInfo, ip)
	if err != nil {
		return nil, nil, fmt.Errorf("RefreshTokens: %w", err)
	}

	return user, tokens, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal whether email exists
		return nil
	}

	otp, err := auth.GenerateOTP()
	if err != nil {
		return fmt.Errorf("ForgotPassword: %w", err)
	}

	otpHash, err := auth.HashPassword(otp)
	if err != nil {
		return fmt.Errorf("ForgotPassword: hash OTP: %w", err)
	}

	pr := &models.PasswordReset{
		ID:        uuid.New(),
		UserID:    user.ID,
		OTPHash:   otpHash,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := s.userRepo.CreatePasswordReset(ctx, pr); err != nil {
		return fmt.Errorf("ForgotPassword: %w", err)
	}

	go func() {
		_ = s.mailer.SendPasswordReset(context.Background(), email, user.Name, otp)
	}()

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		return ErrUserNotFound
	}

	pr, err := s.userRepo.FindPasswordResetByUserID(ctx, user.ID)
	if err != nil {
		return ErrInvalidOTP
	}

	if !auth.CheckPassword(otp, pr.OTPHash) {
		return ErrInvalidOTP
	}

	passwordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("ResetPassword: %w", err)
	}

	user.PasswordHash = passwordHash
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("ResetPassword: %w", err)
	}

	if err := s.userRepo.MarkPasswordResetUsed(ctx, pr.ID); err != nil {
		return fmt.Errorf("ResetPassword: mark used: %w", err)
	}

	// Invalidate all sessions
	_ = s.userRepo.DeleteUserSessions(ctx, user.ID)

	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	tokenHash := auth.HashToken(token)
	ev, err := s.userRepo.FindEmailVerificationByTokenHash(ctx, tokenHash)
	if err != nil {
		return ErrInvalidToken
	}

	user, err := s.userRepo.FindUserByID(ctx, ev.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	user.IsEmailVerified = true
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("VerifyEmail: %w", err)
	}

	if err := s.userRepo.MarkEmailVerificationUsed(ctx, ev.ID); err != nil {
		return fmt.Errorf("VerifyEmail: mark used: %w", err)
	}

	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetProfile: %w", err)
	}
	return user, nil
}

func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, name string, avatarURL *string) (*models.User, error) {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("UpdateProfile: %w", err)
	}

	if name != "" {
		user.Name = name
	}
	if avatarURL != nil {
		user.AvatarURL = avatarURL
	}

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("UpdateProfile: %w", err)
	}
	return user, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("ChangePassword: %w", err)
	}

	if !auth.CheckPassword(currentPassword, user.PasswordHash) {
		return ErrInvalidCredentials
	}

	passwordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("ChangePassword: %w", err)
	}

	user.PasswordHash = passwordHash
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("ChangePassword: %w", err)
	}

	return nil
}

func (s *AuthService) GetSessions(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	return s.userRepo.FindSessionsByUserID(ctx, userID)
}

func (s *AuthService) RevokeSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	session, err := s.userRepo.FindSessionByID(ctx, sessionID)
	if err != nil {
		return ErrSessionNotFound
	}
	if session.UserID != userID {
		return ErrSessionNotFound
	}
	return s.userRepo.DeleteSession(ctx, sessionID)
}

// API Keys

func (s *AuthService) CreateAPIKey(ctx context.Context, userID uuid.UUID, name string, scopes []string) (*models.APIKey, string, error) {
	rawKey := "orb_" + uuid.New().String()
	keyHash := auth.HashToken(rawKey)
	keyPrefix := rawKey[:12]

	apiKey := &models.APIKey{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		Scopes:    scopes,
	}

	if err := s.userRepo.CreateAPIKey(ctx, apiKey); err != nil {
		return nil, "", fmt.Errorf("CreateAPIKey: %w", err)
	}

	return apiKey, rawKey, nil
}

func (s *AuthService) ListAPIKeys(ctx context.Context, userID uuid.UUID) ([]models.APIKey, error) {
	return s.userRepo.FindAPIKeysByUserID(ctx, userID)
}

func (s *AuthService) DeleteAPIKey(ctx context.Context, keyID, userID uuid.UUID) error {
	return s.userRepo.DeleteAPIKey(ctx, keyID, userID)
}

func (s *AuthService) GetConfig() *config.Config {
	return s.cfg
}

func (s *AuthService) generateTokens(ctx context.Context, user *models.User, deviceInfo, ip string) (*AuthTokens, error) {
	sessionID := uuid.New()

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.IsSuperAdmin, s.cfg.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("generateTokens: access: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, sessionID, s.cfg.JWTRefreshSecret)
	if err != nil {
		return nil, fmt.Errorf("generateTokens: refresh: %w", err)
	}

	var devInfo, ipAddr *string
	if deviceInfo != "" {
		devInfo = &deviceInfo
	}
	if ip != "" {
		ipAddr = &ip
	}

	session := &models.Session{
		ID:               sessionID,
		UserID:           user.ID,
		RefreshTokenHash: auth.HashToken(refreshToken),
		DeviceInfo:       devInfo,
		IPAddress:        ipAddr,
		ExpiresAt:        time.Now().Add(auth.RefreshTokenTTL),
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("generateTokens: create session: %w", err)
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
