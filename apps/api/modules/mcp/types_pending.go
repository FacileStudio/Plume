package mcp

import mcpkit "github.com/FacileStudio/tronc/mcp"

// PendingItem is one action waiting on the confirm page.
type PendingItem struct {
	mcpkit.PendingItem
	Document *PendingDocument `json:"document,omitempty"`
	Signer   *PendingSigner   `json:"signer,omitempty"`
	Client   *PendingClient   `json:"client,omitempty"`
	Field    *PendingField    `json:"field,omitempty"`
}

// PendingClient is the client a deletion targets, as it is now.
type PendingClient struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
}

// PendingField is the draft field a deletion targets, as it is now.
type PendingField struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Page  int    `json:"page"`
	Label string `json:"label"`
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
