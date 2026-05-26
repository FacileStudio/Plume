package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"api/internal/errors"
	"api/internal/urlsafe"
	"api/schemas"

	"gorm.io/gorm"
)

const (
	deliveryTimeout = 15 * time.Second
	maxAttempts     = 3
)

var retryBackoffs = []time.Duration{
	1 * time.Second,
	5 * time.Second,
	20 * time.Second,
}

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
	if err := urlsafe.Validate(req.URL); err != nil {
		return nil, errors.Invalid(err.Error())
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
	if err := urlsafe.Validate(req.URL); err != nil {
		return nil, errors.Invalid(err.Error())
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

func (s *Service) Test(ctx context.Context, ownerID int64, webhookID int64) error {
	record, err := s.findWebhook(ctx, ownerID, webhookID)
	if err != nil {
		return err
	}

	payload := EventPayload{
		EventID:    newEventID(),
		EventType:  "test.ping",
		OccurredAt: time.Now().UTC(),
		Document: EventDocument{
			ID:        0,
			Name:      "Sample document",
			FileName:  "sample.pdf",
			Status:    "draft",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return errors.Internal("failed to marshal test payload", err)
	}

	if err := s.deliverOnce(ctx, record, body); err != nil {
		return errors.Invalid("delivery failed — check the URL and try again")
	}
	s.markDelivered(record.ID)
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
	if payload.EventID == "" {
		payload.EventID = newEventID()
	}
	if payload.OccurredAt.IsZero() {
		payload.OccurredAt = time.Now().UTC()
	}

	var webhooks []schemas.Webhook
	if err := s.orm.Where("owner_id = ? AND enabled = ?", ownerID, true).Find(&webhooks).Error; err != nil {
		slog.Error("failed to load webhooks for dispatch", slog.Any("error", err))
		return
	}
	if len(webhooks) == 0 {
		return
	}

	if payload.Owner == nil {
		var owner schemas.User
		if err := s.orm.Where("id = ?", ownerID).First(&owner).Error; err == nil {
			payload.Owner = &EventOwner{ID: owner.ID, Name: owner.Name, Email: owner.Email}
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal webhook payload",
			slog.String("event_type", payload.EventType),
			slog.Any("error", err),
		)
		return
	}

	for i := range webhooks {
		wh := webhooks[i]
		go s.deliverWithRetry(&wh, payload.EventType, body)
	}
}

func (s *Service) deliverWithRetry(wh *schemas.Webhook, eventType string, body []byte) {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), deliveryTimeout)
		err := s.deliverOnce(ctx, wh, body)
		cancel()
		if err == nil {
			s.markDelivered(wh.ID)
			return
		}
		slog.Warn("webhook delivery attempt failed",
			slog.Int64("webhook_id", wh.ID),
			slog.String("event_type", eventType),
			slog.Int("attempt", attempt+1),
			slog.Any("error", err),
		)
		if attempt < maxAttempts-1 {
			time.Sleep(retryBackoffs[attempt])
		}
	}
	slog.Error("webhook delivery exhausted retries",
		slog.Int64("webhook_id", wh.ID),
		slog.String("event_type", eventType),
	)
}

func (s *Service) deliverOnce(ctx context.Context, wh *schemas.Webhook, body []byte) error {
	if err := urlsafe.Validate(wh.URL); err != nil {
		return fmt.Errorf("unsafe URL: %w", err)
	}

	signature := sign(wh.Secret, body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Plume-Webhook/1.0")
	req.Header.Set("x-plume-signature-256", fmt.Sprintf("sha256=%s", signature))

	client := &http.Client{
		Timeout:   deliveryTimeout,
		Transport: urlsafe.SafeTransport(),
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("dispatch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("receiver returned %d", resp.StatusCode)
}

func (s *Service) markDelivered(webhookID int64) {
	now := time.Now().UTC()
	if err := s.orm.Model(&schemas.Webhook{}).
		Where("id = ?", webhookID).
		Update("last_sent_at", now).Error; err != nil {
		slog.Warn("failed to update webhook last_sent_at",
			slog.Int64("webhook_id", webhookID),
			slog.Any("error", err),
		)
	}
}

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func newEventID() string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("evt_%d", time.Now().UnixNano())
	}
	return "evt_" + hex.EncodeToString(bytes)
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
