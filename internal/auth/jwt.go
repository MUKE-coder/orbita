package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 30 * 24 * time.Hour
)

type Claims struct {
	UserID       uuid.UUID  `json:"user_id"`
	Email        string     `json:"email"`
	IsSuperAdmin bool       `json:"is_super_admin"`
	OrgID        *uuid.UUID `json:"org_id,omitempty"`
	// MustChangePassword rides in the token so the gate costs no database read
	// per request. Tokens are short-lived (15m) and reissued the moment the
	// password is changed, so the flag can't go stale for long.
	MustChangePassword bool `json:"must_change_password,omitempty"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID uuid.UUID, email string, isSuperAdmin, mustChangePassword bool, secret string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:             userID,
		Email:              email,
		IsSuperAdmin:       isSuperAdmin,
		MustChangePassword: mustChangePassword,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "orbita",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("GenerateAccessToken: %w", err)
	}
	return signed, nil
}

func GenerateRefreshToken(userID uuid.UUID, sessionID uuid.UUID, secret string) (string, error) {
	now := time.Now()
	claims := RefreshClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "orbita",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("GenerateRefreshToken: %w", err)
	}
	return signed, nil
}

func ValidateAccessToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("ValidateAccessToken: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("ValidateAccessToken: invalid token claims")
	}

	return claims, nil
}

func ValidateRefreshToken(tokenString, secret string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("ValidateRefreshToken: %w", err)
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("ValidateRefreshToken: invalid token claims")
	}

	return claims, nil
}
