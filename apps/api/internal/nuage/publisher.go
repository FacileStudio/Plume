package nuage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const (
	maxAttempts    = 5
	maxConcurrent  = 4
	defaultBackoff = 2 * time.Second
)

var errNoSpace = errors.New("document has no space and NUAGE_DEFAULT_SPACE_FACILE_ID is not set")

// Job is one signed document to archive. Load is called inside the worker so
// the PDF is read off the request path.
type Job struct {
	DocumentID    int64
	Title         string
	SpaceFacileID string
	SpaceName     string
	LinkedTo      string
	OnPublished   func(spaceFacileID, linkedTo string)
	Load          func() ([]byte, error)
}

// Publisher archives signed documents in Nuage in the background. A nil
// Publisher is valid and does nothing.
type Publisher struct {
	client         *Client
	defaultSpaceID string
	defaultName    string
	log            *slog.Logger
	backoff        time.Duration
	slots          chan struct{}
}

// NewPublisher wires a client to the default space used for documents that
// belong to no Plume space.
func NewPublisher(client *Client, defaultSpaceID, defaultName string, log *slog.Logger) *Publisher {
	return &Publisher{
		client:         client,
		defaultSpaceID: defaultSpaceID,
		defaultName:    defaultName,
		log:            log,
		backoff:        defaultBackoff,
		slots:          make(chan struct{}, maxConcurrent),
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
		if t, err := resolveTarget(job, p.defaultSpaceID, p.defaultName); err == nil {
			job.OnPublished(t.SpaceFacileID, t.LinkedTo)
		}
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
	target, err := resolveTarget(job, p.defaultSpaceID, p.defaultName)
	if err != nil {
		return err
	}
	data, err := job.Load()
	if err != nil {
		return fmt.Errorf("read final pdf: %w", err)
	}
	spaceID, err := p.client.EnsureSpace(ctx, target.SpaceFacileID, target.SpaceName)
	if err != nil {
		return fmt.Errorf("provision space: %w", err)
	}
	if err := p.client.Upload(ctx, spaceID, filename(job.Title), target.LinkedTo, data); err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	return nil
}

// Target is where a job lands in Nuage: a space and the linked_to key that
// makes a retry idempotent.
type Target struct {
	SpaceFacileID string
	SpaceName     string
	LinkedTo      string
}

// resolveTarget picks the Nuage space and idempotency key for a job: the
// job's own space and linked_to when set, else the default space and
// plume:document:<id>. The two fall back independently.
func resolveTarget(job Job, defaultSpaceID, defaultName string) (Target, error) {
	t := Target{SpaceFacileID: job.SpaceFacileID, SpaceName: job.SpaceName, LinkedTo: job.LinkedTo}
	if t.SpaceFacileID == "" {
		t.SpaceFacileID, t.SpaceName = defaultSpaceID, defaultName
	}
	if t.SpaceFacileID == "" {
		return Target{}, errNoSpace
	}
	if t.SpaceName == "" {
		t.SpaceName = defaultName
	}
	if t.LinkedTo == "" {
		t.LinkedTo = fmt.Sprintf("plume:document:%d", job.DocumentID)
	}
	return t, nil
}

func retryable(err error) bool {
	if errors.Is(err, errNoSpace) {
		return false
	}
	var se *StatusError
	if errors.As(err, &se) {
		return se.Retryable()
	}
	return true
}

func filename(title string) string {
	title = strings.TrimSpace(strings.NewReplacer("/", "-", "\\", "-", "\"", "").Replace(title))
	if title == "" {
		title = "document"
	}
	if !strings.HasSuffix(strings.ToLower(title), ".pdf") {
		title += ".pdf"
	}
	return title
}
