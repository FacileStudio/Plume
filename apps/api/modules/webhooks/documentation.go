package webhooks

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "webhooks",
	Description: "Manage outgoing webhooks. Each delivery is signed with HMAC-SHA256 via the `x-plume-signature-256` header (format: `sha256=<hex>`). Failed deliveries retry up to 3 times with exponential backoff. Event types: document.created, document.sent, document.completed, document.declined, document.deleted, signer.added, signer.email_opened, signer.viewed, signer.signed, signer.declined, signer.reminded.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/webhooks",
			Summary:      "Create a webhook",
			Auth:         "bearer",
			RequestBody:  "CreateWebhookRequest",
			ResponseBody: "WebhookResponse",
		},
		{
			Method:       "GET",
			Path:         "/webhooks",
			Summary:      "List webhooks",
			Auth:         "bearer",
			ResponseBody: "[]WebhookResponse",
		},
		{
			Method:       "GET",
			Path:         "/webhooks/{id}",
			Summary:      "Get a webhook",
			Auth:         "bearer",
			ResponseBody: "WebhookResponse",
			PathParams:   []documentation.Field{{Name: "id", Type: "int", Description: "Webhook ID"}},
		},
		{
			Method:       "PUT",
			Path:         "/webhooks/{id}",
			Summary:      "Update a webhook",
			Auth:         "bearer",
			RequestBody:  "UpdateWebhookRequest",
			ResponseBody: "WebhookResponse",
			PathParams:   []documentation.Field{{Name: "id", Type: "int", Description: "Webhook ID"}},
		},
		{
			Method:      "DELETE",
			Path:        "/webhooks/{id}",
			Summary:     "Delete a webhook",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "id", Type: "int", Description: "Webhook ID"}},
		},
		{
			Method:      "POST",
			Path:        "/webhooks/{id}/test",
			Summary:     "Send a test event",
			Description: "Delivers a sample payload (event_type=test.ping) so the receiver can verify signature handling.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "id", Type: "int", Description: "Webhook ID"}},
		},
	},
}
