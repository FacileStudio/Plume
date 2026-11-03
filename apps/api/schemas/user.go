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

// AvatarSelectExpr is Avatar() as SQL, for any join that reads a user's picture without
// loading the row. It has to stay in step with Avatar below, hence both being here.
const AvatarSelectExpr = `COALESCE(users.oidc_picture_url, '')`

// Avatar is the picture to render: the stable Registre URL from the picture claim, or ""
// for a user who never signed in through SSO. Registre answers 404 for a user without a
// photo, and the client falls back to initials. avatar_upload_path is kept in the table
// but no longer read: Plume has no upload form, and the profile picture belongs to Registre.
func (u User) Avatar() string {
	return u.OIDCPictureURL
}

// AvatarOrigin names where Avatar came from, so the client can say *why* a picture is not
// editable here instead of leaving the user hunting for a control that will never appear.
func (u User) AvatarOrigin() string {
	if u.OIDCPictureURL != "" {
		return "oidc"
	}
	return ""
}
