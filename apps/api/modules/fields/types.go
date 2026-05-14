package fields

type CreateFieldRequest struct {
	SignerID  int64   `json:"signer_id"`
	FieldType string  `json:"field_type"`
	Page      int     `json:"page"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	Required  bool    `json:"required"`
}

type UpdateFieldRequest struct {
	FieldType string  `json:"field_type"`
	Page      int     `json:"page"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	Required  bool    `json:"required"`
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
	Value      string  `json:"value"`
}
