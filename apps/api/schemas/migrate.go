package schemas

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Plume/apps/api/migrations"
	"gorm.io/gorm"
)

// Migrate applies the migrations and then the boot data steps, with no OIDC
// issuer and no migration log. The tests build their schemas through it, so
// they run exactly the SQL production runs.
func Migrate(db *gorm.DB) error {
	return MigrateWithIssuer(context.Background(), db, "", slog.New(slog.DiscardHandler))
}

// MigrateWithIssuer applies the ordered migrations under migrations/ and then
// Reconcile. The issuer is what the identity backfill needs: porte matches an
// account on (provider, subject) and the provider is the issuer, so
// backfilling with a placeholder would leave every existing SSO user
// unmatched and quietly fall through to the email path.
func MigrateWithIssuer(ctx context.Context, db *gorm.DB, issuer string, logger *slog.Logger) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if err := migrations.Up(ctx, migrations.Config(sqlDB, logger)); err != nil {
		return err
	}
	return Reconcile(db, issuer)
}

// Reconcile runs the idempotent data steps that ran at every boot before the
// schema moved to migrations, in the same order. It changes rows, never the
// schema: a facile_id for every space, the avatar columns, and the
// credentials porte reads. AdoptPorte runs last because it reads the users
// rows the avatar backfill has just settled.
func Reconcile(db *gorm.DB, issuer string) error {
	if err := db.Exec(backfillFacileIDs).Error; err != nil {
		return err
	}
	if err := backfillAvatarColumns(db); err != nil {
		return err
	}
	return AdoptPorte(db, issuer)
}

// backfillFacileIDs gives every space that predates facile_id, or was written
// without Space.BeforeCreate, its suite-wide id. A NULL never collides in the
// unique idx_spaces_facile_id, so the rows it reaches can always be updated.
const backfillFacileIDs = `UPDATE spaces
	SET facile_id = 'fac_' || substr(md5(random()::text || id::text || clock_timestamp()::text), 1, 24)
	WHERE facile_id IS NULL OR facile_id = ''`

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
//
// The final statement clears any remaining NULL upload paths: the column is a plain
// string and a NULL here would fail to scan, so any row still NULL is simply set to an
// empty string.
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
	return db.Exec(`UPDATE users SET avatar_upload_path = '' WHERE avatar_upload_path IS NULL`).Error
}
