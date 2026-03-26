package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

type NotificationService struct {
	notifRepo *repository.NotificationRepository
}

func NewNotificationService(notifRepo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifRepo: notifRepo}
}

func (s *NotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID, notifType, title, body string) error {
	n := &models.Notification{
		ID:             uuid.New(),
		UserID:         userID,
		OrganizationID: orgID,
		Type:           notifType,
		Title:          title,
		Body:           &body,
	}
	return s.notifRepo.Create(ctx, n)
}

func (s *NotificationService) ListNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, int64, error) {
	notifs, err := s.notifRepo.ListByUserID(ctx, userID, limit)
	if err != nil {
		return nil, 0, err
	}
	unread, _ := s.notifRepo.CountUnread(ctx, userID)
	return notifs, unread, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return s.notifRepo.MarkRead(ctx, id, userID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.notifRepo.MarkAllRead(ctx, userID)
}

func (s *NotificationService) GetSettings(ctx context.Context, orgID, userID uuid.UUID) ([]models.NotificationSetting, error) {
	return s.notifRepo.GetSettings(ctx, orgID, userID)
}

func (s *NotificationService) UpdateSettings(ctx context.Context, orgID, userID uuid.UUID, eventType string, emailEnabled, webhookEnabled bool, webhookURL *string) error {
	setting := &models.NotificationSetting{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		EventType:      eventType,
		EmailEnabled:   emailEnabled,
		WebhookEnabled: webhookEnabled,
		WebhookURL:     webhookURL,
	}
	return s.notifRepo.UpsertSetting(ctx, setting)
}

// Audit logging

func (s *NotificationService) LogAudit(ctx context.Context, orgID *uuid.UUID, userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, metadata interface{}, ip string) error {
	meta, _ := json.Marshal(metadata)
	var resType *string
	if resourceType != "" {
		resType = &resourceType
	}
	var ipAddr *string
	if ip != "" {
		ipAddr = &ip
	}

	log := &models.AuditLog{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Action:         action,
		ResourceType:   resType,
		ResourceID:     resourceID,
		Metadata:       meta,
		IP:             ipAddr,
	}
	return s.notifRepo.CreateAuditLog(ctx, log)
}

func (s *NotificationService) ListAuditLogs(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]models.AuditLog, int64, error) {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	return s.notifRepo.ListAuditLogs(ctx, orgID, pageSize, offset)
}

func (s *NotificationService) SendWebhookNotification(ctx context.Context, webhookURL string, payload interface{}) error {
	// TODO: real impl — POST JSON to webhookURL
	_ = webhookURL
	_ = payload
	return nil
}

func (s *NotificationService) SendDeployNotification(ctx context.Context, orgID uuid.UUID, appName, status string, memberUserIDs []uuid.UUID) {
	title := fmt.Sprintf("Deploy %s: %s", status, appName)
	body := fmt.Sprintf("Application %s deployment %s", appName, status)
	for _, uid := range memberUserIDs {
		_ = s.CreateNotification(ctx, uid, &orgID, models.NotifTypeDeploy, title, body)
	}
}
