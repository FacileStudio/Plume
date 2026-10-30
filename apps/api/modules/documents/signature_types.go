package documents

import "time"

// SignatureRequest is the JSON part of POST /documents/signature-requests.
// Send defaults to true: the signers are invited as soon as the document is
// stored. When Fields is empty, each signer and approver gets one required
// signature box on the last page.
type SignatureRequest struct {
	Name       string            `json:"name"`
	SpaceID    *int64            `json:"space_id,omitempty"`
	Sequential bool              `json:"sequential"`
	Send       *bool             `json:"send,omitempty"`
	Signers    []SignatureSigner `json:"signers"`
	Fields     []SignatureField  `json:"fields,omitempty"`
	Link
}

// SignatureSigner is one participant of a signature request.
type SignatureSigner struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role,omitempty"`
	Order int    `json:"order,omitempty"`
}

// SignatureField places a field for the signer at index Signer of the
// request. Coordinates are percentages of the page, pages start at 1.
type SignatureField struct {
	Signer    int     `json:"signer"`
	FieldType string  `json:"field_type"`
	Page      int     `json:"page"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	Required  bool    `json:"required"`
	Label     string  `json:"label,omitempty"`
}

// SignatureResponse is a document and its signers, as returned when a
// signature request is created and by the lookup by external_ref.
type SignatureResponse struct {
	Document DocumentResponse        `json:"document"`
	Signers  []SignatureSignerStatus `json:"signers"`
}

// SignatureSignerStatus is a signer with the link to hand them.
type SignatureSignerStatus struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Status     string     `json:"status"`
	OrderNum   int        `json:"order_num"`
	SigningURL string     `json:"signing_url,omitempty"`
	SignedAt   *time.Time `json:"signed_at,omitempty"`
}
