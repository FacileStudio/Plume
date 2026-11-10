package spaces

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/schemas"
)

const keptRows = `SELECT
	(SELECT jsonb_agg(to_jsonb(x) ORDER BY x.id) FROM documents x WHERE x.id IN (%[1]s))::text,
	(SELECT jsonb_agg(to_jsonb(x) ORDER BY x.id) FROM signers x WHERE x.document_id IN (%[1]s))::text,
	(SELECT jsonb_agg(to_jsonb(x) ORDER BY x.id) FROM fields x WHERE x.document_id IN (%[1]s))::text`

const textColumns = `SELECT table_name, column_name FROM information_schema.columns
	WHERE table_schema = current_schema() AND data_type IN ('text', 'character varying', 'json', 'jsonb')
	ORDER BY table_name, column_name`

func snapshotKept(t *testing.T, ids string) string {
	t.Helper()
	var docs, signers, fields string
	if err := testDB.Raw(fmt.Sprintf(keptRows, ids)).Row().Scan(&docs, &signers, &fields); err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return docs + "\n" + signers + "\n" + fields
}

func hashFiles(t *testing.T, dir string, paths []string) map[string]string {
	t.Helper()
	sums := map[string]string{}
	for _, path := range paths {
		data, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		sum := sha256.Sum256(data)
		sums[path] = hex.EncodeToString(sum[:])
	}
	return sums
}

// leftovers names every text column outside the kept legal records that
// still holds one of the needles.
func leftovers(t *testing.T, ids string, needles ...string) []string {
	t.Helper()
	var columns []struct{ TableName, ColumnName string }
	if err := testDB.Raw(textColumns).Scan(&columns).Error; err != nil {
		t.Fatalf("list columns: %v", err)
	}
	exempt := map[string]string{"documents": "id", "signers": "document_id", "fields": "document_id"}
	var found []string
	for _, c := range columns {
		where := fmt.Sprintf(`%q::text ILIKE ?`, c.ColumnName)
		if key, ok := exempt[c.TableName]; ok {
			where += fmt.Sprintf(` AND %q NOT IN (%s)`, key, ids)
		}
		for _, needle := range needles {
			if n := count(t, fmt.Sprintf(`SELECT count(*) FROM %q WHERE %s`, c.TableName, where), "%"+needle+"%"); n > 0 {
				found = append(found, fmt.Sprintf("%s.%s has %q (%d)", c.TableName, c.ColumnName, needle, n))
			}
		}
	}
	return found
}

func (f erasureFixture) seedRecords(t *testing.T, shared, alone schemas.Space) (string, []string) {
	t.Helper()
	if err := testDB.Exec(`UPDATE users SET name = 'Gone Person' WHERE id = ?`, f.gone).Error; err != nil {
		t.Fatalf("name: %v", err)
	}
	signedAt := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	sent := document(t, f.gone, nil, "sent", "1/10_s.pdf", "pending")
	signed := document(t, f.gone, nil, "signed", "1/11_t.pdf", "completed")
	declined := document(t, f.gone, &alone.ID, "declined-alone", "1/12_u.pdf", "declined")
	inShared := document(t, f.gone, &shared.ID, "signed-shared", "1/13_v.pdf", "completed")
	theirs := document(t, f.other, nil, "signed-by-gone", "2/14_w.pdf", "completed")
	mustCreate(t, &schemas.Signer{DocumentID: theirs.ID, Name: "Gone Person", Email: "gone@facile.studio",
		Token: "gone-token", Status: "signed", SignedAt: &signedAt, IPAddress: "192.0.2.9", UserAgent: "Firefox"})
	document(t, f.gone, nil, "draft", "1/15_x.pdf", "draft")
	statements := []string{
		`UPDATE documents SET signed_hash = 'abc', original_hash = 'def',
			client_id = (SELECT id FROM clients WHERE name = 'Acme') WHERE name = 'signed'`,
		`UPDATE documents SET external_ref = 'ardoise:42', nuage_linked_to = 'ardoise:42' WHERE name = 'signed-shared'`,
	}
	for _, statement := range statements {
		if err := testDB.Exec(statement).Error; err != nil {
			t.Fatalf("seed records: %v", err)
		}
	}
	files := []string{"1/10_s.pdf", "1/11_t.pdf", "1/11_t_signed.pdf", "1/12_u.pdf", "1/13_v.pdf", "2/14_w.pdf",
		fmt.Sprintf("audit/audit_%d.pdf", signed.ID), fmt.Sprintf("certificates/cert_%d.pdf", signed.ID)}
	for _, path := range append(files, "1/15_x.pdf") {
		writeFile(t, filepath.Join(f.dir, path))
	}
	ids := fmt.Sprintf("%d, %d, %d, %d, %d", sent.ID, signed.ID, declined.ID, inShared.ID, theirs.ID)
	return ids, files
}

func TestErasePersonLeavesEverySentDocumentByteIdentical(t *testing.T) {
	f := seedErasure(t)
	shared, alone, _ := f.seedAccount(t)
	ids, files := f.seedRecords(t, shared, alone)
	rowsBefore, filesBefore := snapshotKept(t, ids), hashFiles(t, f.dir, files)
	if found := leftovers(t, ids, "gone@facile.studio", "Gone Person", "sub-gone"); len(found) < 3 {
		t.Fatalf("the scan misses the account before erasure: %v", found)
	}

	f.erase(t)

	if after := snapshotKept(t, ids); after != rowsBefore {
		t.Errorf("sent documents changed:\nbefore %s\nafter  %s", rowsBefore, after)
	}
	if after := hashFiles(t, f.dir, files); fmt.Sprint(after) != fmt.Sprint(filesBefore) {
		t.Errorf("sent documents' files changed: %v, want %v", after, filesBefore)
	}
	if exists(filepath.Join(f.dir, "1/15_x.pdf")) || count(t, `SELECT count(*) FROM documents WHERE name = 'draft'`) != 0 {
		t.Errorf("the draft that was never sent survived")
	}
	if found := leftovers(t, ids, "gone@facile.studio", "Gone Person", "sub-gone"); len(found) > 0 {
		t.Errorf("the erased person is still named outside the legal records: %v", found)
	}
	if want := []int64{2, 3, 4}; !slices.Equal(*f.frozen, want) {
		t.Errorf("frozen audit trails = %v, want the erased owner's completed and declined documents %v", *f.frozen, want)
	}
	if n := count(t, `SELECT count(*) FROM spaces WHERE name = 'Solo'`); n != 1 {
		t.Errorf("the space holding a declined document was deleted")
	}
	if n := count(t, `SELECT count(*) FROM clients WHERE name = 'Acme'`); n != 1 {
		t.Errorf("the client a signed document names was deleted")
	}
}
