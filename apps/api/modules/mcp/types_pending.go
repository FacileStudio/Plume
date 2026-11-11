package mcp

import mcpkit "github.com/FacileStudio/tronc/mcp"

// PendingItem is one action waiting on the confirm page.
type PendingItem struct {
	mcpkit.PendingItem
	Document *PendingDocument `json:"document,omitempty"`
	Signer   *PendingSigner   `json:"signer,omitempty"`
}

// PendingDocument is the document an action targets, as it is now.
type PendingDocument struct {
	ID      int64    `json:"id"`
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Signers []string `json:"signers"`
}

// PendingSigner is the signer a reminder targets, as they are now.
type PendingSigner struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
