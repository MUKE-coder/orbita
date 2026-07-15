package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/orbita-sh/orbita/cmd/grit/internal/ui"
)

// cloudGithubAuthCmd stores the user's GitHub token for repo creation/push.
// The full repo/push flow is Phase 4; the credential store lives here so init
// and deploy share it. The token is written to ~/.grit/github with 0600 perms.
func cloudGithubAuthCmd() *cobra.Command {
	var token string
	c := &cobra.Command{
		Use:   "github-auth",
		Short: "Store a GitHub token for repo creation/push (used by grit deploy)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				fmt.Print("GitHub token (repo + admin:repo_hook scope): ")
				r := bufio.NewReader(os.Stdin)
				line, _ := r.ReadString('\n')
				token = strings.TrimSpace(line)
			}
			if !strings.HasPrefix(token, "ghp_") && !strings.HasPrefix(token, "github_pat_") {
				ui.ErrorLine("that doesn't look like a GitHub token",
					"create one at https://github.com/settings/tokens with repo + admin:repo_hook scope")
				return fmt.Errorf("invalid token")
			}
			if err := storeGitHubToken(token); err != nil {
				return err
			}
			ui.Success("GitHub token stored")
			return nil
		},
	}
	c.Flags().StringVar(&token, "token", "", "GitHub token (prompted if omitted)")
	return c
}

// gitHubTokenPath returns ~/.grit/github (honors GRIT_HOME for tests).
func gitHubTokenPath() (string, error) {
	if p := os.Getenv("GRIT_GITHUB_TOKEN_FILE"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".grit", "github"), nil
}

func storeGitHubToken(token string) error {
	p, err := gitHubTokenPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(token+"\n"), 0o600)
}

// LoadGitHubToken reads the stored token, or "" if none.
func LoadGitHubToken() string {
	p, err := gitHubTokenPath()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
