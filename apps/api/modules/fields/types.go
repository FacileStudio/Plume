package fields

// CreateFieldRequest is the body of POST /documents/{docId}/fields.
type CreateFieldRequest struct {
	SignerID  int64   `json:"signer_id"`
	FieldType string  `json:"field_type"`
	Page      int     `json:"page"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	Required  bool    `json:"required"`
	Label     string  `json:"label"`
}

// UpdateFieldRequest is the body of PUT /documents/{docId}/fields/{fieldId}.
type UpdateFieldRequest struct {
	FieldType string  `json:"field_type"`
	Page      int     `json:"page"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	Required  bool    `json:"required"`
	Label     string  `json:"label"`
}

// FieldResponse describes a field as returned by the API.
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
