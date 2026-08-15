package auth

// RegisterRequest is the body of POST /auth/register.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the body of POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is returned after a successful registration or login.
type AuthResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

// UpdateProfileRequest is the body of PUT /auth/me.
type UpdateProfileRequest struct {
	Name                 string `json:"name"`
	Email                string `json:"email"`
	ReminderIntervalDays *int   `json:"reminder_interval_days,omitempty"`
}

// ChangePasswordRequest is the body of PUT /auth/password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ProfileResponse describes a user's profile as returned by /auth/me.
type ProfileResponse struct {
	ID                   string `json:"id"`
	Email                string `json:"email"`
	Name                 string `json:"name"`
	AvatarURL            string `json:"avatar_url,omitempty"`
	AvatarSource         string `json:"avatar_source,omitempty"`
	ReminderIntervalDays int    `json:"reminder_interval_days"`
	CreatedAt            string `json:"created_at"`
}

// Data carries an email address for API documentation purposes.
type Data struct {
	Email string `json:"email"`
}

func (d *Data) GetEmail() string {
	if d == nil {
		return ""
	}
	return d.Email
}
