package spaces

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/tronc/testdb"
)

func count(t *testing.T, query string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := testDB.Raw(query, args...).Scan(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func mustCreate(t *testing.T, row any) {
	t.Helper()
	if err := testDB.Create(row).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("%PDF"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func document(t *testing.T, owner int64, space *int64, name, storage string) schemas.Document {
	t.Helper()
	doc := schemas.Document{Name: name, OwnerID: owner, SpaceID: space, StoragePath: storage, Status: "pending"}
	mustCreate(t, &doc)
	mustCreate(t, &schemas.Signer{DocumentID: doc.ID, Name: "Signer", Email: "signer@example.com", Token: name})
	mustCreate(t, &schemas.Field{DocumentID: doc.ID, FieldType: "signature"})
	return doc
}

type erasureFixture struct {
	gone, other int64
	dir         string
	deletions   *Deletions
}

func seedErasure(t *testing.T) erasureFixture {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	newService(t)
	if err := testDB.Exec(`TRUNCATE documents, signers, fields, clients, webhooks, smtp_configs RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear: %v", err)
	}
	gone := createUser(t, "gone@facile.studio")
	other := createUser(t, "other@facile.studio")
	linkIdentity(t, gone, "sub-gone")
	linkIdentity(t, other, "sub-other")

	shared := schemas.Space{Name: "Studio", OwnerID: other}
	mustCreate(t, &shared)
	addMember(t, shared.ID, other, "owner")
	addMember(t, shared.ID, gone, "member")
	alone := schemas.Space{Name: "Solo", OwnerID: gone}
	mustCreate(t, &alone)
	addMember(t, alone.ID, gone, "owner")
	mine := schemas.Space{Name: "Mine", OwnerID: gone}
	mustCreate(t, &mine)
	addMember(t, mine.ID, gone, "owner")
	addMember(t, mine.ID, other, "member")

	dir := t.TempDir()
	document(t, gone, nil, "personal", "1/1_a.pdf")
	document(t, gone, &shared.ID, "in-shared", "1/2_b.pdf")
	document(t, gone, &alone.ID, "in-alone", "1/3_c.pdf")
	document(t, other, nil, "theirs", "2/4_d.pdf")
	for _, path := range []string{"1/1_a.pdf", "1/1_a_signed.pdf", "1/2_b.pdf", "1/3_c.pdf", "2/4_d.pdf"} {
		writeFile(t, filepath.Join(dir, path))
	}
	writeFile(t, filepath.Join(dir, "audit", "audit_1.pdf"))

	client := schemas.Client{Name: "Acme", OwnerID: gone}
	mustCreate(t, &client)
	if err := testDB.Model(&schemas.Document{}).Where("name = ?", "in-shared").Update("client_id", client.ID).Error; err != nil {
		t.Fatalf("attach client: %v", err)
	}
	mustCreate(t, &schemas.Client{Name: "Theirs", OwnerID: other})
	mustCreate(t, &schemas.Webhook{OwnerID: gone, URL: "https://example.com/a", Secret: "s"})
	mustCreate(t, &schemas.Webhook{OwnerID: other, URL: "https://example.com/b", Secret: "s"})
	mustCreate(t, &schemas.SmtpConfig{OwnerID: gone, Host: "smtp.example.com", Password: "secret"})

	deletions := NewDeletions(testDB, directory.Client{Issuer: testIssuer}, nil, dir,
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	return erasureFixture{gone: gone, other: other, dir: dir, deletions: deletions}
}

func TestErasePersonDeletesPersonalDataAndKeepsSharedDocuments(t *testing.T) {
	f := seedErasure(t)
	for range 2 {
		if err := f.deletions.ErasePerson(context.Background(), testIssuer, "sub-gone"); err != nil {
			t.Fatalf("erase: %v", err)
		}
	}

	for query, want := range map[string]int64{
		`SELECT count(*) FROM users WHERE id = ?`:                 f.gone,
		`SELECT count(*) FROM space_members WHERE user_id = ?`:    f.gone,
		`SELECT count(*) FROM clients WHERE owner_id = ?`:         f.gone,
		`SELECT count(*) FROM webhooks WHERE owner_id = ?`:        f.gone,
		`SELECT count(*) FROM smtp_configs WHERE owner_id = ?`:    f.gone,
		`SELECT count(*) FROM porte_identities WHERE user_id = ?`: f.gone,
	} {
		if got := count(t, query, want); got != 0 {
			t.Errorf("%s = %d, want 0", query, got)
		}
	}
	for name, want := range map[string]int64{"personal": 0, "in-alone": 0, "in-shared": 1, "theirs": 1} {
		if got := count(t, `SELECT count(*) FROM documents WHERE name = ?`, name); got != want {
			t.Errorf("documents named %q = %d, want %d", name, got, want)
		}
	}
	if n := count(t, `SELECT count(*) FROM documents WHERE name = 'in-shared' AND owner_id = 0 AND client_id IS NULL`); n != 1 {
		t.Errorf("the shared document kept its owner or client")
	}
	if n := count(t, `SELECT count(*) FROM signers`); n != 2 {
		t.Errorf("signers = %d, want the 2 of the kept documents", n)
	}
	if n := count(t, `SELECT count(*) FROM fields`); n != 2 {
		t.Errorf("fields = %d, want the 2 of the kept documents", n)
	}
	if n := count(t, `SELECT count(*) FROM spaces WHERE name = 'Solo'`); n != 0 {
		t.Errorf("a space only the erased person belonged to survived")
	}
	if n := count(t, `SELECT count(*) FROM spaces WHERE name = 'Mine' AND owner_id = 0`); n != 1 {
		t.Errorf("a space shared with others was deleted or kept its owner")
	}
	if n := count(t, `SELECT count(*) FROM users WHERE id = ?`, f.other); n != 1 {
		t.Errorf("another account was touched")
	}
	if n := count(t, `SELECT count(*) FROM webhooks WHERE owner_id = ?`, f.other); n != 1 {
		t.Errorf("another account's webhook was touched")
	}
	for path, want := range map[string]bool{
		"1/1_a.pdf": false, "1/1_a_signed.pdf": false, "audit/audit_1.pdf": false,
		"1/3_c.pdf": false, "1/2_b.pdf": true, "2/4_d.pdf": true,
	} {
		if got := exists(filepath.Join(f.dir, path)); got != want {
			t.Errorf("file %s exists = %v, want %v", path, got, want)
		}
	}
}

func TestDeletionStateSurvivesARestart(t *testing.T) {
	seedErasure(t)
	state := deletionState{orm: testDB}
	if err := testDB.Exec(`DELETE FROM directory_deletion_state`).Error; err != nil {
		t.Fatalf("clear: %v", err)
	}

	if saved, err := state.Load(context.Background()); err != nil || saved != nil {
		t.Fatalf("empty Load = %q, %v; want nil", saved, err)
	}
	for _, blob := range []string{"one", "two"} {
		if err := state.Save(context.Background(), []byte(blob)); err != nil {
			t.Fatalf("save: %v", err)
		}
	}
	if saved, err := state.Load(context.Background()); err != nil || string(saved) != "two" {
		t.Fatalf("Load = %q, %v; want two", saved, err)
	}
}
