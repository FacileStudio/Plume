package webhooks

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

// Documentation returns the webhooks module's route table for the API reference.
func Documentation() documentation.Module {
	routes := createWebhookRoute()
	routes = append(routes, listWebhooksRoute()...)
	routes = append(routes, getWebhookRoute()...)
	routes = append(routes, updateWebhookRoute()...)
	routes = append(routes, deleteWebhookRoute()...)
	routes = append(routes, testWebhookRoute()...)
	return documentation.Module{
		Name:        "webhooks",
		Description: "Webhook management and delivery configuration routes.",
		Routes:      routes,
	}
}

// createWebhookRoute returns the POST /webhooks route entry.
func createWebhookRoute() []documentation.Route {
	return []documentation.Route{
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
	}
}

// listWebhooksRoute returns the GET /webhooks route entry.
func listWebhooksRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/webhooks",
			Summary:      "List webhooks",
			Description:  "Returns all webhook endpoints configured by the calling user.",
			Auth:         "bearer",
			ResponseBody: []WebhookResponse{},
		},
	}
}

// getWebhookRoute returns the GET /webhooks/{id} route entry.
func getWebhookRoute() []documentation.Route {
	return []documentation.Route{
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
	}
}

// updateWebhookRoute returns the PUT /webhooks/{id} route entry.
func updateWebhookRoute() []documentation.Route {
	return []documentation.Route{
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
	}
}

// deleteWebhookRoute returns the DELETE /webhooks/{id} route entry.
func deleteWebhookRoute() []documentation.Route {
	return []documentation.Route{
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
	}
}

// testWebhookRoute returns the POST /webhooks/{id}/test route entry.
func testWebhookRoute() []documentation.Route {
	return []documentation.Route{
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
	}
}
