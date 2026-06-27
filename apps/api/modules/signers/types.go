package signers

import "time"

type AddSignerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Order int    `json:"order"`
}

type SubmitSignatureRequest struct {
	Fields []FieldValue `json:"fields"`
}

type FieldValue struct {
	FieldID int64  `json:"field_id"`
	Value   string `json:"value"`
}

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

type SigningView struct {
	Document        DocumentInfo             `json:"document"`
	Signer          SignerResponse           `json:"signer"`
	Fields          []FieldResponse          `json:"fields"`
	CompletedFields []CompletedFieldResponse `json:"completed_fields"`
}

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

type DocumentInfo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FileName string `json:"file_name"`
	Status   string `json:"status"`
}
