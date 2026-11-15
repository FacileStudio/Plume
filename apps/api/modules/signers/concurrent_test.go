package signers

import (
	"context"
	stderrors "errors"
	"fmt"
	"sync"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	"github.com/FacileStudio/Plume/apps/api/modules/smtp"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

func seedRace(t *testing.T) (*Service, schemas.Document, schemas.Field) {
	t.Helper()
	db := requireDB(t)
	smtpSvc := smtp.NewService(db)
	webhookSvc := webhooks.NewService(db)
	docSvc := documents.NewService(db, smtpSvc, webhookSvc, "http://localhost:5173", t.TempDir())
	service := NewService(db, docSvc, webhookSvc, smtpSvc, nil, "http://localhost:5173")

	doc := schemas.Document{Name: "Race", Status: "pending", FileName: "race.pdf", OwnerID: 1}
	signer := schemas.Signer{Name: "Alice", Email: "alice@example.com", Role: "signer", Status: "pending", Token: "tok-race"}
	field := schemas.Field{FieldType: "text", Required: true, Label: "Name"}
	if err := db.Create(&doc).Error; err != nil {
		t.Fatalf("create document: %v", err)
	}
	signer.DocumentID = doc.ID
	if err := db.Create(&signer).Error; err != nil {
		t.Fatalf("create signer: %v", err)
	}
	field.DocumentID, field.SignerID = doc.ID, signer.ID
	if err := db.Create(&field).Error; err != nil {
		t.Fatalf("create field: %v", err)
	}
	return service, doc, field
}

func submitTogether(service *Service, fieldID int64, attempts int) []error {
	errs := make([]error, attempts)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			req := &SubmitSignatureRequest{Fields: []FieldValue{{FieldID: fieldID, Value: fmt.Sprintf("value-%d", i)}}}
			errs[i] = service.SubmitSignature(context.Background(), "tok-race", req, "127.0.0.1", "go-test")
		}(i)
	}
	close(start)
	wg.Wait()
	return errs
}

func rejectionCause(err error) error {
	var typed *errors.Error
	if stderrors.As(err, &typed) {
		return typed.Cause
	}
	return err
}

func soleWinner(t *testing.T, errs []error) int {
	t.Helper()
	winner := -1
	for i, err := range errs {
		if err == nil {
			if winner != -1 {
				t.Fatalf("two submissions succeeded: %d and %d", winner, i)
			}
			winner = i
		} else if msg := err.Error(); msg != errAlreadyActed && msg != "document is not available for signing" {
			t.Fatalf("loser %d got %q (cause %v), want a rejection", i, msg, rejectionCause(err))
		}
	}
	if winner == -1 {
		t.Fatal("no submission succeeded")
	}
	return winner
}

func TestConcurrentSubmissionsWithOneTokenYieldOneWinner(t *testing.T) {
	service, doc, field := seedRace(t)

	winner := soleWinner(t, submitTogether(service, field.ID, 6))

	var stored schemas.Field
	if err := testDB.First(&stored, field.ID).Error; err != nil {
		t.Fatalf("reload field: %v", err)
	}
	if want := fmt.Sprintf("value-%d", winner); stored.Value != want {
		t.Fatalf("stored value = %q, want the winner's %q", stored.Value, want)
	}
	var after schemas.Document
	if err := testDB.First(&after, doc.ID).Error; err != nil || after.Status != "completed" {
		t.Fatalf("document = %+v, err %v, want completed", after, err)
	}
}
