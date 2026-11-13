package auth

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/FacileStudio/porte/kit"
)

// fillNewAccountSQL writes the values GORM wrote into every account it
// created, which the kit's insert leaves NULL: empty strings for the text
// columns and Go's zero time for the two timestamps. oidc_subject stays NULL,
// as it does on every GORM-created row, because its unique index needs it.
// Since migration 00002 the column defaults write the same values; this stays
// so an account is still filled on a schema whose defaults a rollback to an
// AutoMigrate release took away.
const fillNewAccountSQL = `UPDATE users SET
	avatar_url = '', avatar_source = '', avatar_upload_path = '', password_hash = '',
	oidc_access_token = '', oidc_refresh_token = '',
	oidc_token_expiry = '0001-01-01 00:00:00+00', profile_synced_at = '0001-01-01 00:00:00+00'
	WHERE id = $1`

// Columns maps porte/kit onto Plume's users table: the picture lives in
// oidc_picture_url and there is no admin flag.
func Columns() kit.Columns {
	return kit.Columns{Picture: "oidc_picture_url"}
}

// NewUsers builds the kit store over Plume's users table. It is porte's
// UserStore at sign-in and the resolver behind RequireAuth. Erasure is not
// routed through it: Plume keeps the account row as a tombstone for the
// signed documents that name it, so the kit's EraseAccount, which deletes
// the row, is never wired.
func NewUsers(ctx context.Context, db *sql.DB, logger *slog.Logger) (*kit.Store, error) {
	return kit.NewStore(ctx, kit.Store{DB: db, Columns: Columns(), OnCreate: fillNewAccount, Logger: logger})
}

func fillNewAccount(ctx context.Context, tx *sql.Tx, userID int64) error {
	_, err := tx.ExecContext(ctx, fillNewAccountSQL, userID)
	return err
}
