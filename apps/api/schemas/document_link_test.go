package schemas

import "testing"

func TestMigrateAddsLinkColumnsToAPopulatedDocumentsTable(t *testing.T) {
	db := requireDB(t)
	for _, column := range []string{"external_ref", "nuage_space_facile_id", "nuage_linked_to", "nuage_published_at"} {
		if err := db.Exec(`ALTER TABLE documents DROP COLUMN IF EXISTS ` + column).Error; err != nil {
			t.Fatalf("drop %s: %v", column, err)
		}
	}
	if err := db.Exec(`DROP INDEX IF EXISTS idx_documents_owner_external_ref`).Error; err != nil {
		t.Fatalf("drop index: %v", err)
	}
	if err := db.Exec(`TRUNCATE documents RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear documents: %v", err)
	}
	if err := db.Exec(`INSERT INTO documents (name, status, owner_id) VALUES ('old-a', 'completed', 1), ('old-b', 'draft', 1)`).Error; err != nil {
		t.Fatalf("insert legacy documents: %v", err)
	}

	for i := 0; i < 2; i++ {
		if err := Migrate(db); err != nil {
			t.Fatalf("migrate %d: %v", i+1, err)
		}
	}

	var rows []Document
	if err := db.Order("id").Find(&rows).Error; err != nil || len(rows) != 2 {
		t.Fatalf("read documents: %v (%d rows)", err, len(rows))
	}
	for _, row := range rows {
		if row.ExternalRef != "" || row.NuageSpaceFacileID != "" || row.NuageLinkedTo != "" || row.NuagePublishedAt != nil {
			t.Errorf("legacy document %s picked up link data: %+v", row.Name, row)
		}
	}

	if err := db.Create(&Document{Name: "a", OwnerID: 1, ExternalRef: "x"}).Error; err != nil {
		t.Fatalf("first linked document: %v", err)
	}
	if err := db.Create(&Document{Name: "b", OwnerID: 1, ExternalRef: "x"}).Error; err == nil {
		t.Error("two documents of one owner share an external_ref")
	}
	if err := db.Create(&Document{Name: "c", OwnerID: 2, ExternalRef: "x"}).Error; err != nil {
		t.Errorf("another owner may reuse the ref: %v", err)
	}
	if err := db.Create(&Document{Name: "d", OwnerID: 1}).Error; err != nil {
		t.Errorf("unlinked documents must stay unconstrained: %v", err)
	}
}
