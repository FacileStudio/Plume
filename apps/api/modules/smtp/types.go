package smtp

type SaveConfigRequest struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
}

type ConfigResponse struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
	UpdatedAt string `json:"updated_at"`
}

type TestRequest struct {
	To string `json:"to"`
}
