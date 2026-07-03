package clients

type CreateClientRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	Phone   string `json:"phone"`
	Notes   string `json:"notes"`
}

type UpdateClientRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	Phone   string `json:"phone"`
	Notes   string `json:"notes"`
}

type ClientResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Company   string `json:"company"`
	Phone     string `json:"phone"`
	Notes     string `json:"notes"`
	OwnerID   int64  `json:"owner_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
