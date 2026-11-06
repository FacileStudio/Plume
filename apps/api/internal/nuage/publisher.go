package nuage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	maxAttempts    = 5
	maxConcurrent  = 4
	defaultBackoff = 2 * time.Second
)

// TokenSource mints the Nuage token a job acts with, for one Plume user.
type TokenSource interface {
	Token(ctx context.Context, userID int64) (string, error)
}

// Job is one signed document to archive in its owner's Plume folder. Load is
// called inside the worker so the PDF is read off the request path.
type Job struct {
	DocumentID  int64
	OwnerID     int64
	Title       string
	LinkedTo    string
	OnPublished func(linkedTo string)
	Load        func() ([]byte, error)
}

// Publisher archives signed documents in Nuage in the background. A nil
// Publisher is valid and does nothing.
type Publisher struct {
	client  *Client
	tokens  TokenSource
	log     *slog.Logger
	backoff time.Duration
	slots   chan struct{}
}

// NewPublisher wires a client to the source of delegated tokens.
func NewPublisher(client *Client, tokens TokenSource, log *slog.Logger) *Publisher {
	return &Publisher{
		client:  client,
		tokens:  tokens,
		log:     log,
		backoff: defaultBackoff,
		slots:   make(chan struct{}, maxConcurrent),
	}
}

// Enqueue starts the publish in a goroutine and returns at once. Failures are
// logged and never reach the caller.
func (p *Publisher) Enqueue(job Job) {
	if p == nil {
		return
	}
	go p.run(context.Background(), job)
}

func (p *Publisher) run(ctx context.Context, job Job) {
	select {
	case p.slots <- struct{}{}:
		defer func() { <-p.slots }()
	case <-ctx.Done():
		return
	}

	log := p.log.With(slog.Int64("document_id", job.DocumentID))
	if err := p.publishWithRetry(ctx, job, log); err != nil {
		log.Error("nuage publish failed", slog.Any("error", err))
		return
	}
	if job.OnPublished != nil {
		job.OnPublished(objectKey(job))
	}
	log.Info("signed document published to nuage")
}

func (p *Publisher) publishWithRetry(ctx context.Context, job Job, log *slog.Logger) error {
	var err error
	delay := p.backoff
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = p.publish(ctx, job)
		if err == nil || !retryable(err) {
			return err
		}
		if attempt == maxAttempts {
			break
		}
		log.Warn("nuage publish attempt failed", slog.Int("attempt", attempt), slog.Any("error", err))
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
		delay *= 2
	}
	return fmt.Errorf("giving up after %d attempts: %w", maxAttempts, err)
}

func (p *Publisher) publish(ctx context.Context, job Job) error {
	data, err := job.Load()
	if err != nil {
		return fmt.Errorf("read final pdf: %w", err)
	}
	token, err := p.tokens.Token(ctx, job.OwnerID)
	if err != nil {
		return fmt.Errorf("delegated token: %w", err)
	}
	if err := p.client.Put(ctx, token, objectKey(job), filename(job.Title), data); err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	return nil
}

// objectKey is the key the PDF is stored under: the job's linked_to, else
// plume:document:<id>. Storing under the same key replaces the object, so a
// retry never duplicates the file.
func objectKey(job Job) string {
	if job.LinkedTo != "" {
		return job.LinkedTo
	}
	return fmt.Sprintf("plume:document:%d", job.DocumentID)
}

func retryable(err error) bool {
	if errors.Is(err, ErrNotConnected) {
		return false
	}
	var se *StatusError
	if errors.As(err, &se) {
		return se.Retryable()
	}
	return true
}

const maxNameBytes = 255

func filename(title string) string {
	title = strings.Map(func(r rune) rune {
		switch {
		case r == '/' || r == '\\':
			return '-'
		case r == '"' || unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zl, unicode.Zp):
			return -1
		}
		return r
	}, title)
	title = strings.TrimSpace(title)
	if title == "" {
		title = "document"
	}
	if !strings.HasSuffix(strings.ToLower(title), ".pdf") {
		title += ".pdf"
	}
	return truncateName(title)
}

func truncateName(name string) string {
	if len(name) <= maxNameBytes {
		return name
	}
	const suffix = ".pdf"
	head := name[:len(name)-len(suffix)]
	head = head[:maxNameBytes-len(suffix)]
	for !utf8.ValidString(head) {
		head = head[:len(head)-1]
	}
	return strings.TrimSpace(head) + suffix
}
