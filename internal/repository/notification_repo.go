package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Notifications

func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	if err := r.db.WithContext(ctx).Create(n).Error; err != nil {
		return fmt.Errorf("NotifRepo.Create: %w", err)
	}
	return nil
}

func (r *NotificationRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error) {
	var notifs []models.Notification
	q := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&notifs).Error; err != nil {
		return nil, fmt.Errorf("NotifRepo.ListByUserID: %w", err)
	}
	return notifs, nil
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read = false", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("NotifRepo.CountUnread: %w", err)
	}
	return count, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).Update("read", true).Error; err != nil {
		return fmt.Errorf("NotifRepo.MarkRead: %w", err)
	}
	return nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read = false", userID).Update("read", true).Error; err != nil {
		return fmt.Errorf("NotifRepo.MarkAllRead: %w", err)
	}
	return nil
}

// Audit Logs

func (r *NotificationRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("NotifRepo.CreateAuditLog: %w", err)
	}
	return nil
}

func (r *NotificationRepository) ListAuditLogs(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	q := r.db.WithContext(ctx).Model(&models.AuditLog{}).Where("organization_id = ?", orgID)
	q.Count(&total)

	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("NotifRepo.ListAuditLogs: %w", err)
	}
	return logs, total, nil
}

// Notification Settings

func (r *NotificationRepository) GetSettings(ctx context.Context, orgID, userID uuid.UUID) ([]models.NotificationSetting, error) {
	var settings []models.NotificationSetting
	if err := r.db.WithContext(ctx).Where("organization_id = ? AND user_id = ?", orgID, userID).
		Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("NotifRepo.GetSettings: %w", err)
	}
	return settings, nil
}

func (r *NotificationRepository) UpsertSetting(ctx context.Context, setting *models.NotificationSetting) error {
	if err := r.db.WithContext(ctx).Save(setting).Error; err != nil {
		return fmt.Errorf("NotifRepo.UpsertSetting: %w", err)
	}
	return nil
}
