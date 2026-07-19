package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/google/uuid"
	"golang.org/x/crypto/hkdf"
)

func Encrypt(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("Encrypt: new cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("Encrypt: new GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("Encrypt: generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("Decrypt: decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("Decrypt: new cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("Decrypt: new GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("Decrypt: ciphertext too short")
	}

	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", fmt.Errorf("Decrypt: open: %w", err)
	}

	return string(plaintext), nil
}

func DeriveOrgKey(masterKey []byte, orgID uuid.UUID) ([]byte, error) {
	hkdfReader := hkdf.New(sha256.New, masterKey, orgID[:], []byte("orbita-org-key"))
	derivedKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, derivedKey); err != nil {
		return nil, fmt.Errorf("DeriveOrgKey: %w", err)
	}
	return derivedKey, nil
}

// DerivePlatformKey derives the key for instance-wide secrets that belong to no
// organisation — the Resend API key, for example.
//
// Separate info string from DeriveOrgKey so platform and tenant ciphertexts can
// never be decrypted with each other's keys, and so the master key is still
// never used to encrypt anything directly.
func DerivePlatformKey(masterKey []byte) ([]byte, error) {
	hkdfReader := hkdf.New(sha256.New, masterKey, []byte("orbita-platform"), []byte("orbita-platform-key"))
	derivedKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, derivedKey); err != nil {
		return nil, fmt.Errorf("DerivePlatformKey: %w", err)
	}
	return derivedKey, nil
}

// passwordAlphabet omits characters that are easy to confuse when a password is
// read off a screen and typed by hand — 0/O, 1/l/I — because these are handed
// over person to person rather than pasted from a manager.
const passwordAlphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789@#%+=?"

// GeneratePassword returns a cryptographically random password of n characters.
// Used for admin-provisioned accounts, which the user must change at first
// sign-in anyway.
func GeneratePassword(n int) (string, error) {
	if n < 12 {
		n = 12
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("GeneratePassword: %w", err)
	}
	// Modulo bias is negligible here: 62-char alphabet over a 256-value byte,
	// and the output is a throwaway credential, not a long-lived key.
	out := make([]byte, n)
	for i, v := range b {
		out[i] = passwordAlphabet[int(v)%len(passwordAlphabet)]
	}
	return string(out), nil
}

func GenerateRandomToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("GenerateRandomToken: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func GenerateOTP() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("GenerateOTP: %w", err)
	}
	otp := (int(b[0])<<16 | int(b[1])<<8 | int(b[2])) % 1000000
	return fmt.Sprintf("%06d", otp), nil
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(h[:])
}
