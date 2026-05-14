package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"api/internal/errors"
	"api/schemas"

	"gorm.io/gorm"
)

type Service struct {
	orm *gorm.DB
}

func NewService(orm *gorm.DB) *Service {
	return &Service{orm: orm}
}

func (s *Service) create(ctx context.Context, ownerID int64, req *CreateWebhookRequest) (*WebhookResponse, error) {
	if req.URL == "" {
		return nil, errors.Invalid("url is required")
	}

	record := &schemas.Webhook{
		OwnerID: ownerID,
		URL:     req.URL,
		Secret:  req.Secret,
		Enabled: true,
	}
	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, errors.Internal("failed to create webhook", err)
	}

	return toResponse(record), nil
}

func (s *Service) list(ctx context.Context, ownerID int64) ([]WebhookResponse, error) {
	var records []schemas.Webhook
	if err := s.orm.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at desc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list webhooks", err)
	}

	out := make([]WebhookResponse, len(records))
	for i := range records {
		out[i] = *toResponse(&records[i])
	}
	return out, nil
}

func (s *Service) get(ctx context.Context, ownerID int64, webhookID int64) (*WebhookResponse, error) {
	record, err := s.findWebhook(ctx, ownerID, webhookID)
	if err != nil {
		return nil, err
	}
	return toResponse(record), nil
}

func (s *Service) update(ctx context.Context, ownerID int64, webhookID int64, req *UpdateWebhookRequest) (*WebhookResponse, error) {
	record, err := s.findWebhook(ctx, ownerID, webhookID)
	if err != nil {
		return nil, err
	}

	if req.URL == "" {
		return nil, errors.Invalid("url is required")
	}

	record.URL = req.URL
	if req.Secret != "" {
		record.Secret = req.Secret
	}
	record.Enabled = req.Enabled

	if err := s.orm.WithContext(ctx).Save(record).Error; err != nil {
		return nil, errors.Internal("failed to update webhook", err)
	}
	return toResponse(record), nil
}

func (s *Service) delete(ctx context.Context, ownerID int64, webhookID int64) error {
	record, err := s.findWebhook(ctx, ownerID, webhookID)
	if err != nil {
		return err
	}

	if err := s.orm.WithContext(ctx).Delete(record).Error; err != nil {
		return errors.Internal("failed to delete webhook", err)
	}
	return nil
}

func (s *Service) findWebhook(ctx context.Context, ownerID int64, webhookID int64) (*schemas.Webhook, error) {
	var record schemas.Webhook
	err := s.orm.WithContext(ctx).Where("id = ? AND owner_id = ?", webhookID, ownerID).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("webhook not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to read webhook", err)
	}
	return &record, nil
}

func (s *Service) Dispatch(ownerID int64, payload EventPayload) {
	var webhooks []schemas.Webhook
	if err := s.orm.Where("owner_id = ? AND enabled = ?", ownerID, true).Find(&webhooks).Error; err != nil {
		slog.Error("failed to load webhooks for dispatch", slog.Any("error", err))
		return
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal webhook payload", slog.Any("error", err))
		return
	}

	for i := range webhooks {
		go deliverWebhook(&webhooks[i], body)
	}
}

func deliverWebhook(wh *schemas.Webhook, body []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	signature := sign(wh.Secret, body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(body))
	if err != nil {
		slog.Error("failed to build webhook request", slog.Int64("webhook_id", wh.ID), slog.Any("error", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Plume-Webhook/1.0")
	req.Header.Set("x-plume-signature-256", fmt.Sprintf("sha256=%s", signature))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to deliver webhook", slog.Int64("webhook_id", wh.ID), slog.Any("error", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		slog.Warn("webhook endpoint returned error", slog.Int64("webhook_id", wh.ID), slog.Int("status", resp.StatusCode))
	}
}

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func toResponse(w *schemas.Webhook) *WebhookResponse {
	resp := &WebhookResponse{
		ID:        w.ID,
		URL:       w.URL,
		Enabled:   w.Enabled,
		CreatedAt: w.CreatedAt.Format(time.RFC3339),
		UpdatedAt: w.UpdatedAt.Format(time.RFC3339),
	}
	if w.LastSentAt != nil {
		formatted := w.LastSentAt.Format(time.RFC3339)
		resp.LastSentAt = &formatted
	}
	return resp
}
