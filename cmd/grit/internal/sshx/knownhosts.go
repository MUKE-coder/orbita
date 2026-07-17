package sshx

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh/knownhosts"
)

// HostKeyState says what the user's ~/.ssh/known_hosts has to say about a
// server we just connected to.
type HostKeyState int

const (
	// HostKeyAbsent means known_hosts has no entry for this host. Nothing to
	// clean up — the user's next `ssh` will simply prompt to accept the key.
	HostKeyAbsent HostKeyState = iota
	// HostKeyMatch means the saved key matches the server. All good.
	HostKeyMatch
	// HostKeyMismatch means known_hosts has a *different* key for this host —
	// almost always a rebuilt/reimaged VPS reusing the IP. The user's own
	// `ssh` will refuse to connect until the stale entry is removed.
	HostKeyMismatch
)

// KnownHostsPath returns ~/.ssh/known_hosts, or "" if the home dir is unknown.
func KnownHostsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "known_hosts")
}

// KnownHostEntry returns the key this host is stored under in known_hosts:
// "host" on port 22, "[host]:port" otherwise — OpenSSH's own convention.
func KnownHostEntry(t Target) string {
	return knownhosts.Normalize(t.Addr())
}

// KnownHostState compares the host key the server presented during the
// handshake against ~/.ssh/known_hosts.
//
// This matters even though we connect with a trust-on-first-use callback: the
// operator will `ssh` into this box themselves afterwards, and OpenSSH *will*
// refuse a mismatched key ("REMOTE HOST IDENTIFICATION HAS CHANGED").
func (c *Client) KnownHostState() (HostKeyState, error) {
	if c.hostKey == nil {
		return HostKeyAbsent, errors.New("no host key captured for this connection")
	}
	path := KnownHostsPath()
	if path == "" {
		return HostKeyAbsent, nil
	}
	if _, err := os.Stat(path); err != nil {
		// No known_hosts file at all — nothing is stale.
		return HostKeyAbsent, nil //nolint:nilerr // absence is not an error here
	}

	cb, err := knownhosts.New(path)
	if err != nil {
		return HostKeyAbsent, fmt.Errorf("read %s: %w", path, err)
	}

	// Replay the exact hostname/remote pair from the handshake, so hashed
	// entries and host aliases resolve the way OpenSSH would resolve them.
	err = cb(c.hostKeyName, c.hostKeyRemot, c.hostKey)
	if err == nil {
		return HostKeyMatch, nil
	}
	var ke *knownhosts.KeyError
	if errors.As(err, &ke) {
		// Want non-empty => we know this host, under a different key.
		if len(ke.Want) > 0 {
			return HostKeyMismatch, nil
		}
		return HostKeyAbsent, nil
	}
	var re *knownhosts.RevokedError
	if errors.As(err, &re) {
		return HostKeyMismatch, nil
	}
	return HostKeyAbsent, err
}

// ForgetHostCommand is the command a user would run by hand to drop the stale
// entry. Printed so they can do it themselves if we can't.
func ForgetHostCommand(t Target) string {
	return "ssh-keygen -R " + KnownHostEntry(t)
}

// ForgetHost removes the saved host key for t from ~/.ssh/known_hosts by
// running `ssh-keygen -R <host>`.
//
// We shell out instead of rewriting known_hosts ourselves because entries may
// be hashed (HashKnownHosts yes, the default on many distros) — ssh-keygen
// matches those correctly and we would not.
func ForgetHost(t Target) (string, error) {
	entry := KnownHostEntry(t)
	bin, err := exec.LookPath("ssh-keygen")
	if err != nil {
		return "", fmt.Errorf("ssh-keygen not found in PATH — run %q by hand: %w",
			ForgetHostCommand(t), err)
	}
	out, err := exec.Command(bin, "-R", entry).CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s: %w", ForgetHostCommand(t), err)
	}
	return strings.TrimSpace(string(out)), nil
}
