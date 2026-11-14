package mcp

import (
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// list is the curated tool set. Reads are open world: teammates and the people who sign
// name the documents, signers and clients. Sending, reminding and deleting are the
// sensitive tools. docs/mcp.md lists what is left out on purpose.
func (t *toolset) list() []mcpkit.Tool {
	return append(append(append(t.reads(), t.writes()...), t.sensitive()...), t.deletions()...)
}

// reads are the three read-only tools.
func (t *toolset) reads() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolList, Title: "List documents",
			Description: "List the user's personal documents, or one space's with space_id, newest first, 25 a page: " +
				"id, name, status (draft, pending, completed, declined) and a link to open it in Plume. Also returns " +
				"the counts by status and the spaces the user belongs to. Pass status, client_id or query to narrow it " +
				"and next_cursor back as cursor for the next page.",
			Annotations: &mcpkit.Annotations{ReadOnly: true, Idempotent: true, OpenWorld: true},
		}, t.listDocuments),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolGet, Title: "Get a document",
			Description: "Show one document with its client, its signers (name, email, role, status, order, how " +
				"many fields each fills, and when they viewed, signed or were last reminded) and its fields (id, " +
				"signer, type, page). Signing links are private to each signer and never returned.",
			Annotations: &mcpkit.Annotations{ReadOnly: true, Idempotent: true, OpenWorld: true},
		}, t.getDocument),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolClients, Title: "List clients",
			Description: "List the user's clients, newest first, 50 a page: id, name, email, company, phone and notes. " +
				"Pass query to find one by name, company or email.",
			Annotations: &mcpkit.Annotations{ReadOnly: true, Idempotent: true, OpenWorld: true},
		}, t.listClients),
	}
}

// writes are the tools that create or change drafts and clients without emailing anyone.
func (t *toolset) writes() []mcpkit.Tool {
	return append(append(t.draftWrites(), t.signerWrites()...), t.clientWrites()...)
}

// draftWrites create a draft and change its settings.
func (t *toolset) draftWrites() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolCreate, Title: "Create a document",
			Description: "Upload a PDF (base64, at most 3 MB) as a draft with 1 to 20 signers. Each signer and " +
				"approver gets a required signature box on the last page; a viewer only receives the completed " +
				"document. Nothing is emailed: call plume_send_document when the draft is ready. Fine field " +
				"placement is done in Plume through the returned url.",
			Annotations: &mcpkit.Annotations{},
		}, t.createDocument),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolUpdate, Title: "Edit a draft",
			Description: "Change a draft's name, whether signers sign in order (sequential), or its client. Fields " +
				"left out stay as they are. A document already sent cannot change.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.updateDocument),
	}
}

// signerWrites add and remove the people on a draft.
func (t *toolset) signerWrites() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolAddSigner, Title: "Add a signer",
			Description: "Add a person to a draft as signer (default), approver or viewer, with a required signature " +
				"box on the last page unless they are a viewer or no_signature_box is set. Nothing is emailed until " +
				"plume_send_document.",
			Annotations: &mcpkit.Annotations{},
		}, t.addSigner),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolRemoveSigner, Title: "Remove a signer",
			Description: "Take a person off a draft the user owns, with the fields placed for them. A document " +
				"already sent keeps its signers.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.removeSigner),
	}
}

// clientWrites create and change clients.
func (t *toolset) clientWrites() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolSaveClient, Title: "Save a client",
			Description: "Create a client (leave out client_id; name is required) or change one: fields left out " +
				"stay as they are, an empty string clears one. Use the client's id as client_id on a document.",
			Annotations: &mcpkit.Annotations{Destructive: true},
		}, t.saveClient),
	}
}

// sensitive are the tools the user confirms.
func (t *toolset) sensitive() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolSend, Title: "Send for signature",
			Description: "Send a draft: Plume emails each signer and approver a private signing link (in order when " +
				"the document is sequential) and the document becomes pending. The user always confirms: in your " +
				"client when it can ask, otherwise the result has status pending_confirmation and a confirm_url to " +
				"give the user. Nothing is sent before they confirm.",
			Annotations: &mcpkit.Annotations{OpenWorld: true},
		}, t.sendDocument),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolRemind, Title: "Remind a signer",
			Description: "Email a signer who has not signed yet their signing link again, on a document the user owns " +
				"that is out for signature. Plume refuses a second reminder within a short cooldown. The user always " +
				"confirms, as with plume_send_document.",
			Annotations: &mcpkit.Annotations{OpenWorld: true},
		}, t.remindSigner),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolDelete, Title: "Delete a document",
			Description: "Delete a document, its signers, fields and file for good, whatever its status. The user " +
				"always confirms: in your client when it can ask, otherwise the result has status " +
				"pending_confirmation and a confirm_url to give the user. Nothing is deleted before they confirm.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.deleteDocument),
	}
}

// deletions are the confirmed tools that remove a client or a draft's field.
func (t *toolset) deletions() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolDeleteClient, Title: "Delete a client",
			Description: "Delete one of the user's clients for good. The documents filed under it stay, with no " +
				"client. The user always confirms: in your client when it can ask, otherwise the result has status " +
				"pending_confirmation and a confirm_url to give the user. Nothing is deleted before they confirm.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.deleteClient),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolDeleteField, Title: "Delete a field",
			Description: "Delete one field, such as a signature box, from a draft; field ids are in " +
				"plume_get_document. A document already sent keeps its fields. The user always confirms, as with " +
				"plume_delete_client.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.deleteField),
	}
}
