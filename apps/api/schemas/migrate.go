package schemas

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	db.Exec("DROP INDEX IF EXISTS idx_signers_token")

	if err := db.AutoMigrate(&User{}, &Session{}, &Space{}, &SpaceMember{}, &Document{}, &Signer{}, &Field{}, &Webhook{}, &SmtpConfig{}, &Client{}); err != nil {
		return err
	}

	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_space_members_space_user ON space_members (space_id, user_id)")

	return backfillAvatarColumns(db)
}

// backfillAvatarColumns moves the avatar state onto the two columns that now own it.
//
// The filename decides which files to keep, not avatar_source. That column was added
// after the upload feature, so the oldest uploaded avatars have it empty, and keying on
// avatar_source = 'upload' quietly drops their picture — this cost Sablier two of its four
// rows in rehearsal. Uploads have always been named "user-<id>-<nanos>" and the old OIDC
// downloads "oidc-<id>-<nanos>", so anything that is not an oidc- copy is somebody's file
// and is kept.
//
// The second statement is the one specific to having stored profile.Picture verbatim:
// Authentik never omits the claim, so every user without a photo has a
// "data:image/svg+xml;base64,…" rendering of their own initials sitting in
// oidc_picture_url. Under the new rule that column means "there is an SSO photo", so a
// stale data: URI would report a photo that is really a placeholder and suppress the
// upload fallback forever.
//
// The oidc- copies on the volume are left where they are: a migration that deletes files
// has to be right the first time, and they are a few hundred kilobytes a later sweep can
// take once this has proven itself. avatar_url and avatar_source likewise stay in the
// table, unread, until the next release drops them — expanding first means a rollback is
// redeploying the old binary, not restoring a backup.
func backfillAvatarColumns(db *gorm.DB) error {
	if !db.Migrator().HasColumn(&User{}, "avatar_url") {
		return nil
	}
	if err := db.Exec(
		`UPDATE users SET avatar_upload_path = replace(avatar_url, '/files/', '')
		 WHERE coalesce(avatar_url, '') <> ''
		   AND avatar_url NOT LIKE '/files/avatars/oidc-%'
		   AND coalesce(avatar_upload_path, '') = ''`).Error; err != nil {
		return err
	}
	if err := db.Exec(
		`UPDATE users SET oidc_picture_url = ''
		 WHERE coalesce(oidc_picture_url, '') <> ''
		   AND oidc_picture_url NOT LIKE 'https://%'`).Error; err != nil {
		return err
	}
	// A NULL here would fail to scan into the plain string the model declares.
	return db.Exec(`UPDATE users SET avatar_upload_path = '' WHERE avatar_upload_path IS NULL`).Error
}
