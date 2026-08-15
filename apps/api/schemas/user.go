package schemas

import "time"

// User is an authenticated account, holding local password credentials and the
// OIDC, avatar and reminder state porte and the UI read.
type User struct {
	ID                   int64     `gorm:"column:id;primaryKey"`
	Email                string    `gorm:"column:email;uniqueIndex"`
	Name                 string    `gorm:"column:name"`
	AvatarURL            string    `gorm:"column:avatar_url"`
	AvatarSource         string    `gorm:"column:avatar_source"`
	AvatarUploadPath     string    `gorm:"column:avatar_upload_path"`
	OIDCPictureURL       string    `gorm:"column:oidc_picture_url"`
	OIDCSubject          *string   `gorm:"column:oidc_subject;uniqueIndex"`
	PasswordHash         string    `gorm:"column:password_hash"`
	ReminderIntervalDays int       `gorm:"column:reminder_interval_days;default:3"`
	OIDCAccessToken      string    `gorm:"column:oidc_access_token"`
	OIDCRefreshToken     string    `gorm:"column:oidc_refresh_token"`
	OIDCTokenExpiry      time.Time `gorm:"column:oidc_token_expiry"`
	ProfileSyncedAt      time.Time `gorm:"column:profile_synced_at"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (User) TableName() string { return "users" }

// AvatarFilePrefix is where main.go mounts the upload volume. It is part of the derived
// value rather than something the client prepends: the other branch of Avatar is an
// absolute Porte URL, and a client that concatenates turns that into "/apihttps://…".
// One value, usable as an src attribute exactly as it arrives.
const AvatarFilePrefix = "/api/files/"

// AvatarSelectExpr is Avatar() as SQL, for any join that reads a user's picture without
// loading the row. It has to stay in step with Avatar below — hence both being here, one
// above the other, rather than one in Go and one buried in a Select string.
const AvatarSelectExpr = `COALESCE(NULLIF(users.oidc_picture_url, ''), ` +
	`NULLIF('` + AvatarFilePrefix + `' || COALESCE(users.avatar_upload_path, ''), '` + AvatarFilePrefix + `'), '')`

// Avatar is the picture to render. It is derived from the two sources rather than stored
// alongside them: a photo set in Porte always wins, an upload shows only when the IdP
// offers none, and because nothing is written there is no third value that can drift out
// of agreement with the two that matter.
func (u User) Avatar() string {
	if u.OIDCPictureURL != "" {
		return u.OIDCPictureURL
	}
	if u.AvatarUploadPath != "" {
		return AvatarFilePrefix + u.AvatarUploadPath
	}
	return ""
}

// AvatarOrigin names where Avatar came from, so the client can say *why* a picture is not
// editable here instead of leaving the user hunting for a control that will never appear.
func (u User) AvatarOrigin() string {
	switch {
	case u.OIDCPictureURL != "":
		return "oidc"
	case u.AvatarUploadPath != "":
		return "upload"
	default:
		return ""
	}
}
