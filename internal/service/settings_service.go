package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/orbita-sh/orbita/internal/auth"
	"github.com/orbita-sh/orbita/internal/config"
	"github.com/orbita-sh/orbita/internal/mailer"
	"github.com/orbita-sh/orbita/internal/repository"
)

// SettingsService owns instance-wide configuration.
//
// Values set in the dashboard take precedence over environment variables, so an
// operator can turn on email after install without touching compose. Reads are
// cached briefly because the mailer resolves settings on every send.
type SettingsService struct {
	repo *repository.SettingsRepository
	cfg  *config.Config

	mu       sync.RWMutex
	cached   *mailer.Settings
	cachedAt time.Time
}

// settingsCacheTTL keeps a hot path (every outbound email) off the database
// without making a settings change feel unresponsive.
const settingsCacheTTL = 30 * time.Second

func NewSettingsService(repo *repository.SettingsRepository, cfg *config.Config) *SettingsService {
	return &SettingsService{repo: repo, cfg: cfg}
}

// EmailSettings is the API view: whether a key is configured, never the key.
type EmailSettings struct {
	Configured    bool   `json:"configured"`
	HasAPIKey     bool   `json:"has_api_key"`
	EmailFrom     string `json:"email_from"`
	EmailFromName string `json:"email_from_name"`
	// Source tells the operator where the live values come from, so a stale env
	// var doesn't look like a dashboard bug.
	Source string `json:"source"` // "dashboard" | "environment" | "unset"
}

// MailerSettings resolves the credentials to send with: dashboard first, then
// environment. Safe for concurrent use; never returns an error, since a send
// failure is reported by the mailer itself.
func (s *SettingsService) MailerSettings(ctx context.Context) mailer.Settings {
	s.mu.RLock()
	if s.cached != nil && time.Since(s.cachedAt) < settingsCacheTTL {
		cached := *s.cached
		s.mu.RUnlock()
		return cached
	}
	s.mu.RUnlock()

	resolved := s.resolve(ctx)

	s.mu.Lock()
	s.cached = &resolved
	s.cachedAt = time.Now()
	s.mu.Unlock()

	return resolved
}

// resolve reads the database and layers it over the environment.
func (s *SettingsService) resolve(ctx context.Context) mailer.Settings {
	out := mailer.Settings{
		APIKey:    s.cfg.ResendAPIKey,
		FromEmail: s.cfg.ResendFromEmail,
		FromName:  "Orbita",
	}
	// The env default is a placeholder, not a real deliverable address.
	if out.FromEmail == "orbita@localhost" {
		out.FromEmail = ""
	}

	row, err := s.repo.Get(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("settings: falling back to environment")
		return out
	}

	if row.ResendAPIKeyEnc != nil && *row.ResendAPIKeyEnc != "" {
		key, err := s.platformKey()
		if err == nil {
			if plain, err := auth.Decrypt(*row.ResendAPIKeyEnc, key); err == nil {
				out.APIKey = plain
			} else {
				log.Error().Err(err).Msg("settings: could not decrypt the stored Resend key")
			}
		}
	}
	if row.EmailFrom != nil && *row.EmailFrom != "" {
		out.FromEmail = *row.EmailFrom
	}
	if row.EmailFromName != nil && *row.EmailFromName != "" {
		out.FromName = *row.EmailFromName
	}
	return out
}

// GetEmailSettings returns the dashboard view of email configuration.
func (s *SettingsService) GetEmailSettings(ctx context.Context) (*EmailSettings, error) {
	row, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	live := s.MailerSettings(ctx)

	source := "unset"
	switch {
	case row.ResendAPIKeyEnc != nil && *row.ResendAPIKeyEnc != "":
		source = "dashboard"
	case s.cfg.ResendAPIKey != "":
		source = "environment"
	}

	return &EmailSettings{
		Configured:    live.Configured(),
		HasAPIKey:     live.APIKey != "",
		EmailFrom:     live.FromEmail,
		EmailFromName: live.FromName,
		Source:        source,
	}, nil
}

// UpdateEmailSettingsInput carries a settings change. A nil APIKey leaves the
// stored key untouched; an empty string clears it.
type UpdateEmailSettingsInput struct {
	APIKey        *string
	EmailFrom     string
	EmailFromName string
}

func (s *SettingsService) UpdateEmailSettings(ctx context.Context, in UpdateEmailSettingsInput, actor uuid.UUID) error {
	row, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}

	if in.APIKey != nil {
		trimmed := strings.TrimSpace(*in.APIKey)
		if trimmed == "" {
			row.ResendAPIKeyEnc = nil
		} else {
			key, err := s.platformKey()
			if err != nil {
				return err
			}
			enc, err := auth.Encrypt(trimmed, key)
			if err != nil {
				return fmt.Errorf("UpdateEmailSettings: encrypt: %w", err)
			}
			row.ResendAPIKeyEnc = &enc
		}
	}

	from := strings.TrimSpace(in.EmailFrom)
	row.EmailFrom = &from
	name := strings.TrimSpace(in.EmailFromName)
	row.EmailFromName = &name
	row.UpdatedBy = &actor
	row.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, row); err != nil {
		return err
	}

	s.invalidate()
	return nil
}

// invalidate drops the cache so the next send picks up the change immediately.
func (s *SettingsService) invalidate() {
	s.mu.Lock()
	s.cached = nil
	s.mu.Unlock()
}

func (s *SettingsService) platformKey() ([]byte, error) {
	key, err := auth.DerivePlatformKey([]byte(s.cfg.EncryptionMasterKey))
	if err != nil {
		return nil, fmt.Errorf("settings: derive platform key: %w", err)
	}
	return key, nil
}

// EmailConfigured reports whether this instance can send email. Drives the
// credential-handover fallback when provisioning users.
func (s *SettingsService) EmailConfigured(ctx context.Context) bool {
	return s.MailerSettings(ctx).Configured()
}
