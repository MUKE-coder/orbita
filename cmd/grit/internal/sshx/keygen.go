package sshx

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// KeyPair is a generated SSH keypair. PrivatePEM is the OpenSSH private key;
// PublicAuthorized is the single-line authorized_keys entry to install on the
// server.
type KeyPair struct {
	PrivatePEM       []byte
	PublicAuthorized string
	PrivateKeyPath   string // set once SaveTo writes it
}

// GenerateEd25519 creates a fresh ed25519 SSH keypair in memory. The private
// key never leaves the operator's machine — only PublicAuthorized is sent to
// the server (installed for the deploy user by the hardening step), so there's
// no scp-the-key-back dance.
func GenerateEd25519(comment string) (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("GenerateEd25519: %w", err)
	}

	pemBlock, err := ssh.MarshalPrivateKey(priv, comment)
	if err != nil {
		return nil, fmt.Errorf("GenerateEd25519: marshal private: %w", err)
	}

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("GenerateEd25519: public: %w", err)
	}
	authorized := string(ssh.MarshalAuthorizedKey(sshPub))
	if comment != "" {
		// MarshalAuthorizedKey ends with a newline; append the comment before it.
		authorized = authorized[:len(authorized)-1] + " " + comment + "\n"
	}

	return &KeyPair{
		PrivatePEM:       pem.EncodeToMemory(pemBlock),
		PublicAuthorized: authorized,
	}, nil
}

// SaveTo writes the private key to ~/.ssh/<name> (0600) and the public key to
// ~/.ssh/<name>.pub (0644), returning the private key path. Won't overwrite an
// existing key of the same name.
func (k *KeyPair) SaveTo(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return "", err
	}
	priv := filepath.Join(sshDir, name)
	if _, err := os.Stat(priv); err == nil {
		return "", fmt.Errorf("key %s already exists — choose another name or remove it", priv)
	}
	if err := os.WriteFile(priv, k.PrivatePEM, 0o600); err != nil {
		return "", err
	}
	if err := os.WriteFile(priv+".pub", []byte(k.PublicAuthorized), 0o644); err != nil {
		return "", err
	}
	k.PrivateKeyPath = priv
	return priv, nil
}
