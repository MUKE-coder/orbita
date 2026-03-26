package mailer

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"
)

type Mailer struct {
	client    *resend.Client
	fromEmail string
}

func New(apiKey, fromEmail string) *Mailer {
	return &Mailer{
		client:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
	}
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

	_, err := m.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    fmt.Sprintf("Orbita <%s>", m.fromEmail),
		To:      []string{to},
		Subject: "Verify your email — Orbita",
		Html:    html,
	})
	if err != nil {
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

	_, err := m.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    fmt.Sprintf("Orbita <%s>", m.fromEmail),
		To:      []string{to},
		Subject: "Password Reset Code — Orbita",
		Html:    html,
	})
	if err != nil {
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

	_, err := m.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    fmt.Sprintf("Orbita <%s>", m.fromEmail),
		To:      []string{to},
		Subject: fmt.Sprintf("Invitation to join %s — Orbita", orgName),
		Html:    html,
	})
	if err != nil {
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

	_, err := m.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    fmt.Sprintf("Orbita <%s>", m.fromEmail),
		To:      []string{to},
		Subject: fmt.Sprintf("Deploy %s: %s — Orbita", status, appName),
		Html:    html,
	})
	if err != nil {
		return fmt.Errorf("SendDeployNotification: %w", err)
	}
	return nil
}
