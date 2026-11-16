package smtp

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/internal/smtptest"
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
	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_smtp_test", Migrate: schemas.Migrate})
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

func savedConfig(t *testing.T, service *Service, server *smtptest.Server) string {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testDB.Exec(`TRUNCATE users, smtp_configs RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatal(err)
	}
	owner := schemas.User{Email: "owner@facile.studio", Name: "Owner"}
	if err := testDB.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	_, err := service.saveConfig(context.Background(), owner.ID, &SaveConfigRequest{
		Host: server.Host(), Port: server.Port(), Username: "mailer", Password: "hunter2", FromEmail: "from@facile.studio", FromName: "Plume",
	})
	if err != nil {
		t.Fatal(err)
	}
	var stored string
	if err := testDB.Raw(`SELECT password FROM smtp_configs WHERE owner_id = ?`, owner.ID).Scan(&stored).Error; err != nil {
		t.Fatal(err)
	}
	return stored
}

func TestSMTPPasswordIsCiphertextAtRestAndOpenedToSend(t *testing.T) {
	server := smtptest.Start(t)
	service := NewService(testDB).AllowPrivateHosts(true).WithSealer(sealerOf(t, "ab"))
	stored := savedConfig(t, service, server)
	if !strings.HasPrefix(stored, "pv1:") || strings.Contains(stored, "hunter2") {
		t.Fatalf("the stored password must be ciphertext, got %q", stored)
	}
	if err := service.testConfig(context.Background(), ownerOf(t), "to@example.test"); err != nil {
		t.Fatalf("test send: %v", err)
	}
	server.WaitBodies(t, 1)
	if got := server.Passwords(); len(got) != 1 || got[0] != "hunter2" {
		t.Fatalf("the server must receive the opened password, got %v", got)
	}
}

func TestSMTPPasswordStaysPlainWithoutAKey(t *testing.T) {
	server := smtptest.Start(t)
	service := NewService(testDB).AllowPrivateHosts(true)
	if stored := savedConfig(t, service, server); stored != "hunter2" {
		t.Fatalf("without a key the password is stored as typed, got %q", stored)
	}
	if err := service.testConfig(context.Background(), ownerOf(t), "to@example.test"); err != nil {
		t.Fatalf("test send: %v", err)
	}
	if got := server.Passwords(); len(got) != 1 || got[0] != "hunter2" {
		t.Fatalf("passwords %v", got)
	}
}

func TestSMTPSendFailsWhenTheSealedPasswordIsUnreadable(t *testing.T) {
	server := smtptest.Start(t)
	sealed := NewService(testDB).AllowPrivateHosts(true).WithSealer(sealerOf(t, "ab"))
	savedConfig(t, sealed, server)
	for name, service := range map[string]*Service{
		"wrong key": NewService(testDB).AllowPrivateHosts(true).WithSealer(sealerOf(t, "cd")),
		"no key":    NewService(testDB).AllowPrivateHosts(true),
	} {
		err := service.testConfig(context.Background(), ownerOf(t), "to@example.test")
		if err == nil || !strings.Contains(err.Error(), "cannot be read") {
			t.Fatalf("%s: the send must fail with a clear error, got %v", name, err)
		}
	}
	if got := server.Passwords(); len(got) != 0 {
		t.Fatalf("the ciphertext must never be sent as the password, server saw %v", got)
	}
	record := schemas.SmtpConfig{Password: "pv1:garbage"}
	if _, err := NewService(testDB).openPassword(record.Password); !errors.Is(err, errPasswordUnreadable) {
		t.Fatalf("got %v", err)
	}
}

func ownerOf(t *testing.T) int64 {
	t.Helper()
	var id int64
	if err := testDB.Raw(`SELECT id FROM users LIMIT 1`).Scan(&id).Error; err != nil {
		t.Fatal(err)
	}
	return id
}
