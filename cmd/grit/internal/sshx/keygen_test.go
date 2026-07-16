package sshx

import (
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestGenerateEd25519(t *testing.T) {
	kp, err := GenerateEd25519("deploy@orbita")
	if err != nil {
		t.Fatal(err)
	}

	// Public key parses as a valid authorized_keys line with the comment.
	pub, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(kp.PublicAuthorized))
	if err != nil {
		t.Fatalf("public key not parseable: %v", err)
	}
	if pub.Type() != "ssh-ed25519" {
		t.Errorf("key type = %q, want ssh-ed25519", pub.Type())
	}
	if comment != "deploy@orbita" {
		t.Errorf("comment = %q", comment)
	}

	// Private key parses and its signer's public key matches.
	signer, err := ssh.ParsePrivateKey(kp.PrivatePEM)
	if err != nil {
		t.Fatalf("private key not parseable: %v", err)
	}
	if string(signer.PublicKey().Marshal()) != string(pub.Marshal()) {
		t.Error("private key's public half doesn't match the generated public key")
	}
}

func TestGenerateEd25519UniquePerCall(t *testing.T) {
	a, _ := GenerateEd25519("x")
	b, _ := GenerateEd25519("x")
	if strings.TrimSpace(a.PublicAuthorized) == strings.TrimSpace(b.PublicAuthorized) {
		t.Error("two generated keys are identical")
	}
}
