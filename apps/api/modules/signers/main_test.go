package signers

import (
	"os"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	url, ok := testdb.URL()
	if !ok {
		testdb.Announce("sh scripts/postgres-up.sh")
		os.Exit(m.Run())
	}

	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_signers_test", Migrate: schemas.Migrate})
	if err != nil {
		panic(err)
	}
	testDB = db
	os.Exit(m.Run())
}

// requireDB hands back a database with empty document, signer and field tables, or skips
// loudly. Postgres is the only database this suite runs on, tests included: SQLite would
// build a different schema from the struct tags and then pass, proving nothing about the
// DDL that ships.
func requireDB(t *testing.T) *gorm.DB {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testDB.Exec(`TRUNCATE documents, signers, fields RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear tables: %v", err)
	}
	return testDB
}
