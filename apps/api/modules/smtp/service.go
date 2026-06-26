package smtp

import (
	"context"
	"crypto/tls"
	stderrors "errors"
	"fmt"
	"html"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"

	"api/internal/errors"
	"api/schemas"

	"gorm.io/gorm"
)

const smtpDialTimeout = 10 * time.Second

type Service struct {
	orm *gorm.DB
}

func NewService(orm *gorm.DB) *Service {
	return &Service{orm: orm}
}

func (s *Service) getConfig(ctx context.Context, ownerID int64) (*ConfigResponse, error) {
	record, err := s.findConfig(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return toResponse(record), nil
}

func (s *Service) saveConfig(ctx context.Context, ownerID int64, req *SaveConfigRequest) (*ConfigResponse, error) {
	if req.Host == "" {
		return nil, errors.Invalid("host is required")
	}
	if req.Port == 0 {
		return nil, errors.Invalid("port is required")
	}
	if req.FromEmail == "" {
		return nil, errors.Invalid("from_email is required")
	}

	var record schemas.SmtpConfig
	err := s.orm.WithContext(ctx).Where("owner_id = ?", ownerID).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		record = schemas.SmtpConfig{
			OwnerID:   ownerID,
			Host:      req.Host,
			Port:      req.Port,
			Username:  req.Username,
			Password:  req.Password,
			FromEmail: req.FromEmail,
			FromName:  req.FromName,
		}
		if err := s.orm.WithContext(ctx).Create(&record).Error; err != nil {
			return nil, errors.Internal("failed to create smtp config", err)
		}
		return toResponse(&record), nil
	}
	if err != nil {
		return nil, errors.Internal("failed to read smtp config", err)
	}

	record.Host = req.Host
	record.Port = req.Port
	record.Username = req.Username
	if req.Password != "" {
		record.Password = req.Password
	}
	record.FromEmail = req.FromEmail
	record.FromName = req.FromName

	if err := s.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, errors.Internal("failed to update smtp config", err)
	}
	return toResponse(&record), nil
}

func (s *Service) deleteConfig(ctx context.Context, ownerID int64) error {
	result := s.orm.WithContext(ctx).Where("owner_id = ?", ownerID).Delete(&schemas.SmtpConfig{})
	if result.Error != nil {
		return errors.Internal("failed to delete smtp config", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("smtp config not found")
	}
	return nil
}

func (s *Service) testConfig(ctx context.Context, ownerID int64, to string) error {
	if to == "" {
		return errors.Invalid("to is required")
	}

	record, err := s.findConfig(ctx, ownerID)
	if err != nil {
		return err
	}

	subject := "Plume — SMTP Configuration Test"
	body := renderEmail(emailContent{
		domain:  "",
		heading: "SMTP test successful",
		body:    `<p style="margin:0;color:#3f3f46;line-height:1.6;font-size:15px;">Your SMTP configuration is working correctly. Plume can now send emails on your behalf.</p>`,
	})

	if err := sendEmail(record, to, subject, body); err != nil {
		return errors.Invalid(fmt.Sprintf("SMTP test failed: %s", err.Error()))
	}
	return nil
}

func (s *Service) SendSigningEmail(ownerID int64, signerName string, signerEmail string, documentName string, signingToken string, domain string) {
	var record schemas.SmtpConfig
	err := s.orm.Where("owner_id = ?", ownerID).First(&record).Error
	if err != nil {
		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("failed to load smtp config for signing email", slog.Any("error", err))
		}
		return
	}

	signingURL := fmt.Sprintf("%s/share/%s", domain, signingToken)

	subject := fmt.Sprintf("Signature requested — %s", documentName)
	body := renderEmail(emailContent{
		domain:  domain,
		heading: "Signature requested",
		body: fmt.Sprintf(`<p style="margin:0 0 14px;color:#3f3f46;line-height:1.6;font-size:15px;">Hi %s,</p>
    <p style="margin:0;color:#3f3f46;line-height:1.6;font-size:15px;">You have been invited to review and sign <strong style="color:#18181b;">%s</strong>.</p>`,
			html.EscapeString(signerName), html.EscapeString(documentName)),
		ctaLabel: "Review &amp; Sign",
		ctaURL:   signingURL,
	})

	if err := sendEmail(&record, signerEmail, subject, body); err != nil {
		slog.Error("failed to send signing email",
			slog.String("signer_email", signerEmail),
			slog.Int64("owner_id", ownerID),
			slog.Any("error", err),
		)
	}
}

func (s *Service) SendNotificationEmail(ownerID int64, docID int64, signerName string, documentName string, eventType string, domain string) {
	var config schemas.SmtpConfig
	err := s.orm.Where("owner_id = ?", ownerID).First(&config).Error
	if err != nil {
		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("failed to load smtp config for notification email", slog.Any("error", err))
		}
		return
	}

	var user schemas.User
	if err := s.orm.Where("id = ?", ownerID).First(&user).Error; err != nil {
		slog.Error("failed to load owner for notification email", slog.Any("error", err))
		return
	}

	var action string
	if eventType == "signed" {
		action = "signed"
	} else {
		action = "declined"
	}

	subject := fmt.Sprintf("Document %s — %s", action, documentName)
	docURL := fmt.Sprintf("%s/documents/%d", domain, docID)
	body := renderEmail(emailContent{
		domain:  domain,
		heading: fmt.Sprintf("Document %s", action),
		body: fmt.Sprintf(`<p style="margin:0 0 14px;color:#3f3f46;line-height:1.6;font-size:15px;">Hi %s,</p>
    <p style="margin:0;color:#3f3f46;line-height:1.6;font-size:15px;"><strong style="color:#18181b;">%s</strong> has %s <strong style="color:#18181b;">%s</strong>.</p>`,
			html.EscapeString(user.Name), html.EscapeString(signerName), action, html.EscapeString(documentName)),
		ctaLabel: "View document",
		ctaURL:   docURL,
	})

	if err := sendEmail(&config, user.Email, subject, body); err != nil {
		slog.Error("failed to send notification email",
			slog.String("owner_email", user.Email),
			slog.Int64("owner_id", ownerID),
			slog.Any("error", err),
		)
	}
}

func (s *Service) findConfig(ctx context.Context, ownerID int64) (*schemas.SmtpConfig, error) {
	var record schemas.SmtpConfig
	err := s.orm.WithContext(ctx).Where("owner_id = ?", ownerID).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("smtp config not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to read smtp config", err)
	}
	return &record, nil
}

type emailContent struct {
	domain   string
	heading  string
	body     string
	ctaLabel string
	ctaURL   string
}

func renderEmail(c emailContent) string {
	fontStack := "-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif"

	var brand string
	if domain := strings.TrimRight(c.domain, "/"); domain != "" {
		brand = fmt.Sprintf(`<img src="%s/logo.png" width="44" height="44" alt="Plume" style="display:block;border:0;outline:none;text-decoration:none;" />`, domain)
	} else {
		brand = fmt.Sprintf(`<span style="font-family:%s;font-size:20px;font-weight:700;letter-spacing:-0.02em;color:#18181b;">Plume</span>`, fontStack)
	}

	var cta string
	if c.ctaURL != "" {
		cta = fmt.Sprintf(`
        <tr><td style="padding-top:28px;">
          <table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto;">
            <tr><td align="center" style="border-radius:8px;background:#18181b;">
              <a href="%s" style="display:inline-block;padding:13px 34px;font-family:%s;font-size:15px;font-weight:600;color:#ffffff;text-decoration:none;border-radius:8px;">%s</a>
            </td></tr>
          </table>
        </td></tr>
        <tr><td style="padding-top:28px;">
          <p style="margin:0 0 6px;font-family:%s;color:#a1a1aa;font-size:13px;line-height:1.5;">Or copy and paste this link into your browser:</p>
          <a href="%s" style="font-family:%s;color:#71717a;font-size:13px;line-height:1.5;word-break:break-all;overflow-wrap:anywhere;">%s</a>
        </td></tr>`, c.ctaURL, fontStack, c.ctaLabel, fontStack, c.ctaURL, fontStack, c.ctaURL)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width,initial-scale=1.0" />
</head>
<body style="margin:0;padding:0;background:#f4f4f5;-webkit-font-smoothing:antialiased;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:#f4f4f5;">
    <tr><td align="center" style="padding:40px 16px;">
      <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="max-width:480px;width:100%%;">
        <tr><td align="center" style="padding-bottom:24px;">%s</td></tr>
        <tr><td style="background:#ffffff;border:1px solid #e4e4e7;border-radius:14px;padding:32px;">
          <table role="presentation" width="100%%" cellpadding="0" cellspacing="0">
            <tr><td>
              <h1 style="margin:0 0 4px;font-family:%s;font-size:20px;font-weight:700;letter-spacing:-0.01em;color:#18181b;">%s</h1>
              <p style="margin:0 0 20px;font-family:%s;color:#a1a1aa;font-size:13px;">via Plume</p>
              %s
            </td></tr>%s
          </table>
        </td></tr>
        <tr><td align="center" style="padding-top:20px;">
          <p style="margin:0;font-family:%s;color:#a1a1aa;font-size:12px;line-height:1.5;">Sent securely by Plume · Self-hosted document signing</p>
        </td></tr>
      </table>
    </td></tr>
  </table>
</body>
</html>`, brand, fontStack, html.EscapeString(c.heading), fontStack, c.body, cta, fontStack)
}

func sanitizeHeader(s string) string {
	r := strings.NewReplacer("\r\n", "", "\r", "", "\n", "")
	return r.Replace(s)
}

func sendEmail(config *schemas.SmtpConfig, to string, subject string, htmlBody string) error {
	addr := net.JoinHostPort(config.Host, fmt.Sprintf("%d", config.Port))

	dialer := &net.Dialer{Timeout: smtpDialTimeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("connect %s: %w", addr, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("smtp handshake: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: config.Host}); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if config.Username != "" {
		auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication: %w", err)
		}
	}

	if err := client.Mail(sanitizeHeader(config.FromEmail)); err != nil {
		return fmt.Errorf("from %s: %w", config.FromEmail, err)
	}
	if err := client.Rcpt(sanitizeHeader(to)); err != nil {
		return fmt.Errorf("recipient %s: %w", to, err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}

	headers := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n",
		sanitizeHeader(config.FromName), sanitizeHeader(config.FromEmail), sanitizeHeader(to), sanitizeHeader(subject))

	if _, err := writer.Write([]byte(headers + htmlBody)); err != nil {
		return fmt.Errorf("write body: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}

	return client.Quit()
}

func toResponse(record *schemas.SmtpConfig) *ConfigResponse {
	return &ConfigResponse{
		Host:      record.Host,
		Port:      record.Port,
		Username:  record.Username,
		FromEmail: record.FromEmail,
		FromName:  record.FromName,
		UpdatedAt: record.UpdatedAt.Format(time.RFC3339),
	}
}
