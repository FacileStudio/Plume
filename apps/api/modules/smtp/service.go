package smtp

import (
	"context"
	"crypto/tls"
	stderrors "errors"
	"fmt"
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
	body := `<!DOCTYPE html>
<html>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 40px 20px; background: #f4f4f5;">
  <div style="max-width: 480px; margin: 0 auto; background: #fff; border-radius: 8px; padding: 32px; border: 1px solid #e4e4e7;">
    <h2 style="margin: 0 0 16px; color: #18181b;">SMTP Test Successful</h2>
    <p style="margin: 0; color: #3f3f46; line-height: 1.6;">Your SMTP configuration is working correctly. Plume can now send emails on your behalf.</p>
  </div>
</body>
</html>`

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
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 40px 20px; background: #f4f4f5;">
  <div style="max-width: 480px; margin: 0 auto; background: #fff; border-radius: 8px; padding: 32px; border: 1px solid #e4e4e7;">
    <h2 style="margin: 0 0 8px; color: #18181b;">Signature Requested</h2>
    <p style="margin: 0 0 24px; color: #71717a; font-size: 14px;">via Plume</p>
    <p style="margin: 0 0 8px; color: #3f3f46; line-height: 1.6;">Hi %s,</p>
    <p style="margin: 0 0 24px; color: #3f3f46; line-height: 1.6;">You have been invited to review and sign <strong>%s</strong>.</p>
    <div style="text-align: center; margin: 32px 0;">
      <a href="%s" style="display: inline-block; background: #18181b; color: #fff; text-decoration: none; padding: 12px 32px; border-radius: 6px; font-weight: 500; font-size: 15px;">Review &amp; Sign</a>
    </div>
    <p style="margin: 0; color: #a1a1aa; font-size: 13px; line-height: 1.5;">If the button above doesn't work, copy and paste this link into your browser:<br/>%s</p>
  </div>
</body>
</html>`, signerName, documentName, signingURL, signingURL)

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
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 40px 20px; background: #f4f4f5;">
  <div style="max-width: 480px; margin: 0 auto; background: #fff; border-radius: 8px; padding: 32px; border: 1px solid #e4e4e7;">
    <h2 style="margin: 0 0 8px; color: #18181b;">Document %s</h2>
    <p style="margin: 0 0 24px; color: #71717a; font-size: 14px;">via Plume</p>
    <p style="margin: 0 0 8px; color: #3f3f46; line-height: 1.6;">Hi %s,</p>
    <p style="margin: 0 0 24px; color: #3f3f46; line-height: 1.6;"><strong>%s</strong> has %s <strong>%s</strong>.</p>
    <div style="text-align: center; margin: 32px 0;">
      <a href="%s" style="display: inline-block; background: #18181b; color: #fff; text-decoration: none; padding: 12px 32px; border-radius: 6px; font-weight: 500; font-size: 15px;">View Document</a>
    </div>
    <p style="margin: 0; color: #a1a1aa; font-size: 13px; line-height: 1.5;">If the button above doesn't work, copy and paste this link into your browser:<br/>%s</p>
  </div>
</body>
</html>`, action, user.Name, signerName, action, documentName, docURL, docURL)

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
