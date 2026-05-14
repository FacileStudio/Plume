package auth

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

type UpdateProfileRequest struct {
	Name                 string `json:"name"`
	Email                string `json:"email"`
	ReminderIntervalDays *int   `json:"reminder_interval_days,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type ProfileResponse struct {
	ID                   string `json:"id"`
	Email                string `json:"email"`
	Name                 string `json:"name"`
	ReminderIntervalDays int    `json:"reminder_interval_days"`
	CreatedAt            string `json:"created_at"`
}

type Data struct {
	Email string `json:"email"`
}

func (d *Data) GetEmail() string {
	if d == nil {
		return ""
	}
	return d.Email
}
