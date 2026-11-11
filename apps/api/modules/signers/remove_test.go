package signers

import (
	"context"
	"strconv"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	"github.com/FacileStudio/Plume/apps/api/modules/smtp"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
	"github.com/FacileStudio/Plume/apps/api/schemas"

	"gorm.io/gorm"
)

func TestRemoveSignerDeletesItsFields(t *testing.T) {
	db := requireDB(t)
	smtpSvc := smtp.NewService(db)
	webhookSvc := webhooks.NewService(db)
	docSvc := documents.NewService(db, smtpSvc, webhookSvc, "http://localhost:5173", t.TempDir())
	service := NewService(db, docSvc, webhookSvc, smtpSvc, nil, "http://localhost:5173")

	doc, gone, kept := seedTwoSigners(t, db)

	if err := service.RemoveSigner(context.Background(), "1", strconv.FormatInt(gone.ID, 10)); err != nil {
		t.Fatalf("RemoveSigner() = %v", err)
	}

	var orphans, left, signers int64
	db.Model(&schemas.Field{}).Where("signer_id = ?", gone.ID).Count(&orphans)
	db.Model(&schemas.Field{}).Where("signer_id = ?", kept.ID).Count(&left)
	db.Model(&schemas.Signer{}).Where("document_id = ?", doc.ID).Count(&signers)
	if orphans != 0 || left != 1 || signers != 1 {
		t.Fatalf("after removal: fields of the removed signer=%d (want 0), of the other=%d (want 1), signers=%d (want 1)", orphans, left, signers)
	}
}

// seedTwoSigners stores a draft with two signers, two fields for the first and one for
// the second.
func seedTwoSigners(t *testing.T, db *gorm.DB) (schemas.Document, schemas.Signer, schemas.Signer) {
	t.Helper()
	doc := schemas.Document{Name: "Contract", Status: "draft", FileName: "contract.pdf", OwnerID: 1}
	if err := db.Create(&doc).Error; err != nil {
		t.Fatalf("create document: %v", err)
	}
	gone := schemas.Signer{DocumentID: doc.ID, Name: "Alice", Email: "alice@example.com", Role: "signer", Status: "pending", Token: "tok-gone"}
	kept := schemas.Signer{DocumentID: doc.ID, Name: "Bob", Email: "bob@example.com", Role: "signer", Status: "pending", Token: "tok-kept"}
	for _, signer := range []*schemas.Signer{&gone, &kept} {
		if err := db.Create(signer).Error; err != nil {
			t.Fatalf("create signer: %v", err)
		}
	}
	for _, field := range []*schemas.Field{
		{DocumentID: doc.ID, SignerID: gone.ID, FieldType: "signature"},
		{DocumentID: doc.ID, SignerID: gone.ID, FieldType: "text"},
		{DocumentID: doc.ID, SignerID: kept.ID, FieldType: "signature"},
	} {
		if err := db.Create(field).Error; err != nil {
			t.Fatalf("create field: %v", err)
		}
	}
	return doc, gone, kept
}
