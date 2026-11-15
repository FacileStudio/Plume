package documents

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/internal/hashing"
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

func TestConcurrentFirstRendersOfACompletedDocumentAgree(t *testing.T) {
	svc, id, signedPath := completedDocument(t)
	ctx := context.Background()

	const callers = 10
	var wg sync.WaitGroup
	start := make(chan struct{})
	failures := make(chan error, callers)
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := svc.GetFilePathByDocID(ctx, id)
			failures <- err
		}()
	}
	close(start)
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}
	}

	onDisk, err := hashing.SHA256File(signedPath)
	if err != nil {
		t.Fatalf("hash signed file: %v", err)
	}
	if got := storedHash(t, id); got != onDisk {
		t.Fatalf("signed_hash %s does not match the file on disk %s", got, onDisk)
	}
	leftovers, _ := filepath.Glob(filepath.Join(filepath.Dir(signedPath), "*.tmp"))
	if len(leftovers) != 0 {
		t.Fatalf("temporary renders left behind: %v", leftovers)
	}
}

func TestPendingRenderNeverWritesTheSignedPath(t *testing.T) {
	svc, id, signedPath := completedDocument(t)
	if err := testDB.Model(&schemas.Document{}).Where("id = ?", id).Update("status", "pending").Error; err != nil {
		t.Fatalf("set pending: %v", err)
	}
	path, err := svc.GetFilePathByDocID(context.Background(), id)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if path == signedPath {
		t.Fatal("a pending document was served from the signed path")
	}
	if _, err := os.Stat(signedPath); !os.IsNotExist(err) {
		t.Fatalf("pending render created the signed file: %v", err)
	}
	if got := storedHash(t, id); got != "" {
		t.Fatalf("pending render recorded a hash: %s", got)
	}
}

func TestSignatureInvalidationLeavesACompletedFileAlone(t *testing.T) {
	svc, id, signedPath := completedDocument(t)
	ctx := context.Background()
	if _, err := svc.GetFilePathByDocID(ctx, id); err != nil {
		t.Fatalf("fetch: %v", err)
	}
	before, _ := os.ReadFile(signedPath)
	hash := storedHash(t, id)

	svc.InvalidateGeneratedFiles(ctx, id)

	after, err := os.ReadFile(signedPath)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("signed file changed or vanished: %v", err)
	}
	if got := storedHash(t, id); got != hash || hash == "" {
		t.Fatalf("signed_hash = %q, want %q", got, hash)
	}
}

func TestDroppingTheLegacySignedFileHonoursTheHash(t *testing.T) {
	svc, id, signedPath := completedDocument(t)
	if err := os.WriteFile(signedPath, []byte("stale"), 0o600); err != nil {
		t.Fatalf("write legacy: %v", err)
	}
	if err := testDB.Model(&schemas.Document{}).Where("id = ?", id).Update("signed_hash", "proof").Error; err != nil {
		t.Fatalf("seed hash: %v", err)
	}
	if err := svc.DropLegacySignedFile(testDB, id); err != nil {
		t.Fatalf("drop: %v", err)
	}
	if _, err := os.Stat(signedPath); err != nil {
		t.Fatal("a file with a recorded hash was dropped")
	}
	if err := testDB.Model(&schemas.Document{}).Where("id = ?", id).Update("signed_hash", "").Error; err != nil {
		t.Fatalf("clear hash: %v", err)
	}
	if err := svc.DropLegacySignedFile(testDB, id); err != nil {
		t.Fatalf("drop: %v", err)
	}
	if _, err := os.Stat(signedPath); !os.IsNotExist(err) {
		t.Fatal("the legacy file was not dropped")
	}
}
