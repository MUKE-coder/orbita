package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("CreateUser: %w", err)
	}
	return nil
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("FindUserByEmail: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, fmt.Errorf("FindUserByID: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("UpdateUser: %w", err)
	}
	return nil
}

func (r *UserRepository) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("CountUsers: %w", err)
	}
	return count, nil
}

// Sessions

func (r *UserRepository) CreateSession(ctx context.Context, session *models.Session) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return fmt.Errorf("CreateSession: %w", err)
	}
	return nil
}

func (r *UserRepository) FindSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	var session models.Session
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error; err != nil {
		return nil, fmt.Errorf("FindSessionByID: %w", err)
	}
	return &session, nil
}

func (r *UserRepository) FindSessionByTokenHash(ctx context.Context, hash string) (*models.Session, error) {
	var session models.Session
	if err := r.db.WithContext(ctx).Where("refresh_token_hash = ?", hash).First(&session).Error; err != nil {
		return nil, fmt.Errorf("FindSessionByTokenHash: %w", err)
	}
	return &session, nil
}

func (r *UserRepository) FindSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	var sessions []models.Session
	if err := r.db.WithContext(ctx).Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("FindSessionsByUserID: %w", err)
	}
	return sessions, nil
}

func (r *UserRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Session{}).Error; err != nil {
		return fmt.Errorf("DeleteSession: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.Session{}).Error; err != nil {
		return fmt.Errorf("DeleteUserSessions: %w", err)
	}
	return nil
}

// Email Verification

func (r *UserRepository) CreateEmailVerification(ctx context.Context, ev *models.EmailVerification) error {
	if err := r.db.WithContext(ctx).Create(ev).Error; err != nil {
		return fmt.Errorf("CreateEmailVerification: %w", err)
	}
	return nil
}

func (r *UserRepository) FindEmailVerificationByTokenHash(ctx context.Context, hash string) (*models.EmailVerification, error) {
	var ev models.EmailVerification
	if err := r.db.WithContext(ctx).Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", hash, time.Now()).
		First(&ev).Error; err != nil {
		return nil, fmt.Errorf("FindEmailVerificationByTokenHash: %w", err)
	}
	return &ev, nil
}

func (r *UserRepository) MarkEmailVerificationUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&models.EmailVerification{}).Where("id = ?", id).
		Update("used_at", &now).Error; err != nil {
		return fmt.Errorf("MarkEmailVerificationUsed: %w", err)
	}
	return nil
}

// Password Resets

func (r *UserRepository) CreatePasswordReset(ctx context.Context, pr *models.PasswordReset) error {
	if err := r.db.WithContext(ctx).Create(pr).Error; err != nil {
		return fmt.Errorf("CreatePasswordReset: %w", err)
	}
	return nil
}

func (r *UserRepository) FindPasswordResetByUserID(ctx context.Context, userID uuid.UUID) (*models.PasswordReset, error) {
	var pr models.PasswordReset
	if err := r.db.WithContext(ctx).Where("user_id = ? AND used_at IS NULL AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").First(&pr).Error; err != nil {
		return nil, fmt.Errorf("FindPasswordResetByUserID: %w", err)
	}
	return &pr, nil
}

func (r *UserRepository) MarkPasswordResetUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&models.PasswordReset{}).Where("id = ?", id).
		Update("used_at", &now).Error; err != nil {
		return fmt.Errorf("MarkPasswordResetUsed: %w", err)
	}
	return nil
}

// API Keys

func (r *UserRepository) CreateAPIKey(ctx context.Context, key *models.APIKey) error {
	if err := r.db.WithContext(ctx).Create(key).Error; err != nil {
		return fmt.Errorf("CreateAPIKey: %w", err)
	}
	return nil
}

func (r *UserRepository) FindAPIKeyByHash(ctx context.Context, hash string) (*models.APIKey, error) {
	var key models.APIKey
	if err := r.db.WithContext(ctx).Where("key_hash = ?", hash).First(&key).Error; err != nil {
		return nil, fmt.Errorf("FindAPIKeyByHash: %w", err)
	}
	return &key, nil
}

func (r *UserRepository) FindAPIKeysByUserID(ctx context.Context, userID uuid.UUID) ([]models.APIKey, error) {
	var keys []models.APIKey
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("FindAPIKeysByUserID: %w", err)
	}
	return keys, nil
}

func (r *UserRepository) DeleteAPIKey(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.APIKey{}).Error; err != nil {
		return fmt.Errorf("DeleteAPIKey: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&models.APIKey{}).Where("id = ?", id).
		Update("last_used_at", &now).Error; err != nil {
		return fmt.Errorf("UpdateAPIKeyLastUsed: %w", err)
	}
	return nil
}
