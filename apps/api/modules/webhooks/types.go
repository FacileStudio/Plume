package webhooks

type CreateWebhookRequest struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

type UpdateWebhookRequest struct {
	URL     string `json:"url"`
	Secret  string `json:"secret"`
	Enabled bool   `json:"enabled"`
}

type WebhookResponse struct {
	ID         int64   `json:"id"`
	URL        string  `json:"url"`
	Enabled    bool    `json:"enabled"`
	LastSentAt *string `json:"last_sent_at"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type EventPayload struct {
	EventType string       `json:"event_type"`
	Document  EventDocument `json:"document"`
	Signer    *EventSigner  `json:"signer,omitempty"`
}

type EventDocument struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	FileName string `json:"file_name"`
}

type EventSigner struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
