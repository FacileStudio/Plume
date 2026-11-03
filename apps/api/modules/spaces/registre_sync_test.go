package spaces

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
)

// fakeRegistre answers 404 to the first imports, as an undeployed Registre
// would, and records the order of every call it serves.
type fakeRegistre struct {
	mu      sync.Mutex
	calls   []string
	imports []directory.ImportSpace
	fail    int
}

// ServeHTTP records the call and answers it.
func (f *fakeRegistre) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.URL.Path == "/api/directory/import" && f.fail > 0 {
		f.fail--
		f.calls = append(f.calls, "import:404")
		http.NotFound(w, r)
		return
	}
	f.calls = append(f.calls, strings.TrimPrefix(r.URL.Path, "/api/directory/"))
	if r.URL.Path == "/api/directory/import" {
		body, _ := io.ReadAll(r.Body)
		var payload struct{ Spaces []directory.ImportSpace }
		_ = json.Unmarshal(body, &payload)
		f.imports = payload.Spaces
	}
	_, _ = w.Write([]byte(`{"version":1,"spaces":[]}`))
}

// snapshot reports the calls so far, and whether a snapshot was pulled.
func (f *fakeRegistre) snapshot() ([]string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	calls := append([]string(nil), f.calls...)
	return calls, strings.Contains(strings.Join(calls, ","), "snapshot")
}

// No snapshot is pulled before Registre has accepted the import: applying one
// first would remove every member Registre has not heard of. A Registre that
// answers 404 is retried rather than fatal.
func TestImportRunsBeforeTheFirstSync(t *testing.T) {
	newService(t)
	alice := createUser(t, "alice@facile.studio")
	carol := createUser(t, "carol@facile.studio")
	fake := &fakeRegistre{fail: 2}
	server := httptest.NewServer(fake)
	defer server.Close()
	if err := testDB.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, 'sub-alice')`,
		alice, server.URL).Error; err != nil {
		t.Fatalf("link identity: %v", err)
	}
	local := schemas.Space{Name: "Contracts", OwnerID: alice}
	testDB.Create(&local)
	addMember(t, local.ID, alice, RoleOwner)
	addMember(t, local.ID, carol, RoleMember)
	registre := NewRegistre(testDB, directory.Client{Issuer: server.URL}, slog.Default())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go registre.Run(ctx, time.Millisecond)
	calls, pulled := fake.snapshot()
	for ; !pulled && ctx.Err() == nil; calls, pulled = fake.snapshot() {
		time.Sleep(5 * time.Millisecond)
	}
	cancel()

	if got := strings.Join(calls, ","); !strings.HasPrefix(got, "import:404,import:404,import,snapshot") {
		t.Fatalf("calls = %s, want two failed imports, the import, then the snapshot", got)
	}
	if len(fake.imports) != 1 {
		t.Fatalf("imported %+v, want one space", fake.imports)
	}
	members := fake.imports[0].Members
	if fake.imports[0].Name != "Contracts" || len(members) != 2 ||
		members[0] != (directory.ImportMember{Subject: "sub-alice", Role: RoleOwner}) ||
		members[1] != (directory.ImportMember{Email: "carol@facile.studio", Role: RoleMember}) {
		t.Fatalf("imported %+v", fake.imports)
	}
}
