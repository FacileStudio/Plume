package schemas

import "gorm.io/gorm"

func Migrate(db *gorm.DB) error {
	db.Exec("DROP INDEX IF EXISTS idx_signers_token")

	return db.AutoMigrate(&User{}, &Session{}, &Document{}, &Signer{}, &Field{}, &Webhook{}, &SmtpConfig{})
}
