package auth

import (
	"context"
	"io"
	"log/slog"
	"slices"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/testdb"
)

const testIssuer = "https://sso.test"

func claims(subject, email string, verified bool) porte.Claims {
	return porte.Claims{Provider: testIssuer, Subject: subject, Email: email, EmailVerified: verified, Locale: "de"}
}

func resetUsers(t *testing.T) *kit.Store {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testDB.Exec(`TRUNCATE users RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear users: %v", err)
	}
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	store, err := NewUsers(context.Background(), sqlDB, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewUsers: %v", err)
	}
	return store
}

func link(t *testing.T, userID int64, subject string) {
	t.Helper()
	err := testDB.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)`, userID, testIssuer, subject).Error
	if err != nil {
		t.Fatalf("link identity: %v", err)
	}
}

func TestFirstFacileSignInLinksAnExistingAccountByVerifiedEmail(t *testing.T) {
	store := resetUsers(t)
	ctx := context.Background()

	existing := schemas.User{Email: "camille@facile.studio", Name: "Camille"}
	if err := testDB.Create(&existing).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}

	id, err := store.UpsertFromOIDC(ctx, claims("sub-1", "Camille@Facile.Studio", true))
	if err != nil || id != existing.ID {
		t.Fatalf("first sign-in = %d, %v; want the existing %d", id, err, existing.ID)
	}
	link(t, id, "sub-1")

	again, err := store.UpsertFromOIDC(ctx, claims("sub-1", "camille@facile.studio", true))
	if err != nil || again != existing.ID {
		t.Fatalf("second sign-in = %d, %v; want %d", again, err, existing.ID)
	}

	var user schemas.User
	if err := testDB.First(&user, id).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if user.Locale != "de" {
		t.Errorf("locale = %q, want the sign-in's de", user.Locale)
	}
}

func TestANewSubjectCreatesAnAccountWithItsLocale(t *testing.T) {
	store := resetUsers(t)

	id, err := store.UpsertFromOIDC(context.Background(), claims("sub-new", "new@facile.studio", true))
	if err != nil {
		t.Fatalf("sign-in: %v", err)
	}
	var user schemas.User
	if err := testDB.First(&user, id).Error; err != nil || user.Locale != "de" {
		t.Fatalf("created user = %+v, %v; want locale de", user, err)
	}
}

func TestAnUnverifiedAddressNeverLinksToAnExistingAccount(t *testing.T) {
	store := resetUsers(t)

	if err := testDB.Create(&schemas.User{Email: "noah@facile.studio", Name: "Noah"}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := store.UpsertFromOIDC(context.Background(), claims("sub-2", "noah@facile.studio", false)); err == nil {
		t.Fatal("an unverified address linked to an existing account")
	}
}

func TestARecycledAddressIsNotHandedToANewSubject(t *testing.T) {
	store := resetUsers(t)

	owner := schemas.User{Email: "lea@facile.studio", Name: "Lea"}
	if err := testDB.Create(&owner).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	link(t, owner.ID, "sub-owner")

	if _, err := store.UpsertFromOIDC(context.Background(), claims("sub-newcomer", "lea@facile.studio", true)); err == nil {
		t.Fatal("a second subject of the same issuer took over a linked account")
	}
}

func TestTheKitLeavesOnlyTheseColumnsToFillNewAccount(t *testing.T) {
	store := resetUsers(t)
	want := []string{"avatar_url", "avatar_source", "avatar_upload_path", "oidc_subject", "password_hash",
		"oidc_access_token", "oidc_refresh_token", "oidc_token_expiry", "profile_synced_at"}
	if got := store.Unwritten(); !slices.Equal(got, want) {
		t.Fatalf("Unwritten() = %v, want %v", got, want)
	}
}

func nullColumns(t *testing.T, userID int64) []string {
	t.Helper()
	var columns []string
	err := testDB.Raw(`SELECT key FROM users u, jsonb_each(to_jsonb(u)) WHERE u.id = ? AND value = 'null'::jsonb ORDER BY key`, userID).
		Scan(&columns).Error
	if err != nil {
		t.Fatalf("read null columns: %v", err)
	}
	return columns
}

func TestAKitSignUpLeavesNullOnlyWhereAGormAccountDoes(t *testing.T) {
	store := resetUsers(t)
	old := schemas.User{Email: "old@facile.studio", Name: "Old"}
	if err := testDB.Create(&old).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	id, err := store.UpsertFromOIDC(context.Background(), claims("sub-kit", "kit@facile.studio", true))
	if err != nil {
		t.Fatalf("sign-in: %v", err)
	}
	if got, want := nullColumns(t, id), nullColumns(t, old.ID); !slices.Equal(got, want) {
		t.Fatalf("NULL columns after a kit sign-up = %v, a GORM account has %v", got, want)
	}
	var user schemas.User
	if err := testDB.First(&user, id).Error; err != nil || user.CreatedAt.IsZero() || user.ReminderIntervalDays != 3 {
		t.Fatalf("new account = %+v, %v", user, err)
	}
}

func TestTheStoredEmailChangesOnlyWhenTheProviderVerifiedIt(t *testing.T) {
	store := resetUsers(t)
	ctx := context.Background()
	id, err := store.UpsertFromOIDC(ctx, claims("sub-mail", "first@facile.studio", true))
	if err != nil {
		t.Fatalf("sign-in: %v", err)
	}
	link(t, id, "sub-mail")
	for _, c := range []struct {
		email    string
		verified bool
		want     string
	}{{"unverified@facile.studio", false, "first@facile.studio"}, {"second@facile.studio", true, "second@facile.studio"}} {
		if _, err := store.UpsertFromOIDC(ctx, claims("sub-mail", c.email, c.verified)); err != nil {
			t.Fatalf("sign-in as %s: %v", c.email, err)
		}
		var user schemas.User
		if err := testDB.First(&user, id).Error; err != nil || user.Email != c.want {
			t.Fatalf("email after signing in as %s (verified %v) = %q, want %q", c.email, c.verified, user.Email, c.want)
		}
	}
}
