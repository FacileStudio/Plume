package reminders

import (
	"context"
	stderrors "errors"
	"log/slog"
	"strconv"
	"time"

	"api/internal/errors"
	"api/modules/smtp"
	"api/schemas"

	"gorm.io/gorm"
)

const (
	manualResendCooldown = time.Minute
	minSignerAge         = 24 * time.Hour
	tickInterval         = time.Hour
)

type Service struct {
	orm     *gorm.DB
	smtpSvc *smtp.Service
	domain  string
	now     func() time.Time
}

func NewService(orm *gorm.DB, smtpSvc *smtp.Service, domain string) *Service {
	return &Service{
		orm:     orm,
		smtpSvc: smtpSvc,
		domain:  domain,
		now:     func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) RemindSigner(ctx context.Context, ownerID string, signerID string) (*RemindResponse, error) {
	uid, err := strconv.ParseInt(ownerID, 10, 64)
	if err != nil {
		return nil, errors.NotFound("signer not found")
	}
	sid, err := strconv.ParseInt(signerID, 10, 64)
	if err != nil {
		return nil, errors.NotFound("signer not found")
	}

	var signer schemas.Signer
	if err := s.orm.WithContext(ctx).Where("id = ?", sid).First(&signer).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("signer not found")
		}
		return nil, errors.Internal("failed to read signer", err)
	}

	var doc schemas.Document
	if err := s.orm.WithContext(ctx).Where("id = ? AND owner_id = ?", signer.DocumentID, uid).First(&doc).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("signer not found")
		}
		return nil, errors.Internal("failed to verify document ownership", err)
	}

	if signer.Status != "pending" {
		return nil, errors.Invalid("only pending signers can be reminded")
	}
	if doc.Status != "pending" {
		return nil, errors.Invalid("document is not pending signature")
	}
	if signer.Token == "" {
		return nil, errors.Invalid("signer has no signing link")
	}

	now := s.now()
	if signer.LastRemindedAt != nil && now.Sub(*signer.LastRemindedAt) < manualResendCooldown {
		return nil, errors.New("resource_exhausted", "please wait before resending the reminder", nil)
	}

	if err := s.orm.WithContext(ctx).
		Model(&schemas.Signer{}).
		Where("id = ?", signer.ID).
		Update("last_reminded_at", now).Error; err != nil {
		return nil, errors.Internal("failed to update reminder timestamp", err)
	}

	go s.smtpSvc.SendSigningEmail(uid, signer.Name, signer.Email, doc.Name, signer.Token, s.domain)

	return &RemindResponse{Status: "ok", RemindedAt: now}, nil
}

func (s *Service) RunOnce(ctx context.Context) (int, error) {
	now := s.now()

	var users []schemas.User
	if err := s.orm.WithContext(ctx).
		Where("reminder_interval_days > 0").
		Find(&users).Error; err != nil {
		return 0, err
	}

	sent := 0
	for i := range users {
		user := users[i]
		count, err := s.runForUser(ctx, &user, now)
		if err != nil {
			slog.Warn("reminder run failed for user",
				slog.Int64("user_id", user.ID),
				slog.Any("error", err),
			)
			continue
		}
		sent += count
	}
	return sent, nil
}

func (s *Service) runForUser(ctx context.Context, user *schemas.User, now time.Time) (int, error) {
	interval := time.Duration(user.ReminderIntervalDays) * 24 * time.Hour
	cutoff := now.Add(-interval)
	minAgeCutoff := now.Add(-minSignerAge)

	type row struct {
		SignerID       int64
		SignerName     string
		SignerEmail    string
		SignerToken    string
		SignerCreated  time.Time
		SignerReminded *time.Time
		DocumentName   string
	}

	var rows []row
	err := s.orm.WithContext(ctx).
		Table("signers s").
		Select(`s.id as signer_id,
			s.name as signer_name,
			s.email as signer_email,
			s.token as signer_token,
			s.created_at as signer_created,
			s.last_reminded_at as signer_reminded,
			d.name as document_name`).
		Joins("join documents d on d.id = s.document_id").
		Where("d.owner_id = ?", user.ID).
		Where("d.status = ?", "pending").
		Where("s.status = ?", "pending").
		Where("s.token <> ''").
		Where("s.role <> ?", "viewer").
		Where("s.created_at <= ?", minAgeCutoff).
		Where("s.last_reminded_at IS NULL OR s.last_reminded_at < ?", cutoff).
		Scan(&rows).Error
	if err != nil {
		return 0, err
	}

	sent := 0
	for _, r := range rows {
		if !shouldRemind(r.SignerCreated, r.SignerReminded, user.ReminderIntervalDays, now) {
			continue
		}
		if err := s.orm.WithContext(ctx).
			Model(&schemas.Signer{}).
			Where("id = ?", r.SignerID).
			Update("last_reminded_at", now).Error; err != nil {
			slog.Warn("failed to update reminder timestamp",
				slog.Int64("signer_id", r.SignerID),
				slog.Any("error", err),
			)
			continue
		}
		go s.smtpSvc.SendSigningEmail(user.ID, r.SignerName, r.SignerEmail, r.DocumentName, r.SignerToken, s.domain)
		sent++
	}
	return sent, nil
}

func shouldRemind(createdAt time.Time, lastRemindedAt *time.Time, intervalDays int, now time.Time) bool {
	if intervalDays <= 0 {
		return false
	}
	if now.Sub(createdAt) < minSignerAge {
		return false
	}
	if lastRemindedAt == nil {
		return true
	}
	interval := time.Duration(intervalDays) * 24 * time.Hour
	return now.Sub(*lastRemindedAt) >= interval
}

func Start(ctx context.Context, service *Service, logger *slog.Logger) {
	go service.loop(ctx, logger)
}

func (s *Service) loop(ctx context.Context, logger *slog.Logger) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("reminder ticker stopped")
			return
		case <-ticker.C:
			runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			count, err := s.RunOnce(runCtx)
			cancel()
			if err != nil {
				logger.Warn("reminder run failed", slog.Any("error", err))
				continue
			}
			logger.Info("reminder run complete", slog.Int("sent", count))
		}
	}
}
