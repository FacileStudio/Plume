package documents

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/modules/smtp"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
	"github.com/FacileStudio/Plume/apps/api/schemas"
)

func completedDocument(t *testing.T) (*Service, int64, string) {
	t.Helper()
	_, owner := signatureService(t)
	dir := t.TempDir()
	svc := NewService(testDB, smtp.NewService(testDB), webhooks.NewService(testDB), "http://plume.test", dir)
	if err := os.WriteFile(filepath.Join(dir, "doc.pdf"), twoPagePDF(t), 0o600); err != nil {
		t.Fatalf("write original: %v", err)
	}
	doc := schemas.Document{Name: "Frozen", FileName: "doc.pdf", Status: "completed", OwnerID: owner, StoragePath: "doc.pdf"}
	if err := testDB.Create(&doc).Error; err != nil {
		t.Fatalf("create document: %v", err)
	}
	signer := schemas.Signer{DocumentID: doc.ID, Name: "A", Email: "a@x.test", Token: "tok-frozen", Status: "signed", Role: "signer"}
	if err := testDB.Create(&signer).Error; err != nil {
		t.Fatalf("create signer: %v", err)
	}
	field := schemas.Field{DocumentID: doc.ID, SignerID: signer.ID, Page: 1, X: 10, Y: 10, Width: 100, Height: 30, FieldType: "text", Value: "hello"}
	if err := testDB.Create(&field).Error; err != nil {
		t.Fatalf("create field: %v", err)
	}
	return svc, doc.ID, filepath.Join(dir, "doc_signed.pdf")
}

func storedHash(t *testing.T, id int64) string {
	t.Helper()
	var doc schemas.Document
	if err := testDB.First(&doc, id).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	return doc.SignedHash
}

func TestCompletedSignedFileIgnoresMtimes(t *testing.T) {
	svc, id, signedPath := completedDocument(t)
	ctx := context.Background()

	if _, err := svc.GetFilePathByDocID(ctx, id); err != nil {
		t.Fatalf("first fetch: %v", err)
	}
	before, err := os.ReadFile(signedPath)
	if err != nil {
		t.Fatalf("read signed: %v", err)
	}
	hash := storedHash(t, id)
	if hash == "" {
		t.Fatal("signed_hash was not recorded on first fetch")
	}

	future := time.Now().Add(48 * time.Hour)
	if err := os.Chtimes(filepath.Join(filepath.Dir(signedPath), "doc.pdf"), future, future); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
	if _, err := svc.GetFilePathByDocID(ctx, id); err != nil {
		t.Fatalf("second fetch: %v", err)
	}
	after, _ := os.ReadFile(signedPath)
	if !bytes.Equal(before, after) {
		t.Fatal("signed file changed after the original's mtime moved")
	}
	if got := storedHash(t, id); got != hash {
		t.Fatalf("signed_hash changed: %s -> %s", hash, got)
	}
}

func TestCompletedMissingSignedFileKeepsStoredHash(t *testing.T) {
	svc, id, signedPath := completedDocument(t)
	ctx := context.Background()

	if err := testDB.Model(&schemas.Document{}).Where("id = ?", id).Update("signed_hash", "recorded-proof").Error; err != nil {
		t.Fatalf("seed hash: %v", err)
	}
	if _, err := svc.GetFilePathByDocID(ctx, id); err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if _, err := os.Stat(signedPath); err != nil {
		t.Fatalf("signed file was not re-rendered: %v", err)
	}
	if got := storedHash(t, id); got != "recorded-proof" {
		t.Fatalf("stored hash overwritten: %s", got)
	}
}
