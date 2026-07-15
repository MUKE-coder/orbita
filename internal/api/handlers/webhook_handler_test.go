package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func signBody(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyGitHubSignature(t *testing.T) {
	body := []byte(`{"ref":"refs/heads/main"}`)
	secret := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	tests := []struct {
		name      string
		body      []byte
		signature string
		secret    string
		want      bool
	}{
		{"valid signature", body, signBody(body, secret), secret, true},
		{"wrong secret", body, signBody(body, "other-secret"), secret, false},
		{"tampered body", []byte(`{"ref":"refs/heads/evil"}`), signBody(body, secret), secret, false},
		{"missing signature", body, "", secret, false},
		{"malformed signature (no prefix)", body, hex.EncodeToString([]byte("junk")), secret, false},
		{"prefix only", body, "sha256=", secret, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := verifyGitHubSignature(tt.body, tt.signature, tt.secret); got != tt.want {
				t.Errorf("verifyGitHubSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}
