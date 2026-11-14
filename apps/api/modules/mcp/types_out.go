package mcp

import "time"

// documentOut is one document as a model reads it.
type documentOut struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	FileName   string    `json:"file_name"`
	Sequential bool      `json:"sequential"`
	ClientID   *int64    `json:"client_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	URL        string    `json:"url"`
}

// countsOut counts the documents of one scope by state.
type countsOut struct {
	Total     int64 `json:"total"`
	Pending   int64 `json:"pending"`
	Completed int64 `json:"completed"`
}

// spaceOut is a space the person belongs to.
type spaceOut struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// listOutput is plume_list_documents's result.
type listOutput struct {
	Documents  []documentOut `json:"documents"`
	Counts     countsOut     `json:"counts"`
	Spaces     []spaceOut    `json:"spaces"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

// signerOut is one signer, without the private signing link.
type signerOut struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	Order          int        `json:"order"`
	Fields         int        `json:"fields"`
	ViewedAt       *time.Time `json:"viewed_at,omitempty"`
	SignedAt       *time.Time `json:"signed_at,omitempty"`
	LastRemindedAt *time.Time `json:"last_reminded_at,omitempty"`
}

// fieldOut is one field placed on the document, for the signer who fills it.
type fieldOut struct {
	ID       int64  `json:"id"`
	SignerID int64  `json:"signer_id"`
	Type     string `json:"type"`
	Page     int    `json:"page"`
	Required bool   `json:"required"`
	Label    string `json:"label,omitempty"`
}

// documentOutput is a document with its signers and fields: plume_get_document's result
// and what the tools that change a draft return.
type documentOutput struct {
	Status   string      `json:"status,omitempty"`
	Document documentOut `json:"document"`
	Client   *clientOut  `json:"client,omitempty"`
	Signers  []signerOut `json:"signers"`
	Fields   []fieldOut  `json:"fields,omitempty"`
}

// signerOutput is plume_add_signer's and plume_remove_signer's result.
type signerOutput struct {
	Status     string     `json:"status"`
	DocumentID int64      `json:"document_id"`
	Signer     *signerOut `json:"signer,omitempty"`
}

// actionOutput is the result of a confirmed tool: send, remind or a delete.
type actionOutput struct {
	Status     string       `json:"status"`
	DocumentID int64        `json:"document_id,omitempty"`
	SignerID   int64        `json:"signer_id,omitempty"`
	ClientID   int64        `json:"client_id,omitempty"`
	FieldID    int64        `json:"field_id,omitempty"`
	Document   *documentOut `json:"document,omitempty"`
	ConfirmURL string       `json:"confirm_url,omitempty"`
	Message    string       `json:"message,omitempty"`
}

// clientOut is one client.
type clientOut struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email,omitempty"`
	Company string `json:"company,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Notes   string `json:"notes,omitempty"`
}

// clientsOutput is plume_list_clients's result.
type clientsOutput struct {
	Clients    []clientOut `json:"clients"`
	NextCursor string      `json:"next_cursor,omitempty"`
}

// clientOutput is plume_save_client's result.
type clientOutput struct {
	Status string    `json:"status"`
	Client clientOut `json:"client"`
}
