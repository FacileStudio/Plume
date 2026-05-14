package documents

import "time"

type UpdateRequest struct {
	Name       string `json:"name"`
	FileName   string `json:"file_name"`
	Sequential *bool  `json:"sequential"`
}

type StatsResponse struct {
	Total     int64 `json:"total"`
	Pending   int64 `json:"pending"`
	Completed int64 `json:"completed"`
}

type DocumentResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	FileName     string    `json:"file_name"`
	OwnerID      int64     `json:"owner_id"`
	Sequential   bool      `json:"sequential"`
	OriginalHash string    `json:"original_hash,omitempty"`
	SignedHash   string    `json:"signed_hash,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
