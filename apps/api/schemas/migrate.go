package schemas

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	db.Exec("DROP INDEX IF EXISTS idx_signers_token")

	if err := db.AutoMigrate(&User{}, &Session{}, &Space{}, &SpaceMember{}, &Document{}, &Signer{}, &Field{}, &Webhook{}, &SmtpConfig{}); err != nil {
		return err
	}

	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_space_members_space_user ON space_members (space_id, user_id)")

	return nil
}
