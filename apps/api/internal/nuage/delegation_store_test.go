package nuage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/tronc/testdb"
)

func TestDelegationRotatesTokensInPortesIdentityTable(t *testing.T) {
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_nuage_test", Migrate: schemas.Migrate})
	if err != nil {
		t.Fatalf("test database: %v", err)
	}
	if err := db.Exec(`TRUNCATE users RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear users: %v", err)
	}
	user := schemas.User{Email: "owner@facile.studio", Name: "Owner"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	store := portepg.New(sqlDB).Identities()

	reg := &fakeRegistre{}
	_, cfg := reg.serve(t)
	err = store.Save(context.Background(), porte.StoredIdentity{
		UserID: user.ID, Provider: cfg.Issuer, Subject: "sub-owner",
		Tokens: porte.TokenSet{AccessToken: "old", RefreshToken: "r0", Expiry: time.Now().Add(-time.Minute)},
	})
	if err != nil {
		t.Fatalf("seed identity: %v", err)
	}
	delegation := NewDelegation(cfg, store, quietLogger())

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if token, err := delegation.Token(context.Background(), user.ID); err != nil || token != "delegated" {
				t.Errorf("token=%q err=%v", token, err)
			}
		}()
	}
	wg.Wait()

	stored, err := store.Find(context.Background(), cfg.Issuer, "sub-owner")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if stored.Tokens.RefreshToken != "rotated" || stored.Tokens.AccessToken != "fresh" || stored.UserID != user.ID {
		t.Fatalf("stored identity = %+v, want the rotated tokens on the same account", stored)
	}
	refreshes := 0
	for _, form := range reg.forms {
		if form.Get("grant_type") == "refresh_token" {
			refreshes++
		}
	}
	if refreshes != 1 {
		t.Fatalf("refresh grants = %d, want 1 under the per-user lock", refreshes)
	}
}
