package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/orbita-sh/orbita/internal/models"
)

type GitRepository struct {
	db *gorm.DB
}

func NewGitRepository(db *gorm.DB) *GitRepository {
	return &GitRepository{db: db}
}

func (r *GitRepository) CreateConnection(ctx context.Context, conn *models.GitConnection) error {
	if err := r.db.WithContext(ctx).Create(conn).Error; err != nil {
		return fmt.Errorf("GitRepo.CreateConnection: %w", err)
	}
	return nil
}

func (r *GitRepository) FindConnectionByID(ctx context.Context, id, orgID uuid.UUID) (*models.GitConnection, error) {
	var conn models.GitConnection
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("id = ?", id).First(&conn).Error; err != nil {
		return nil, fmt.Errorf("GitRepo.FindConnectionByID: %w", err)
	}
	return &conn, nil
}

func (r *GitRepository) ListConnections(ctx context.Context, orgID uuid.UUID) ([]models.GitConnection, error) {
	var conns []models.GitConnection
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Order("created_at DESC").Find(&conns).Error; err != nil {
		return nil, fmt.Errorf("GitRepo.ListConnections: %w", err)
	}
	return conns, nil
}

func (r *GitRepository) DeleteConnection(ctx context.Context, id, orgID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("id = ?", id).Delete(&models.GitConnection{}).Error; err != nil {
		return fmt.Errorf("GitRepo.DeleteConnection: %w", err)
	}
	return nil
}

func (r *GitRepository) FindAppByRepoAndBranch(ctx context.Context, repoURL, branch string) (*models.Application, error) {
	var app models.Application
	if err := r.db.WithContext(ctx).
		Where("source_config->>'repo_url' = ? AND source_config->>'branch' = ? AND auto_deploy = true",
			repoURL, branch).First(&app).Error; err != nil {
		return nil, fmt.Errorf("GitRepo.FindAppByRepoAndBranch: %w", err)
	}
	return &app, nil
}

// Registry Credentials

func (r *GitRepository) CreateRegistryCredential(ctx context.Context, cred *models.RegistryCredential) error {
	if err := r.db.WithContext(ctx).Create(cred).Error; err != nil {
		return fmt.Errorf("GitRepo.CreateRegistryCredential: %w", err)
	}
	return nil
}

func (r *GitRepository) ListRegistryCredentials(ctx context.Context, orgID uuid.UUID) ([]models.RegistryCredential, error) {
	var creds []models.RegistryCredential
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).Find(&creds).Error; err != nil {
		return nil, fmt.Errorf("GitRepo.ListRegistryCredentials: %w", err)
	}
	return creds, nil
}

func (r *GitRepository) DeleteRegistryCredential(ctx context.Context, id, orgID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Scopes(OrgScope(orgID)).
		Where("id = ?", id).Delete(&models.RegistryCredential{}).Error; err != nil {
		return fmt.Errorf("GitRepo.DeleteRegistryCredential: %w", err)
	}
	return nil
}
