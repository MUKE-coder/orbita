package mailer

import (
	"context"
	"errors"
	"fmt"

	"github.com/resend/resend-go/v2"
)

// ErrNotConfigured is returned when no email provider is set up. Callers treat
// it as "this instance doesn't do email" rather than as a failure — the super
// admin hands out credentials directly instead.
var ErrNotConfigured = errors.New("email is not configured on this server")

// Settings is the provider config used for one send.
type Settings struct {
	APIKey    string
	FromEmail string
	FromName  string
}

// Configured reports whether email can actually be sent.
func (s Settings) Configured() bool {
	return s.APIKey != "" && s.FromEmail != ""
}

// from renders the RFC 5322 From header.
func (s Settings) from() string {
	name := s.FromName
	if name == "" {
		name = "Orbita"
	}
	return fmt.Sprintf("%s <%s>", name, s.FromEmail)
}

// Mailer sends transactional email via Resend.
//
// Config is resolved per send rather than captured at construction, because the
// super admin can change the API key and from-address from the dashboard at
// runtime — a mailer holding a client from boot would keep using stale creds
// (or none) until the process restarted.
type Mailer struct {
	resolve func(context.Context) Settings
}

// New builds a mailer over a live settings source.
func New(resolve func(context.Context) Settings) *Mailer {
	return &Mailer{resolve: resolve}
}

// NewStatic builds a mailer over fixed credentials. Used before the database is
// available, and in tests.
func NewStatic(apiKey, fromEmail, fromName string) *Mailer {
	s := Settings{APIKey: apiKey, FromEmail: fromEmail, FromName: fromName}
	return &Mailer{resolve: func(context.Context) Settings { return s }}
}

// settings returns the current config, or ErrNotConfigured.
func (m *Mailer) settings(ctx context.Context) (Settings, error) {
	if m == nil || m.resolve == nil {
		return Settings{}, ErrNotConfigured
	}
	s := m.resolve(ctx)
	if !s.Configured() {
		return Settings{}, ErrNotConfigured
	}
	return s, nil
}

// IsConfigured reports whether this instance can send email. Drives the
// dashboard's "email configured" status and the credential-handover fallback.
func (m *Mailer) IsConfigured(ctx context.Context) bool {
	_, err := m.settings(ctx)
	return err == nil
}

// send delivers one message using the current settings.
func (m *Mailer) send(ctx context.Context, to, subject, html string) error {
	s, err := m.settings(ctx)
	if err != nil {
		return err
	}
	_, err = resend.NewClient(s.APIKey).Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    s.from(),
		To:      []string{to},
		Subject: subject,
		Html:    html,
	})
	return err
}

// SendTest verifies the configured credentials actually work, so the operator
// finds out on the settings page rather than when a real invite goes missing.
func (m *Mailer) SendTest(ctx context.Context, to string) error {
	html := `
		<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #1a1a2e;">Email is working</h2>
			<p>This is a test message from your Orbita instance. If you're reading
			it, invitations and password resets will reach your users.</p>
		</div>
	`
	if err := m.send(ctx, to, "Orbita test email", html); err != nil {
		return fmt.Errorf("SendTest: %w", err)
	}
	return nil
}

func (m *Mailer) SendEmailVerification(ctx context.Context, to, name, verifyURL string) error {
	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #1a1a2e;">Welcome to Orbita, %s!</h2>
			<p>Please verify your email address by clicking the button below:</p>
			<a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #6366f1; color: white; text-decoration: none; border-radius: 6px; margin: 16px 0;">Verify Email</a>
			<p style="color: #666; font-size: 14px;">If you didn't create an account, you can safely ignore this email.</p>
			<p style="color: #666; font-size: 14px;">This link expires in 24 hours.</p>
		</div>
	`, name, verifyURL)

	if err := m.send(ctx, to, "Verify your email — Orbita", html); err != nil {
		return fmt.Errorf("SendEmailVerification: %w", err)
	}
	return nil
}

func (m *Mailer) SendPasswordReset(ctx context.Context, to, name, otp string) error {
	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #1a1a2e;">Password Reset</h2>
			<p>Hi %s, use the code below to reset your password:</p>
			<div style="font-size: 32px; font-weight: bold; letter-spacing: 8px; padding: 16px; background: #f5f5f5; border-radius: 8px; text-align: center; margin: 16px 0;">%s</div>
			<p style="color: #666; font-size: 14px;">This code expires in 10 minutes.</p>
			<p style="color: #666; font-size: 14px;">If you didn't request this, you can safely ignore this email.</p>
		</div>
	`, name, otp)

	if err := m.send(ctx, to, "Password Reset Code — Orbita", html); err != nil {
		return fmt.Errorf("SendPasswordReset: %w", err)
	}
	return nil
}

func (m *Mailer) SendInvite(ctx context.Context, to, orgName, inviterName, acceptURL string) error {
	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #1a1a2e;">You've been invited!</h2>
			<p>%s has invited you to join <strong>%s</strong> on Orbita.</p>
			<a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #6366f1; color: white; text-decoration: none; border-radius: 6px; margin: 16px 0;">Accept Invitation</a>
			<p style="color: #666; font-size: 14px;">This invitation expires in 72 hours.</p>
		</div>
	`, inviterName, orgName, acceptURL)

	if err := m.send(ctx, to, fmt.Sprintf("Invitation to join %s — Orbita", orgName), html); err != nil {
		return fmt.Errorf("SendInvite: %w", err)
	}
	return nil
}

func (m *Mailer) SendDeployNotification(ctx context.Context, to, appName, status, orgName string) error {
	statusColor := "#22c55e"
	if status == "failed" {
		statusColor = "#ef4444"
	}

	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #1a1a2e;">Deployment %s</h2>
			<p>App <strong>%s</strong> in <strong>%s</strong> has been deployed.</p>
			<div style="display: inline-block; padding: 4px 12px; background-color: %s; color: white; border-radius: 4px; font-weight: bold;">%s</div>
		</div>
	`, status, appName, orgName, statusColor, status)

	if err := m.send(ctx, to, fmt.Sprintf("Deploy %s: %s — Orbita", status, appName), html); err != nil {
		return fmt.Errorf("SendDeployNotification: %w", err)
	}
	return nil
}
