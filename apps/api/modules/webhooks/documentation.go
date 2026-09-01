package webhooks

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "webhooks",
	Description: "Webhook management and delivery configuration routes.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/webhooks",
			Summary:      "Create a webhook",
			Description:  "Registers a new webhook endpoint to receive document event notifications. An HMAC signing secret is generated.",
			Auth:         "bearer",
			RequestBody:  CreateWebhookRequest{},
			ResponseBody: WebhookResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid URL or invalid event types."},
			},
		},
		{
			Method:       "GET",
			Path:         "/webhooks",
			Summary:      "List webhooks",
			Description:  "Returns all webhook endpoints configured by the calling user.",
			Auth:         "bearer",
			ResponseBody: []WebhookResponse{},
		},
		{
			Method:       "GET",
			Path:         "/webhooks/{id}",
			Summary:      "Get a webhook",
			Description:  "Returns a single webhook configuration by ID.",
			Auth:         "bearer",
			ResponseBody: WebhookResponse{},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Webhook not found."},
			},
		},
		{
			Method:       "PUT",
			Path:         "/webhooks/{id}",
			Summary:      "Update a webhook",
			Description:  "Updates a webhook endpoint URL, event subscriptions or active status.",
			Auth:         "bearer",
			RequestBody:  UpdateWebhookRequest{},
			ResponseBody: WebhookResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid URL or invalid event types."},
				{Status: 404, Code: "not_found", Description: "Webhook not found."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/webhooks/{id}",
			Summary:     "Delete a webhook",
			Description: "Removes a webhook endpoint.",
			Auth:        "bearer",
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Webhook not found."},
			},
		},
		{
			Method:      "POST",
			Path:        "/webhooks/{id}/test",
			Summary:     "Send a test webhook event",
			Description: "Dispatches a ping event to the webhook endpoint to verify connectivity and signature verification.",
			Auth:        "bearer",
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Webhook not found."},
				{Status: 502, Code: "upstream_error", Description: "The webhook target endpoint returned an error or timed out."},
			},
		},
	},
}
