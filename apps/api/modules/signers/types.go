package signers

import "time"

// AddSignerRequest is the payload for inviting a signer onto a document.
type AddSignerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Order int    `json:"order"`
}

// SubmitSignatureRequest carries the field values a signer submits when
// signing.
type SubmitSignatureRequest struct {
	Fields []FieldValue `json:"fields"`
}

// FieldValue links a submitted value to the field it fills.
type FieldValue struct {
	FieldID int64  `json:"field_id"`
	Value   string `json:"value"`
}

// SignerResponse is the API representation of a signer on a document.
type SignerResponse struct {
	ID             int64      `json:"id"`
	DocumentID     int64      `json:"document_id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	Token          string     `json:"token,omitempty"`
	OrderNum       int        `json:"order_num"`
	SignedAt       *time.Time `json:"signed_at"`
	ViewedAt       *time.Time `json:"viewed_at"`
	EmailOpenedAt  *time.Time `json:"email_opened_at"`
	IPAddress      string     `json:"ip_address,omitempty"`
	UserAgent      string     `json:"user_agent,omitempty"`
	LastRemindedAt *time.Time `json:"last_reminded_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

// FieldResponse is a document field as returned to the client, including any
// stored value.
type FieldResponse struct {
	ID         int64   `json:"id"`
	DocumentID int64   `json:"document_id"`
	SignerID   int64   `json:"signer_id"`
	FieldType  string  `json:"field_type"`
	Page       int     `json:"page"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	Required   bool    `json:"required"`
	Label      string  `json:"label"`
	Value      string  `json:"value"`
}

// SigningView is everything a signer needs to render the signing page: the
// document, their own fields and the fields others have already filled in.
type SigningView struct {
	Document        DocumentInfo             `json:"document"`
	Signer          SignerResponse           `json:"signer"`
	Fields          []FieldResponse          `json:"fields"`
	CompletedFields []CompletedFieldResponse `json:"completed_fields"`
}

// CompletedFieldResponse is a field filled in by another signer, shown
// read-only.
type CompletedFieldResponse struct {
	ID         int64   `json:"id"`
	SignerName string  `json:"signer_name"`
	FieldType  string  `json:"field_type"`
	Label      string  `json:"label"`
	Page       int     `json:"page"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	Value      string  `json:"value"`
}

// DocumentInfo is the document summary embedded in the signing views.
type DocumentInfo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FileName string `json:"file_name"`
	Status   string `json:"status"`
}

// SigningStatusResponse is a token-scoped view of a document's overall signing
// progress, safe to return in any non-draft state (including completed). It
// powers the post-signing audit/progress screen shown to a signer.
type SigningStatusResponse struct {
	Document DocumentInfo         `json:"document"`
	Signer   SigningRosterEntry   `json:"signer"`
	Signers  []SigningRosterEntry `json:"signers"`
}

// SigningRosterEntry is a single participant in the signing workflow, exposing
// only the fields other signers are allowed to see (no email, IP or user agent).
type SigningRosterEntry struct {
	Name     string     `json:"name"`
	Role     string     `json:"role"`
	Status   string     `json:"status"`
	OrderNum int        `json:"order_num"`
	SignedAt *time.Time `json:"signed_at"`
	IsYou    bool       `json:"is_you"`
}
