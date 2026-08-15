package smtp

// SaveConfigRequest is the payload for creating or updating the SMTP
// configuration.
type SaveConfigRequest struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
}

// ConfigResponse is the SMTP configuration as returned to the client, with the
// password omitted.
type ConfigResponse struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
	UpdatedAt string `json:"updated_at"`
}

// TestRequest is the payload for sending a test email to a target address.
type TestRequest struct {
	To string `json:"to"`
}
