package auth

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	master := []byte("0123456789abcdef0123456789abcdef") // 32 bytes
	orgID := uuid.New()

	key, err := DeriveOrgKey(master, orgID)
	if err != nil {
		t.Fatalf("DeriveOrgKey: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("derived key length = %d, want 32 (AES-256)", len(key))
	}

	plaintexts := []string{"", "secret-value", "postgres://u:p@h:5432/db?sslmode=disable", strings.Repeat("x", 10_000)}
	for _, pt := range plaintexts {
		ct, err := Encrypt(pt, key)
		if err != nil {
			t.Fatalf("Encrypt(%q): %v", pt[:min(20, len(pt))], err)
		}
		if ct == pt && pt != "" {
			t.Errorf("ciphertext equals plaintext")
		}
		got, err := Decrypt(ct, key)
		if err != nil {
			t.Fatalf("Decrypt: %v", err)
		}
		if got != pt {
			t.Errorf("round trip mismatch: got %q want %q", got[:min(20, len(got))], pt[:min(20, len(pt))])
		}
	}
}

func TestEncryptNonceUniqueness(t *testing.T) {
	master := []byte("0123456789abcdef0123456789abcdef")
	key, _ := DeriveOrgKey(master, uuid.New())

	a, _ := Encrypt("same plaintext", key)
	b, _ := Encrypt("same plaintext", key)
	if a == b {
		t.Error("two encryptions of the same plaintext produced identical ciphertexts (nonce reuse?)")
	}
}

func TestDeriveOrgKeyIsolation(t *testing.T) {
	master := []byte("0123456789abcdef0123456789abcdef")
	org1, org2 := uuid.New(), uuid.New()

	k1, _ := DeriveOrgKey(master, org1)
	k2, _ := DeriveOrgKey(master, org2)
	if string(k1) == string(k2) {
		t.Fatal("different orgs derived the same key")
	}

	// A value encrypted for org1 must not decrypt with org2's key
	ct, _ := Encrypt("org1-secret", k1)
	if _, err := Decrypt(ct, k2); err == nil {
		t.Error("cross-org decrypt succeeded — tenant isolation broken")
	}

	// Same org derives the same key deterministically
	k1again, _ := DeriveOrgKey(master, org1)
	if string(k1) != string(k1again) {
		t.Error("key derivation is not deterministic for the same org")
	}
}

func TestDecryptRejectsTampering(t *testing.T) {
	master := []byte("0123456789abcdef0123456789abcdef")
	key, _ := DeriveOrgKey(master, uuid.New())

	ct, _ := Encrypt("integrity matters", key)
	raw, _ := base64.StdEncoding.DecodeString(ct)
	raw[len(raw)-1] ^= 0xFF // flip a bit in the auth tag / ciphertext
	tampered := base64.StdEncoding.EncodeToString(raw)

	if _, err := Decrypt(tampered, key); err == nil {
		t.Error("tampered ciphertext decrypted without error (GCM auth not enforced?)")
	}

	if _, err := Decrypt("bm90LWEtY2lwaGVydGV4dA==", key); err == nil {
		t.Error("garbage ciphertext decrypted without error")
	}

	if _, err := Decrypt("AAAA", key); err == nil {
		t.Error("too-short ciphertext accepted")
	}
}
