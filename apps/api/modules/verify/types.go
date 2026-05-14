package verify

import "time"

type Response struct {
	Match    bool         `json:"match"`
	Hash     string       `json:"hash"`
	Variant  string       `json:"variant,omitempty"`
	Document *DocumentDTO `json:"document,omitempty"`
	Signers  []SignerDTO  `json:"signers,omitempty"`
}

type DocumentDTO struct {
	Name        string     `json:"name"`
	FileName    string     `json:"file_name"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type SignerDTO struct {
	Name     string     `json:"name"`
	Email    string     `json:"email"`
	Status   string     `json:"status"`
	SignedAt *time.Time `json:"signed_at,omitempty"`
}
