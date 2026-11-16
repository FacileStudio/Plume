package secrets

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/testdb"

	portepg "github.com/FacileStudio/porte/pg"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	url, ok := testdb.URL()
	if !ok {
		testdb.Announce("sh scripts/postgres-up.sh")
		os.Exit(m.Run())
	}
	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_secrets_test", Migrate: schemas.Migrate})
	if err != nil {
		panic(err)
	}
	testDB = db
	os.Exit(m.Run())
}

func requireDB(t *testing.T) *gorm.DB {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testDB.Exec(`TRUNCATE signers, smtp_configs, webhooks RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear tables: %v", err)
	}
	return testDB
}

func sealerOf(t *testing.T, fill string) *portepg.TokenSealer {
	t.Helper()
	sealer, err := portepg.ParseTokenSealer(strings.Repeat(fill, 32))
	if err != nil {
		t.Fatalf("sealer: %v", err)
	}
	return sealer
}

func quiet() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func column(t *testing.T, db *gorm.DB, table, column string, id int64) string {
	t.Helper()
	var value string
	if err := db.Raw(`SELECT coalesce(`+column+`, '') FROM `+table+` WHERE id = ?`, id).Scan(&value).Error; err != nil {
		t.Fatalf("read %s.%s: %v", table, column, err)
	}
	return value
}

func TestHashTokenIsSHA256Hex(t *testing.T) {
	const want = "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if got := HashToken("test"); got != want {
		t.Fatalf("HashToken(test) = %s, want %s", got, want)
	}
}

func TestSealedValueNeverPassesAsALegacyToken(t *testing.T) {
	sealer := sealerOf(t, "ab")
	sealed, err := sealer.Seal("0123abcd")
	if err != nil {
		t.Fatal(err)
	}
	if legacyToken(sealed) || legacyToken("pv1:anything") || legacyToken("a:b") {
		t.Fatal("a value with the prefix or a colon must not fall back to the plaintext column")
	}
	if !legacyToken("0123abcd") {
		t.Fatal("a plain hex token must be allowed to fall back")
	}
}

func seedSealed(t *testing.T, db *gorm.DB, token string) (schemas.Signer, string) {
	t.Helper()
	stored, hash, err := SealToken(sealerOf(t, "ab"), token)
	if err != nil {
		t.Fatal(err)
	}
	row := schemas.Signer{Name: "A", Email: "a@x.test", Token: stored, TokenHash: hash}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	return row, stored
}

func wantNotFound(t *testing.T, db *gorm.DB, input, why string) {
	t.Helper()
	if _, err := FindSigner(db, input); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("%s, got %v", why, err)
	}
}

func TestFindSignerByHashAndNeverBySealedColumn(t *testing.T) {
	db := requireDB(t)
	const token = "feedfacefeedfacefeedfacefeedfacefeedfacefeedfacefeedfacefeedface"
	row, stored := seedSealed(t, db, token)
	if !Sealed(stored) || row.TokenHash != HashToken(token) {
		t.Fatalf("stored %q hash %q", stored, row.TokenHash)
	}
	if found, err := FindSigner(db, token); err != nil || found.ID != row.ID {
		t.Fatalf("the real token must find the signer: %v", err)
	}
	wantNotFound(t, db, stored, "the sealed column value must not work as a token")
	wantNotFound(t, db, "", "an empty token matches nothing")
}

func TestSealedValueOnARowWithoutAHashIsNotAToken(t *testing.T) {
	db := requireDB(t)
	_, stored := seedSealed(t, db, "feedface")
	unhashed := schemas.Signer{Name: "B", Email: "b@x.test", Token: stored + "x"}
	if err := db.Create(&unhashed).Error; err != nil {
		t.Fatal(err)
	}
	wantNotFound(t, db, stored+"x", "a sealed value on a row without a hash must not work as a token")
}

func TestLegacyPlaintextRowStillOpensUntilConverted(t *testing.T) {
	db := requireDB(t)
	row := schemas.Signer{Name: "A", Email: "a@x.test", Token: "legacyplaintexttoken"}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	if found, err := FindSigner(db, "legacyplaintexttoken"); err != nil || found.ID != row.ID {
		t.Fatalf("an unconverted plaintext row must still open: %v", err)
	}
}
