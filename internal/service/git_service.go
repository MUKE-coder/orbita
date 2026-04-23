package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/models"
	"github.com/orbita-sh/orbita/internal/repository"
)

var (
	ErrGitConnectionNotFound = errors.New("git connection not found")
	ErrUnsupportedProvider   = errors.New("unsupported git provider")
)

type GitService struct {
	gitRepo       *repository.GitRepository
	encryptionKey []byte
	http          *http.Client
}

func NewGitService(gitRepo *repository.GitRepository, encryptionKey []byte) *GitService {
	return &GitService{
		gitRepo:       gitRepo,
		encryptionKey: encryptionKey,
		http:          &http.Client{Timeout: 15 * time.Second},
	}
}

type CreateGitConnectionInput struct {
	Provider     string `json:"provider"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	BaseURL      string `json:"base_url,omitempty"` // for GitLab self-hosted / Gitea
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

// GetDecryptedToken returns the plaintext access token for a connection.
func (s *GitService) GetDecryptedToken(ctx context.Context, connID, orgID uuid.UUID) (string, error) {
	_, token, err := s.GetConnectionAndToken(ctx, connID, orgID)
	return token, err
}

// GetConnectionAndToken returns the connection record plus its decrypted access
// token. Used by the orchestrator to build a git URL with an embedded PAT.
func (s *GitService) GetConnectionAndToken(ctx context.Context, connID, orgID uuid.UUID) (*models.GitConnection, string, error) {
	conn, err := s.gitRepo.FindConnectionByID(ctx, connID, orgID)
	if err != nil {
		return nil, "", ErrGitConnectionNotFound
	}

	orgKey, err := auth.DeriveOrgKey(s.encryptionKey, orgID)
	if err != nil {
		return nil, "", fmt.Errorf("GetConnectionAndToken: derive key: %w", err)
	}

	token, err := auth.Decrypt(conn.AccessTokenEncrypted, orgKey)
	if err != nil {
		return nil, "", fmt.Errorf("GetConnectionAndToken: decrypt: %w", err)
	}

	return conn, token, nil
}

// ---------- provider API calls ----------

type repoMeta struct {
	FullName      string `json:"full_name"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

// ListRepos returns the repositories the connection's token can see.
func (s *GitService) ListRepos(ctx context.Context, connID, orgID uuid.UUID) ([]map[string]string, error) {
	conn, token, err := s.GetConnectionAndToken(ctx, connID, orgID)
	if err != nil {
		return nil, err
	}

	baseURL := extractBaseURL(conn)

	switch conn.Provider {
	case "github":
		return s.listGitHubRepos(ctx, token)
	case "gitlab":
		return s.listGitLabRepos(ctx, token, baseURL)
	case "gitea":
		return s.listGiteaRepos(ctx, token, baseURL)
	}
	return nil, ErrUnsupportedProvider
}

// ListBranches returns branch names for a specific repository.
func (s *GitService) ListBranches(ctx context.Context, connID, orgID uuid.UUID, owner, repo string) ([]string, error) {
	conn, token, err := s.GetConnectionAndToken(ctx, connID, orgID)
	if err != nil {
		return nil, err
	}

	baseURL := extractBaseURL(conn)

	switch conn.Provider {
	case "github":
		return s.listGitHubBranches(ctx, token, owner, repo)
	case "gitlab":
		return s.listGitLabBranches(ctx, token, baseURL, owner, repo)
	case "gitea":
		return s.listGiteaBranches(ctx, token, baseURL, owner, repo)
	}
	return nil, ErrUnsupportedProvider
}

func extractBaseURL(conn *models.GitConnection) string {
	var m map[string]string
	_ = json.Unmarshal(conn.Metadata, &m)
	b := m["base_url"]
	return strings.TrimRight(b, "/")
}

// ---- GitHub ----

func (s *GitService) listGitHubRepos(ctx context.Context, token string) ([]map[string]string, error) {
	// Paginate up to 3 pages (300 repos) — plenty for most agencies
	out := []map[string]string{}
	for page := 1; page <= 3; page++ {
		url := fmt.Sprintf("https://api.github.com/user/repos?per_page=100&sort=updated&page=%d", page)
		var repos []repoMeta
		if err := s.githubAPI(ctx, "GET", url, token, &repos); err != nil {
			return nil, err
		}
		for _, r := range repos {
			out = append(out, map[string]string{
				"full_name":      r.FullName,
				"clone_url":      r.CloneURL,
				"default_branch": r.DefaultBranch,
			})
		}
		if len(repos) < 100 {
			break
		}
	}
	return out, nil
}

func (s *GitService) listGitHubBranches(ctx context.Context, token, owner, repo string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches?per_page=100", owner, repo)
	var branches []struct {
		Name string `json:"name"`
	}
	if err := s.githubAPI(ctx, "GET", url, token, &branches); err != nil {
		return nil, err
	}
	names := make([]string, len(branches))
	for i, b := range branches {
		names[i] = b.Name
	}
	return names, nil
}

func (s *GitService) githubAPI(ctx context.Context, method, url, token string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := s.http.Do(req)
	if err != nil {
		return fmt.Errorf("github: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github: HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// ---- GitLab ----

func (s *GitService) listGitLabRepos(ctx context.Context, token, baseURL string) ([]map[string]string, error) {
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	url := fmt.Sprintf("%s/api/v4/projects?membership=true&per_page=100&order_by=updated_at", baseURL)
	var projects []struct {
		PathWithNamespace string `json:"path_with_namespace"`
		HTTPURLToRepo     string `json:"http_url_to_repo"`
		DefaultBranch     string `json:"default_branch"`
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("PRIVATE-TOKEN", token)

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gitlab: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitlab: HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	out := make([]map[string]string, len(projects))
	for i, p := range projects {
		out[i] = map[string]string{
			"full_name":      p.PathWithNamespace,
			"clone_url":      p.HTTPURLToRepo,
			"default_branch": p.DefaultBranch,
		}
	}
	return out, nil
}

func (s *GitService) listGitLabBranches(ctx context.Context, token, baseURL, owner, repo string) ([]string, error) {
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	project := fmt.Sprintf("%s/%s", owner, repo)
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/branches?per_page=100", baseURL, urlPathEscape(project))

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("PRIVATE-TOKEN", token)
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gitlab: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitlab: HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	var branches []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}
	names := make([]string, len(branches))
	for i, b := range branches {
		names[i] = b.Name
	}
	return names, nil
}

// ---- Gitea (self-hosted) ----

func (s *GitService) listGiteaRepos(ctx context.Context, token, baseURL string) ([]map[string]string, error) {
	if baseURL == "" {
		return nil, errors.New("gitea: base_url required")
	}
	url := fmt.Sprintf("%s/api/v1/repos/search?limit=50&token=%s", baseURL, token)
	var body struct {
		Data []repoMeta `json:"data"`
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "token "+token)
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gitea: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitea: HTTP %d: %s", resp.StatusCode, truncate(string(b), 200))
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	out := make([]map[string]string, len(body.Data))
	for i, r := range body.Data {
		out[i] = map[string]string{
			"full_name":      r.FullName,
			"clone_url":      r.CloneURL,
			"default_branch": r.DefaultBranch,
		}
	}
	return out, nil
}

func (s *GitService) listGiteaBranches(ctx context.Context, token, baseURL, owner, repo string) ([]string, error) {
	if baseURL == "" {
		return nil, errors.New("gitea: base_url required")
	}
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/branches", baseURL, owner, repo)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "token "+token)
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gitea: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitea: HTTP %d: %s", resp.StatusCode, truncate(string(b), 200))
	}
	var branches []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}
	names := make([]string, len(branches))
	for i, b := range branches {
		names[i] = b.Name
	}
	return names, nil
}

// ---- misc ----

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

// urlPathEscape encodes a path segment ("group/subgroup/repo" → "group%2Fsubgroup%2Frepo").
func urlPathEscape(s string) string {
	return strings.ReplaceAll(s, "/", "%2F")
}

// FindAppByRepoAndBranch finds an app configured for auto-deploy on a given repo+branch.
func (s *GitService) FindAppByRepoAndBranch(ctx context.Context, repoURL, branch string) (*models.Application, error) {
	return s.gitRepo.FindAppByRepoAndBranch(ctx, repoURL, branch)
}
