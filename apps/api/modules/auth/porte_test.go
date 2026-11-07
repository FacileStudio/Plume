package auth

import (
	"context"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/testdb"
)

const testIssuer = "https://sso.test"

func claims(subject, email string, verified bool) porte.Claims {
	return porte.Claims{Provider: testIssuer, Subject: subject, Email: email, EmailVerified: verified, Locale: "de"}
}

func resetUsers(t *testing.T) *UserStore {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testDB.Exec(`TRUNCATE users RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear users: %v", err)
	}
	return NewUserStore(testDB)
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
