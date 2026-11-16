package webhooks

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/internal/secrets"
	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/testdb"

	portepg "github.com/FacileStudio/porte/pg"
	"gorm.io/gorm"
)

const (
	fixedSecret    = "whsec_fixed_vector"
	fixedSignature = "cee1af0df8c7132adfa89266e0d76fb3b20502da2dae1ca6192c7efbafe2df71"
)

var fixedBody = []byte(`{"event_type":"document.sent","n":1}`)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	url, ok := testdb.URL()
	if !ok {
		testdb.Announce("sh scripts/postgres-up.sh")
		os.Exit(m.Run())
	}
	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_webhooks_test", Migrate: schemas.Migrate})
	if err != nil {
		panic(err)
	}
	testDB = db
	os.Exit(m.Run())
}

func sealerOf(t *testing.T, fill string) *portepg.TokenSealer {
	t.Helper()
	sealer, err := portepg.ParseTokenSealer(strings.Repeat(fill, 32))
	if err != nil {
		t.Fatal(err)
	}
	return sealer
}

func requireDB(t *testing.T) *gorm.DB {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testDB.Exec(`TRUNCATE users, webhooks RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatal(err)
	}
	return testDB
}

func rawSecret(t *testing.T, db *gorm.DB, id int64) string {
	t.Helper()
	var stored string
	if err := db.Raw(`SELECT secret FROM webhooks WHERE id = ?`, id).Scan(&stored).Error; err != nil {
		t.Fatal(err)
	}
	return stored
}

func TestTheHMACKeepsTheSameVector(t *testing.T) {
	if got := sign(fixedSecret, fixedBody); got != fixedSignature {
		t.Fatalf("sign() = %s, want the pinned vector %s", got, fixedSignature)
	}
}

func TestWebhookSecretIsCiphertextAtRestAndSignsTheSameVector(t *testing.T) {
	db := requireDB(t)
	owner := schemas.User{Email: "owner@facile.studio", Name: "Owner"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	service := NewService(db).WithSealer(sealerOf(t, "ab"))
	created, err := service.create(context.Background(), owner.ID, &CreateWebhookRequest{URL: "https://example.com/hook", Secret: fixedSecret})
	if err != nil {
		t.Fatal(err)
	}
	stored := rawSecret(t, db, created.ID)
	if !secrets.Sealed(stored) || strings.Contains(stored, fixedSecret) {
		t.Fatalf("the stored secret must be ciphertext, got %q", stored)
	}
	record, err := service.findWebhook(context.Background(), owner.ID, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := service.signature(record, fixedBody); err != nil || got != fixedSignature {
		t.Fatalf("signature %q %v, want %s", got, err, fixedSignature)
	}

	rotated := "whsec_rotated"
	if _, err := service.update(context.Background(), owner.ID, created.ID, &UpdateWebhookRequest{URL: "https://example.com/hook", Secret: rotated, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if after := rawSecret(t, db, created.ID); !secrets.Sealed(after) || after == stored {
		t.Fatalf("an updated secret must be sealed anew, got %q", after)
	}
}

func TestConvertingALegacyWebhookKeepsTheReceiversHMAC(t *testing.T) {
	db := requireDB(t)
	legacy := schemas.Webhook{OwnerID: 1, URL: "https://example.com/hook", Secret: fixedSecret, Enabled: true}
	if err := db.Create(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	sealer := sealerOf(t, "ab")
	service := NewService(db).WithSealer(sealer)

	before := legacy
	if got, err := service.signature(&before, fixedBody); err != nil || got != fixedSignature {
		t.Fatalf("before conversion: %q %v", got, err)
	}
	if _, err := secrets.Convert(context.Background(), db, sealer, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatal(err)
	}
	var after schemas.Webhook
	if err := db.First(&after, legacy.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !secrets.Sealed(after.Secret) {
		t.Fatalf("the secret was not converted: %q", after.Secret)
	}
	if got, err := service.signature(&after, fixedBody); err != nil || got != fixedSignature {
		t.Fatalf("after conversion: %q %v, want %s", got, err, fixedSignature)
	}
}

func TestDeliveryFailsWhenTheSealedSecretIsUnreadable(t *testing.T) {
	sealed, err := sealerOf(t, "ab").Seal(fixedSecret)
	if err != nil {
		t.Fatal(err)
	}
	record := &schemas.Webhook{URL: "https://example.com/hook", Secret: sealed}
	for name, service := range map[string]*Service{
		"wrong key": NewService(nil).WithSealer(sealerOf(t, "cd")),
		"no key":    NewService(nil),
	} {
		got, err := service.signature(record, fixedBody)
		if err != errSecretUnreadable || got != "" {
			t.Fatalf("%s: got %q %v, want errSecretUnreadable and no signature", name, got, err)
		}
	}
}
