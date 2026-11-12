package spaces

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/modules/auth"
	"github.com/FacileStudio/porte/directory"
	portepg "github.com/FacileStudio/porte/pg"
)

// erasedFixture is a receiver wired as main wires it, against a Registre that
// lists one erasure, sub-gone, on every poll.
type erasedFixture struct {
	issuer    string
	deletions *Deletions
	ledger    *directory.Deletions
	other     int64
}

func newErasedFixture(t *testing.T) erasedFixture {
	t.Helper()
	newService(t)
	if err := testDB.Exec(`DELETE FROM directory_deletion_state`).Error; err != nil {
		t.Fatalf("clear state: %v", err)
	}
	deletedAt := time.Now().UTC().Add(-time.Minute).Format(time.RFC3339Nano)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"deleted":[{"sub":"sub-gone","deleted_at":"` + deletedAt + `"}]}`))
	}))
	t.Cleanup(server.Close)
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	users, err := auth.NewUsers(context.Background(), sqlDB, logger)
	if err != nil {
		t.Fatalf("NewUsers: %v", err)
	}
	store := portepg.New(sqlDB)
	deletions := NewDeletions(DeletionsConfig{
		ORM: testDB, Client: directory.Client{Issuer: server.URL}, Users: users,
		Freeze: func(context.Context, int64) error { return nil },
		Erase:  store.EraseIdentity, Holds: store.HoldsIdentity, UploadDir: t.TempDir(), Logger: logger,
	})
	f := erasedFixture{issuer: server.URL, deletions: deletions, ledger: ErasedLedger(testDB, server.URL)}
	f.other = f.account(t, "other@facile.studio", "sub-other")
	return f
}

func (f erasedFixture) account(t *testing.T, email, subject string) int64 {
	t.Helper()
	userID := createUser(t, email)
	if err := testDB.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)`,
		userID, f.issuer, subject).Error; err != nil {
		t.Fatalf("link identity: %v", err)
	}
	return userID
}

func (f erasedFixture) poll(t *testing.T) {
	t.Helper()
	if err := f.deletions.receiver.Poll(context.Background()); err != nil {
		t.Fatalf("poll: %v", err)
	}
}

func (f erasedFixture) held(t *testing.T, subject string) int64 {
	t.Helper()
	return count(t, `SELECT count(*) FROM porte_identities WHERE provider = ? AND subject = ?`, f.issuer, subject)
}

// An account re-created for a subject Registre already erased is erased again
// by the next poll of the same receiver, with no restart, and nobody else is
// touched.
func TestARecreatedAccountForAnErasedSubjectIsErasedAgain(t *testing.T) {
	f := newErasedFixture(t)
	gone := f.account(t, "gone@facile.studio", "sub-gone")
	f.poll(t)
	if f.held(t, "sub-gone") != 0 || count(t, `SELECT count(*) FROM users WHERE id = ?`, gone) != 0 {
		t.Fatalf("the first poll did not erase sub-gone")
	}

	again := f.account(t, "gone@facile.studio", "sub-gone")
	f.poll(t)
	if f.held(t, "sub-gone") != 0 || count(t, `SELECT count(*) FROM users WHERE id = ?`, again) != 0 {
		t.Fatalf("the re-created account for sub-gone survived the next poll")
	}
	if f.held(t, "sub-other") != 1 || count(t, `SELECT count(*) FROM users WHERE id = ?`, f.other) != 1 {
		t.Fatalf("another account was touched")
	}
}

// The ledger main hands to oidc.Deps.Erased names the erased subject and no
// other, and a snapshot that still lists the erased person does not give them
// an account again.
func TestAnErasedSubjectIsRefusedAnAccount(t *testing.T) {
	f := newErasedFixture(t)
	f.account(t, "gone@facile.studio", "sub-gone")
	f.poll(t)
	ctx := context.Background()
	for subject, want := range map[string]bool{"sub-gone": true, "sub-other": false} {
		if erased, err := f.ledger.Erased(ctx, f.issuer, subject); err != nil || erased != want {
			t.Fatalf("Erased(%s) = %v, %v; want %v", subject, erased, err, want)
		}
	}

	registre := NewRegistre(testDB, directory.Client{Issuer: f.issuer}, f.ledger, slog.Default())
	snapshot := directory.Snapshot{Version: 1, Spaces: []directory.Space{{ID: "rs-erased", Name: "Studio", Members: []directory.Member{
		{Subject: "sub-other", Email: "other@facile.studio", Role: directory.RoleOwner},
		{Subject: "sub-gone", Email: "gone@facile.studio", Role: directory.RoleMember},
	}}}}
	if err := registre.Apply(ctx, snapshot); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if f.held(t, "sub-gone") != 0 || count(t, `SELECT count(*) FROM users WHERE email = 'gone@facile.studio'`) != 0 {
		t.Fatalf("the snapshot gave the erased subject an account again")
	}
	if count(t, `SELECT count(*) FROM space_members WHERE user_id = ?`, f.other) != 1 {
		t.Fatalf("the snapshot's other member was not applied")
	}
}
