package signing

import (
	"context"
	"os"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/testdb"
)

func TestFreezeAuditTrailWritesOnceAndKeepsTheOwner(t *testing.T) {
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_signing_test", Migrate: schemas.Migrate})
	if err != nil {
		t.Fatalf("test database: %v", err)
	}
	if err := db.Exec(`TRUNCATE users, documents, signers, fields RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear: %v", err)
	}
	owner := schemas.User{Email: "owner@facile.studio", Name: "Owner"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatalf("seed owner: %v", err)
	}
	doc := schemas.Document{Name: "Contract", OwnerID: owner.ID, Status: "completed"}
	if err := db.Create(&doc).Error; err != nil {
		t.Fatalf("seed document: %v", err)
	}
	service := NewService(db, t.TempDir(), nil)

	if err := service.FreezeAuditTrail(context.Background(), doc.ID); err != nil {
		t.Fatalf("freeze: %v", err)
	}
	first, err := os.Stat(service.auditTrailPath(doc.ID))
	if err != nil || first.Size() == 0 {
		t.Fatalf("no audit trail written: %v", err)
	}
	if err := db.Delete(&owner).Error; err != nil {
		t.Fatalf("erase owner: %v", err)
	}
	if err := service.FreezeAuditTrail(context.Background(), doc.ID); err != nil {
		t.Fatalf("second freeze: %v", err)
	}
	second, err := os.Stat(service.auditTrailPath(doc.ID))
	if err != nil || !second.ModTime().Equal(first.ModTime()) || second.Size() != first.Size() {
		t.Fatalf("a frozen audit trail was rewritten after the owner was erased")
	}
}
