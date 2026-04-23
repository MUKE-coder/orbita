package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// GitTokenAdapter bridges GitService to the orchestrator.TokenFetcher interface,
// which uses string IDs to avoid a circular dependency between the two packages.
type GitTokenAdapter struct {
	Git *GitService
}

func NewGitTokenAdapter(g *GitService) *GitTokenAdapter {
	return &GitTokenAdapter{Git: g}
}

// ResolveGitToken implements orchestrator.TokenFetcher.
func (a *GitTokenAdapter) ResolveGitToken(ctx context.Context, connIDStr, orgIDStr string) (provider, token, baseURL string, err error) {
	connID, err := uuid.Parse(connIDStr)
	if err != nil {
		return "", "", "", err
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return "", "", "", err
	}

	conn, t, err := a.Git.GetConnectionAndToken(ctx, connID, orgID)
	if err != nil {
		return "", "", "", err
	}

	var meta map[string]string
	_ = json.Unmarshal(conn.Metadata, &meta)

	return conn.Provider, t, meta["base_url"], nil
}
