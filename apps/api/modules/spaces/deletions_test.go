package spaces

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/modules/auth"
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
	if err := os.WriteFile(path, []byte("%PDF "+path), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func document(t *testing.T, owner int64, space *int64, name, storage, status string) schemas.Document {
	t.Helper()
	doc := schemas.Document{Name: name, OwnerID: owner, SpaceID: space, StoragePath: storage, Status: status}
	mustCreate(t, &doc)
	mustCreate(t, &schemas.Signer{DocumentID: doc.ID, Name: "Signer", Email: "signer@example.com", Token: name})
	mustCreate(t, &schemas.Field{DocumentID: doc.ID, FieldType: "signature", Value: "signed by " + name})
	return doc
}

type erasureFixture struct {
	gone, other int64
	dir         string
	deletions   *Deletions
	frozen      *[]int64
}

func (f erasureFixture) erase(t *testing.T) {
	t.Helper()
	for range 2 {
		if err := f.deletions.receiver.OnPersonDeleted(context.Background(), testIssuer, "sub-gone"); err != nil {
			t.Fatalf("erase: %v", err)
		}
	}
}

func seedErasure(t *testing.T) erasureFixture {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	newService(t)
	if err := testDB.Exec(`TRUNCATE documents, signers, fields, clients, webhooks, smtp_configs, mcp_actions RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear: %v", err)
	}
	gone := createUser(t, "gone@facile.studio")
	other := createUser(t, "other@facile.studio")
	linkIdentity(t, gone, "sub-gone")
	linkIdentity(t, other, "sub-other")
	dir := t.TempDir()
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	users, err := auth.NewUsers(context.Background(), sqlDB, logger)
	if err != nil {
		t.Fatalf("NewUsers: %v", err)
	}
	frozen := &[]int64{}
	freeze := func(_ context.Context, id int64) error {
		*frozen = append(*frozen, id)
		return nil
	}
	deletions := NewDeletions(DeletionsConfig{
		ORM: testDB, Client: directory.Client{Issuer: testIssuer}, Users: users, Freeze: freeze, UploadDir: dir, Logger: logger,
	})
	return erasureFixture{gone: gone, other: other, dir: dir, deletions: deletions, frozen: frozen}
}

func (f erasureFixture) seedAccount(t *testing.T) (shared, alone, mine schemas.Space) {
	t.Helper()
	shared = schemas.Space{Name: "Studio", OwnerID: f.other}
	mustCreate(t, &shared)
	addMember(t, shared.ID, f.other, "owner")
	addMember(t, shared.ID, f.gone, "member")
	alone = schemas.Space{Name: "Solo", OwnerID: f.gone}
	mustCreate(t, &alone)
	addMember(t, alone.ID, f.gone, "owner")
	mine = schemas.Space{Name: "Mine", OwnerID: f.gone}
	mustCreate(t, &mine)
	addMember(t, mine.ID, f.gone, "owner")
	addMember(t, mine.ID, f.other, "member")

	client := schemas.Client{Name: "Acme", OwnerID: f.gone}
	mustCreate(t, &client)
	mustCreate(t, &schemas.Client{Name: "Theirs", OwnerID: f.other})
	mustCreate(t, &schemas.Webhook{OwnerID: f.gone, URL: "https://example.com/a", Secret: "s"})
	mustCreate(t, &schemas.Webhook{OwnerID: f.other, URL: "https://example.com/b", Secret: "s"})
	mustCreate(t, &schemas.SmtpConfig{OwnerID: f.gone, Host: "smtp.example.com", Password: "secret"})
	statements := []string{
		`INSERT INTO porte_sessions (token_hash, user_id, label) VALUES ('login', ?, ''), ('ardoise', ?, 'ardoise')`,
		`INSERT INTO porte_login_codes (code_hash, user_id, expires_at) VALUES ('code', ?, now())`,
		`INSERT INTO porte_magic_links (token_hash, user_id, created_at, expires_at) VALUES ('magic', ?, now(), now())`,
		`UPDATE users SET avatar_upload_path = 'avatars/user-1.png', oidc_picture_url = 'https://sso.test/avatars/sub-gone',
			oidc_refresh_token = 'refresh', locale = 'fr' WHERE id = ?`,
	}
	for _, statement := range statements {
		args := []any{f.gone}
		if statement == statements[0] {
			args = append(args, f.gone)
		}
		if err := testDB.Exec(statement, args...).Error; err != nil {
			t.Fatalf("seed account: %v", err)
		}
	}
	writeFile(t, filepath.Join(f.dir, "avatars", "user-1.png"))
	return shared, alone, mine
}

func TestErasePersonDeletesTheAccountAndItsDrafts(t *testing.T) {
	f := seedErasure(t)
	shared, alone, _ := f.seedAccount(t)
	document(t, f.gone, nil, "personal-draft", "1/1_a.pdf", "draft")
	document(t, f.gone, &shared.ID, "shared-draft", "1/2_b.pdf", "draft")
	document(t, f.gone, &alone.ID, "alone-draft", "1/3_c.pdf", "draft")
	document(t, f.other, nil, "theirs", "2/4_d.pdf", "draft")
	if err := testDB.Exec(`UPDATE documents SET client_id = (SELECT id FROM clients WHERE name = 'Acme') WHERE name = 'shared-draft'`).Error; err != nil {
		t.Fatalf("attach client: %v", err)
	}
	for _, path := range []string{"1/1_a.pdf", "1/1_a_signed.pdf", "1/2_b.pdf", "1/3_c.pdf", "2/4_d.pdf", "audit/audit_1.pdf"} {
		writeFile(t, filepath.Join(f.dir, path))
	}
	f.erase(t)

	for _, table := range []string{"users WHERE id", "space_members WHERE user_id", "clients WHERE owner_id",
		"webhooks WHERE owner_id", "smtp_configs WHERE owner_id", "porte_identities WHERE user_id",
		"porte_sessions WHERE user_id", "porte_login_codes WHERE user_id", "porte_magic_links WHERE user_id"} {
		if got := count(t, `SELECT count(*) FROM `+table+` = ?`, f.gone); got != 0 {
			t.Errorf("%s = erased account: %d rows, want 0", table, got)
		}
	}
	for name, want := range map[string]int64{"personal-draft": 0, "alone-draft": 0, "shared-draft": 1, "theirs": 1} {
		if got := count(t, `SELECT count(*) FROM documents WHERE name = ?`, name); got != want {
			t.Errorf("documents named %q = %d, want %d", name, got, want)
		}
	}
	if n := count(t, `SELECT count(*) FROM documents WHERE name = 'shared-draft' AND owner_id = ? AND client_id IS NULL`, f.other); n != 1 {
		t.Errorf("the shared draft did not pass to the space's member, or kept a deleted client")
	}
	if n := count(t, `SELECT count(*) FROM signers`) + count(t, `SELECT count(*) FROM fields`); n != 4 {
		t.Errorf("signers and fields = %d, want the 4 of the two remaining drafts", n)
	}
	if n := count(t, `SELECT count(*) FROM spaces WHERE name = 'Solo'`); n != 0 {
		t.Errorf("a space only the erased person belonged to survived")
	}
	if n := count(t, `SELECT count(*) FROM spaces WHERE name = 'Mine' AND owner_id = ?`, f.other); n != 1 {
		t.Errorf("a space shared with others was deleted or not handed to its member")
	}
	if n := count(t, `SELECT count(*) FROM users WHERE id = ?`, f.other) + count(t, `SELECT count(*) FROM webhooks WHERE owner_id = ?`, f.other); n != 2 {
		t.Errorf("another account or its webhook was touched")
	}
	for path, want := range map[string]bool{
		"1/1_a.pdf": false, "1/1_a_signed.pdf": false, "audit/audit_1.pdf": false, "1/3_c.pdf": false,
		"avatars/user-1.png": false, "1/2_b.pdf": true, "2/4_d.pdf": true,
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
