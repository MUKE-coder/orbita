// Package sshx runs commands and uploads files on a remote server over SSH.
// orbita init uses it to harden and install Orbita on a fresh VPS.
package sshx

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Target identifies a remote server: user@host[:port].
type Target struct {
	User string
	Host string
	Port string
}

// ParseTarget parses "user@host", "user@host:port", or "host" (defaults user
// root, port 22).
func ParseTarget(s string) (Target, error) {
	t := Target{User: "root", Port: "22"}
	if s == "" {
		return t, fmt.Errorf("empty target")
	}
	if at := strings.LastIndex(s, "@"); at >= 0 {
		t.User = s[:at]
		s = s[at+1:]
	}
	if colon := strings.LastIndex(s, ":"); colon >= 0 {
		t.Host = s[:colon]
		t.Port = s[colon+1:]
	} else {
		t.Host = s
	}
	if t.Host == "" {
		return t, fmt.Errorf("no host in target")
	}
	return t, nil
}

// Addr returns host:port.
func (t Target) Addr() string { return net.JoinHostPort(t.Host, t.Port) }

// Client is a connected SSH session factory.
type Client struct {
	client *ssh.Client
	target Target

	// Host key as presented by the server during the handshake, plus the exact
	// hostname/remote pair OpenSSH's own callback would have seen. Captured so
	// we can tell the user whether their ~/.ssh/known_hosts entry is stale.
	hostKey      ssh.PublicKey
	hostKeyName  string
	hostKeyRemot net.Addr
}

// Auth describes how to authenticate the initial SSH connection to a server.
// For a fresh VPS this is either a key (added at creation) or the provider's
// root password.
type Auth struct {
	KeyFile  string // explicit private key path (optional)
	Password string // root/user password (optional)
}

// Connect dials the target using, in order: the explicit password (if given),
// an explicit key file, the SSH agent, then ~/.ssh/id_ed25519 / id_rsa. Host
// keys are accepted on first use (the operator is provisioning their own box).
func Connect(t Target, auth Auth) (*Client, error) {
	auths, err := authMethods(auth)
	if err != nil {
		return nil, err
	}
	// Trust on first use — the operator is provisioning their own box — but
	// record what the server presented so KnownHostState can compare it against
	// ~/.ssh/known_hosts afterwards.
	var (
		hostKey  ssh.PublicKey
		hostName string
		remote   net.Addr
	)
	cfg := &ssh.ClientConfig{
		User: t.User,
		Auth: auths,
		HostKeyCallback: func(hostname string, rem net.Addr, key ssh.PublicKey) error {
			hostKey, hostName, remote = key, hostname, rem
			return nil
		},
		Timeout: 15 * time.Second,
	}
	client, err := ssh.Dial("tcp", t.Addr(), cfg)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s: %w", t.Addr(), err)
	}
	return &Client{
		client:       client,
		target:       t,
		hostKey:      hostKey,
		hostKeyName:  hostName,
		hostKeyRemot: remote,
	}, nil
}

func (c *Client) Close() error { return c.client.Close() }

// Run executes a command, streaming stdout/stderr to the given writers, and
// returns the exit status.
func (c *Client) Run(cmd string, stdout, stderr io.Writer) error {
	sess, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()
	sess.Stdout = stdout
	sess.Stderr = stderr
	return sess.Run(cmd)
}

// Output runs a command and returns its combined output.
func (c *Client) Output(cmd string) (string, error) {
	sess, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	var buf bytes.Buffer
	sess.Stdout = &buf
	sess.Stderr = &buf
	err = sess.Run(cmd)
	return buf.String(), err
}

// Upload writes content to a remote path via `cat > path` (no scp dependency).
func (c *Client) Upload(content []byte, remotePath string, mode string) error {
	sess, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	stdin, err := sess.StdinPipe()
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("cat > %q && chmod %s %q", remotePath, mode, remotePath)
	if err := sess.Start(cmd); err != nil {
		return err
	}
	if _, err := stdin.Write(content); err != nil {
		return err
	}
	_ = stdin.Close()
	return sess.Wait()
}

func authMethods(auth Auth) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	// Root/user password (fresh VPS with a provider-given password)
	if auth.Password != "" {
		methods = append(methods, ssh.Password(auth.Password))
	}

	// Explicit key file
	if auth.KeyFile != "" {
		signer, err := loadKey(auth.KeyFile)
		if err != nil {
			return nil, err
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	// SSH agent
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		if conn, err := net.Dial("unix", sock); err == nil {
			ag := agent.NewClient(conn)
			methods = append(methods, ssh.PublicKeysCallback(ag.Signers))
		}
	}

	// Default keys
	if home, err := os.UserHomeDir(); err == nil {
		for _, name := range []string{"id_ed25519", "id_rsa"} {
			if signer, err := loadKey(filepath.Join(home, ".ssh", name)); err == nil {
				methods = append(methods, ssh.PublicKeys(signer))
			}
		}
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no SSH auth available — provide a key or password, or start an ssh-agent")
	}
	return methods, nil
}

// TestReachable dials the target's SSH port to check the server is up and
// reachable before we try to authenticate. Returns nil if a TCP connection
// opens within the timeout.
func TestReachable(t Target, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", t.Addr(), timeout)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

func loadKey(path string) (ssh.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(data)
}
