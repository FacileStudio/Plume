package mcp

import (
	documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// Documentation describes the confirm page routes for the API reference. The MCP endpoint
// itself is outside /api: POST /mcp speaks the Model Context Protocol to any MCP client
// holding a Registre access token for this server, and
// /.well-known/oauth-protected-resource/mcp describes how to get one.
func Documentation() documentation.Module {
	return documentation.Module{
		Name: "mcp",
		Description: "Sends, reminders and deletions (of a document, a client or a draft's field) MCP clients asked for, waiting for the person to confirm them. " +
			"A client that cannot ask the person itself hands them a link to the pending page, which uses these " +
			"routes. The MCP endpoint is POST /mcp, outside /api.",
		Routes: mcpkit.PendingRoutes[PendingItem](mcpkit.PendingDoc{
			List: "Open actions, newest first, with their document, signer, client or field as it is now. state is " +
				"ready, changed (the target changed after the request, so it cannot be confirmed) or missing.",
			Confirm: "Runs the action once: sends the draft, emails the reminder, or deletes the document, client " +
				"or field. A target " +
				"changed since the request is refused with 412 and the action closes.",
			Dismiss: "Closes the action without running it.",
			Changed: "The document, signer, client or field changed after the request.",
		}),
	}
}
