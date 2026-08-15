package verify

import "time"

// Response is the result of a verification lookup: whether the hash matched
// and, if so, the document and signer details.
type Response struct {
	Match    bool         `json:"match"`
	Hash     string       `json:"hash"`
	Variant  string       `json:"variant,omitempty"`
	Document *DocumentDTO `json:"document,omitempty"`
	Signers  []SignerDTO  `json:"signers,omitempty"`
}

// DocumentDTO is the verified document's metadata, with the completion time
// when it is completed.
type DocumentDTO struct {
	Name        string     `json:"name"`
	FileName    string     `json:"file_name"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// SignerDTO is a verifiable signer summary with the email masked.
type SignerDTO struct {
	Name     string     `json:"name"`
	Email    string     `json:"email"`
	Status   string     `json:"status"`
	SignedAt *time.Time `json:"signed_at,omitempty"`
}
