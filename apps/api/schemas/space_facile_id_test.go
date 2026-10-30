package schemas

import (
	"regexp"
	"testing"

	"gorm.io/gorm"
)

var facileIDShape = regexp.MustCompile(`^fac_[0-9a-f]{24}$`)

func TestMigrateGivesExistingSpacesAFacileID(t *testing.T) {
	db := requireDB(t)

	if err := db.Exec(`DROP INDEX IF EXISTS idx_spaces_facile_id`).Error; err != nil {
		t.Fatalf("drop index: %v", err)
	}
	if err := db.Exec(`TRUNCATE spaces RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear spaces: %v", err)
	}
	if err := db.Exec(`INSERT INTO spaces (name, owner_id) VALUES ('legacy-null-a', 1), ('legacy-null-b', 1), ('legacy-empty', 1)`).Error; err != nil {
		t.Fatalf("insert legacy spaces: %v", err)
	}
	if err := db.Exec(`UPDATE spaces SET facile_id = NULL WHERE name <> 'legacy-empty'`).Error; err != nil {
		t.Fatalf("null facile_id: %v", err)
	}
	if err := db.Exec(`UPDATE spaces SET facile_id = '' WHERE name = 'legacy-empty'`).Error; err != nil {
		t.Fatalf("empty facile_id: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	first := spaceFacileIDs(t, db)
	seen := map[string]bool{}
	for name, id := range first {
		if !facileIDShape.MatchString(id) {
			t.Errorf("space %s got facile_id %q, want fac_ plus 24 hex", name, id)
		}
		if seen[id] {
			t.Errorf("facile_id %q was given to two spaces", id)
		}
		seen[id] = true
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	for name, id := range spaceFacileIDs(t, db) {
		if first[name] != id {
			t.Errorf("space %s changed facile_id from %q to %q on a second migrate", name, first[name], id)
		}
	}
}

func spaceFacileIDs(t *testing.T, db *gorm.DB) map[string]string {
	t.Helper()
	var rows []struct {
		Name     string
		FacileID string
	}
	if err := db.Raw(`SELECT name, facile_id FROM spaces`).Scan(&rows).Error; err != nil {
		t.Fatalf("read spaces: %v", err)
	}
	ids := make(map[string]string, len(rows))
	for _, row := range rows {
		ids[row.Name] = row.FacileID
	}
	return ids
}
