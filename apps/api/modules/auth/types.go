package auth

// UpdateProfileRequest is the body of PUT /auth/me.
type UpdateProfileRequest struct {
	Name                 string `json:"name"`
	Email                string `json:"email"`
	ReminderIntervalDays *int   `json:"reminder_interval_days,omitempty"`
}

// ProfileResponse describes a user's profile as returned by /auth/me.
type ProfileResponse struct {
	ID                   string `json:"id"`
	Email                string `json:"email"`
	Name                 string `json:"name"`
	AvatarURL            string `json:"avatar_url,omitempty"`
	AvatarSource         string `json:"avatar_source,omitempty"`
	Locale               string `json:"locale"`
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
