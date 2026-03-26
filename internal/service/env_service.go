package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

type EnvService struct {
	envRepo       *repository.EnvRepository
	encryptionKey []byte
}

func NewEnvService(envRepo *repository.EnvRepository, encryptionKey []byte) *EnvService {
	return &EnvService{
		envRepo:       envRepo,
		encryptionKey: encryptionKey,
	}
}

func (s *EnvService) SetEnvVar(ctx context.Context, resourceID uuid.UUID, resourceType, key, value string, isSecret bool, orgID uuid.UUID) error {
	orgKey, err := auth.DeriveOrgKey(s.encryptionKey, orgID)
	if err != nil {
		return fmt.Errorf("SetEnvVar: derive key: %w", err)
	}

	encrypted, err := auth.Encrypt(value, orgKey)
	if err != nil {
		return fmt.Errorf("SetEnvVar: encrypt: %w", err)
	}

	ev := &models.EnvVariable{
		ID:             uuid.New(),
		ResourceID:     resourceID,
		ResourceType:   resourceType,
		OrganizationID: orgID,
		Key:            key,
		ValueEncrypted: encrypted,
		IsSecret:       isSecret,
	}

	return s.envRepo.Upsert(ctx, ev)
}

func (s *EnvService) GetEnvVars(ctx context.Context, resourceID uuid.UUID, resourceType string, orgID uuid.UUID) ([]models.EnvVarDisplay, error) {
	vars, err := s.envRepo.ListByResource(ctx, resourceID, resourceType)
	if err != nil {
		return nil, fmt.Errorf("GetEnvVars: %w", err)
	}

	orgKey, err := auth.DeriveOrgKey(s.encryptionKey, orgID)
	if err != nil {
		return nil, fmt.Errorf("GetEnvVars: derive key: %w", err)
	}

	result := make([]models.EnvVarDisplay, 0, len(vars))
	for _, v := range vars {
		display := models.EnvVarDisplay{
			ID:           v.ID,
			Key:          v.Key,
			IsSecret:     v.IsSecret,
			ResourceID:   v.ResourceID,
			ResourceType: v.ResourceType,
			CreatedAt:    v.CreatedAt,
			UpdatedAt:    v.UpdatedAt,
		}

		if v.IsSecret {
			display.Value = "••••••••"
		} else {
			decrypted, err := auth.Decrypt(v.ValueEncrypted, orgKey)
			if err != nil {
				display.Value = "[decryption error]"
			} else {
				display.Value = decrypted
			}
		}

		result = append(result, display)
	}

	return result, nil
}

func (s *EnvService) GetEnvVarMap(ctx context.Context, resourceID uuid.UUID, resourceType string, orgID uuid.UUID) (map[string]string, error) {
	vars, err := s.envRepo.ListByResource(ctx, resourceID, resourceType)
	if err != nil {
		return nil, fmt.Errorf("GetEnvVarMap: %w", err)
	}

	orgKey, err := auth.DeriveOrgKey(s.encryptionKey, orgID)
	if err != nil {
		return nil, fmt.Errorf("GetEnvVarMap: derive key: %w", err)
	}

	result := make(map[string]string, len(vars))
	for _, v := range vars {
		decrypted, err := auth.Decrypt(v.ValueEncrypted, orgKey)
		if err != nil {
			continue
		}
		result[v.Key] = decrypted
	}

	return result, nil
}

func (s *EnvService) DeleteEnvVar(ctx context.Context, envVarID, orgID uuid.UUID) error {
	return s.envRepo.Delete(ctx, envVarID, orgID)
}

func (s *EnvService) BulkSetEnvVars(ctx context.Context, resourceID uuid.UUID, resourceType string, envVars map[string]string, orgID uuid.UUID) error {
	for key, value := range envVars {
		if err := s.SetEnvVar(ctx, resourceID, resourceType, key, value, false, orgID); err != nil {
			return fmt.Errorf("BulkSetEnvVars: key %s: %w", key, err)
		}
	}
	return nil
}

func (s *EnvService) ImportFromDotenv(content string) (map[string]string, error) {
	result := make(map[string]string)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		result[key] = value
	}

	return result, nil
}

func (s *EnvService) GetSecretValues(ctx context.Context, resourceID uuid.UUID, resourceType string, orgID uuid.UUID) ([]string, error) {
	vars, err := s.envRepo.ListByResource(ctx, resourceID, resourceType)
	if err != nil {
		return nil, err
	}

	orgKey, err := auth.DeriveOrgKey(s.encryptionKey, orgID)
	if err != nil {
		return nil, err
	}

	var secrets []string
	for _, v := range vars {
		if v.IsSecret {
			decrypted, err := auth.Decrypt(v.ValueEncrypted, orgKey)
			if err == nil {
				secrets = append(secrets, decrypted)
			}
		}
	}
	return secrets, nil
}
