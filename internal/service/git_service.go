package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

var (
	ErrGitConnectionNotFound = errors.New("git connection not found")
)

type GitService struct {
	gitRepo       *repository.GitRepository
	encryptionKey []byte
}

func NewGitService(gitRepo *repository.GitRepository, encryptionKey []byte) *GitService {
	return &GitService{
		gitRepo:       gitRepo,
		encryptionKey: encryptionKey,
	}
}

type CreateGitConnectionInput struct {
	Provider     string `json:"provider"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	BaseURL      string `json:"base_url,omitempty"` // for Gitea
}

func (s *GitService) CreateConnection(ctx context.Context, orgID uuid.UUID, input CreateGitConnectionInput) (*models.GitConnection, error) {
	orgKey, err := auth.DeriveOrgKey(s.encryptionKey, orgID)
	if err != nil {
		return nil, fmt.Errorf("CreateConnection: derive key: %w", err)
	}

	encToken, err := auth.Encrypt(input.AccessToken, orgKey)
	if err != nil {
		return nil, fmt.Errorf("CreateConnection: encrypt token: %w", err)
	}

	var encRefresh *string
	if input.RefreshToken != "" {
		r, err := auth.Encrypt(input.RefreshToken, orgKey)
		if err != nil {
			return nil, fmt.Errorf("CreateConnection: encrypt refresh: %w", err)
		}
		encRefresh = &r
	}

	metadata, _ := json.Marshal(map[string]string{
		"base_url": input.BaseURL,
	})

	conn := &models.GitConnection{
		ID:                    uuid.New(),
		OrganizationID:        orgID,
		Provider:              input.Provider,
		AccessTokenEncrypted:  encToken,
		RefreshTokenEncrypted: encRefresh,
		Metadata:              metadata,
	}

	if err := s.gitRepo.CreateConnection(ctx, conn); err != nil {
		return nil, fmt.Errorf("CreateConnection: %w", err)
	}

	return conn, nil
}

func (s *GitService) ListConnections(ctx context.Context, orgID uuid.UUID) ([]models.GitConnection, error) {
	return s.gitRepo.ListConnections(ctx, orgID)
}

func (s *GitService) DeleteConnection(ctx context.Context, id, orgID uuid.UUID) error {
	return s.gitRepo.DeleteConnection(ctx, id, orgID)
}

func (s *GitService) GetDecryptedToken(ctx context.Context, connID, orgID uuid.UUID) (string, error) {
	conn, err := s.gitRepo.FindConnectionByID(ctx, connID, orgID)
	if err != nil {
		return "", ErrGitConnectionNotFound
	}

	orgKey, err := auth.DeriveOrgKey(s.encryptionKey, orgID)
	if err != nil {
		return "", fmt.Errorf("GetDecryptedToken: derive key: %w", err)
	}

	token, err := auth.Decrypt(conn.AccessTokenEncrypted, orgKey)
	if err != nil {
		return "", fmt.Errorf("GetDecryptedToken: decrypt: %w", err)
	}

	return token, nil
}

// ListRepos returns a stub list of repos for the connected provider
func (s *GitService) ListRepos(ctx context.Context, connID, orgID uuid.UUID) ([]map[string]string, error) {
	// TODO: real impl — use provider API with decrypted token
	return []map[string]string{
		{"full_name": "user/my-app", "clone_url": "https://github.com/user/my-app.git", "default_branch": "main"},
		{"full_name": "user/api-service", "clone_url": "https://github.com/user/api-service.git", "default_branch": "main"},
	}, nil
}

// ListBranches returns a stub list of branches
func (s *GitService) ListBranches(ctx context.Context, connID, orgID uuid.UUID, owner, repo string) ([]string, error) {
	// TODO: real impl — use provider API
	return []string{"main", "develop", "staging"}, nil
}

// FindAppByRepoAndBranch finds an app configured for auto-deploy on a given repo+branch
func (s *GitService) FindAppByRepoAndBranch(ctx context.Context, repoURL, branch string) (*models.Application, error) {
	return s.gitRepo.FindAppByRepoAndBranch(ctx, repoURL, branch)
}
