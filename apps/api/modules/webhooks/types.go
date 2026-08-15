package webhooks

import "time"

// CreateWebhookRequest is the payload for registering a delivery target.
type CreateWebhookRequest struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

// UpdateWebhookRequest is the payload for editing a delivery target's URL,
// secret or enabled state.
type UpdateWebhookRequest struct {
	URL     string `json:"url"`
	Secret  string `json:"secret"`
	Enabled bool   `json:"enabled"`
}

// WebhookResponse is the delivery target as returned to the client, without
// the secret.
type WebhookResponse struct {
	ID         int64   `json:"id"`
	URL        string  `json:"url"`
	Enabled    bool    `json:"enabled"`
	LastSentAt *string `json:"last_sent_at"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// EventPayload is a signed webhook delivery: the event metadata plus the
// document and optional signer it concerns.
type EventPayload struct {
	EventID    string        `json:"event_id"`
	EventType  string        `json:"event_type"`
	OccurredAt time.Time     `json:"occurred_at"`
	Owner      *EventOwner   `json:"owner,omitempty"`
	Document   EventDocument `json:"document"`
	Signer     *EventSigner  `json:"signer,omitempty"`
}

// EventOwner identifies the account that owns the event.
type EventOwner struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// EventDocument is the document state attached to an event.
type EventDocument struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	FileName   string    `json:"file_name"`
	URL        string    `json:"url,omitempty"`
	Sequential bool      `json:"sequential"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// EventSigner is the signer state attached to a signer-scoped event.
type EventSigner struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Role          string     `json:"role"`
	Status        string     `json:"status,omitempty"`
	OrderNum      int        `json:"order_num"`
	SigningURL    string     `json:"signing_url,omitempty"`
	SignedAt      *time.Time `json:"signed_at,omitempty"`
	ViewedAt      *time.Time `json:"viewed_at,omitempty"`
	EmailOpenedAt *time.Time `json:"email_opened_at,omitempty"`
}

const (
	EventDocumentCreated   = "document.created"
	EventDocumentSent      = "document.sent"
	EventDocumentCompleted = "document.completed"
	EventDocumentDeclined  = "document.declined"
	EventDocumentDeleted   = "document.deleted"
	EventSignerAdded       = "signer.added"
	EventSignerEmailOpened = "signer.email_opened"
	EventSignerViewed      = "signer.viewed"
	EventSignerSigned      = "signer.signed"
	EventSignerDeclined    = "signer.declined"
	EventSignerReminded    = "signer.reminded"
)
