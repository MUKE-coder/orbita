// Package github ensures the target repo exists and pushes the current project,
// using the token stored by `orbita github-auth`. It shells out to `git`
// for push (respecting the local repo) and uses the GitHub REST API to create
// the repo when missing.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// Client wraps the GitHub REST API with a personal access token.
type Client struct {
	Token string
	HTTP  *http.Client
}

func New(token string) *Client {
	return &Client{Token: token, HTTP: &http.Client{Timeout: 30 * time.Second}}
}

// RepoExists reports whether owner/name exists and is accessible with the token.
func (c *Client) RepoExists(ctx context.Context, owner, name string) (bool, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, name), nil)
	c.auth(req)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		return true, nil
	case 404:
		return false, nil
	default:
		return false, fmt.Errorf("github repo check: HTTP %d", resp.StatusCode)
	}
}

// CreateRepo creates a repo. If owner is the authenticated user, uses /user/repos;
// otherwise creates under the org (/orgs/{owner}/repos). Private by default.
func (c *Client) CreateRepo(ctx context.Context, owner, name string, private bool) error {
	login, err := c.authUser(ctx)
	if err != nil {
		return err
	}
	url := "https://api.github.com/user/repos"
	if !strings.EqualFold(login, owner) {
		url = fmt.Sprintf("https://api.github.com/orgs/%s/repos", owner)
	}
	body, _ := json.Marshal(map[string]interface{}{"name": name, "private": private})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	c.auth(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("github create repo: HTTP %d", resp.StatusCode)
	}
	return nil
}

// EnsureAndPush makes sure owner/name exists (creating it private if missing) and
// pushes the given branch from dir over HTTPS with the token embedded.
func (c *Client) EnsureAndPush(ctx context.Context, dir, owner, name, branch string) error {
	exists, err := c.RepoExists(ctx, owner, name)
	if err != nil {
		return err
	}
	if !exists {
		if err := c.CreateRepo(ctx, owner, name, true); err != nil {
			return err
		}
	}

	// Ensure a git repo exists with a commit.
	if err := ensureGitRepo(dir, branch); err != nil {
		return err
	}

	remote := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", c.Token, owner, name)
	// Push the branch, setting the authenticated remote inline (no persisted creds).
	if out, err := runGit(dir, "push", remote, "HEAD:refs/heads/"+branch); err != nil {
		return fmt.Errorf("git push: %w: %s", err, out)
	}
	return nil
}

func (c *Client) auth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
}

func (c *Client) authUser(ctx context.Context) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	c.auth(req)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("github auth: HTTP %d (check your token)", resp.StatusCode)
	}
	var u struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return "", err
	}
	return u.Login, nil
}

// ensureGitRepo inits a repo + commit if needed so there's something to push.
func ensureGitRepo(dir, branch string) error {
	if out, err := runGit(dir, "rev-parse", "--is-inside-work-tree"); err != nil || !strings.Contains(out, "true") {
		if _, err := runGit(dir, "init"); err != nil {
			return err
		}
	}
	_, _ = runGit(dir, "checkout", "-B", branch)
	_, _ = runGit(dir, "add", "-A")
	// Commit only if there's something staged (ignore "nothing to commit").
	_, _ = runGit(dir, "-c", "user.email=grit@deploy", "-c", "user.name=grit", "commit", "-m", "orbita deploy")
	return nil
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}
