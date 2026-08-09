package signers

import (
	"context"
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	"github.com/FacileStudio/Plume/apps/api/modules/smtp"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
	"github.com/FacileStudio/Plume/apps/api/schemas"
)

func TestIsSignersTurn(t *testing.T) {
	sig := func(id int64, role, status string, order int) schemas.Signer {
		return schemas.Signer{ID: id, Role: role, Status: status, OrderNum: order}
	}

	cases := []struct {
		name       string
		sequential bool
		current    schemas.Signer
		all        []schemas.Signer
		want       bool
	}{
		{
			name:       "non-sequential always allowed",
			sequential: false,
			current:    sig(2, "signer", "pending", 2),
			all:        []schemas.Signer{sig(1, "signer", "pending", 1), sig(2, "signer", "pending", 2)},
			want:       true,
		},
		{
			name:       "first in order may sign",
			sequential: true,
			current:    sig(1, "signer", "pending", 1),
			all:        []schemas.Signer{sig(1, "signer", "pending", 1), sig(2, "signer", "pending", 2)},
			want:       true,
		},
		{
			name:       "blocked while earlier signer still pending",
			sequential: true,
			current:    sig(2, "signer", "pending", 2),
			all:        []schemas.Signer{sig(1, "signer", "pending", 1), sig(2, "signer", "pending", 2)},
			want:       false,
		},
		{
			name:       "allowed once earlier signer signed",
			sequential: true,
			current:    sig(2, "signer", "pending", 2),
			all:        []schemas.Signer{sig(1, "signer", "signed", 1), sig(2, "signer", "pending", 2)},
			want:       true,
		},
		{
			name:       "earlier viewer does not block",
			sequential: true,
			current:    sig(2, "signer", "pending", 2),
			all:        []schemas.Signer{sig(1, "viewer", "pending", 1), sig(2, "signer", "pending", 2)},
			want:       true,
		},
		{
			name:       "ties at same order may sign in parallel",
			sequential: true,
			current:    sig(2, "signer", "pending", 1),
			all:        []schemas.Signer{sig(1, "signer", "pending", 1), sig(2, "signer", "pending", 1)},
			want:       true,
		},
		{
			name:       "current viewer never gated",
			sequential: true,
			current:    sig(3, "viewer", "pending", 5),
			all:        []schemas.Signer{sig(1, "signer", "pending", 1), sig(3, "viewer", "pending", 5)},
			want:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSignersTurn(tc.sequential, tc.current, tc.all); got != tc.want {
				t.Fatalf("isSignersTurn() = %v, want %v", got, tc.want)
			}
		})
	}
}

// A signing product that stamps a PDF with empty boxes has no product. The asterisk the
// signer page draws on a required field is an affordance; this is the enforcement.
func TestSubmitSignatureRequiresRequiredFields(t *testing.T) {
	db := requireDB(t)
	uploadDir := t.TempDir()

	smtpSvc := smtp.NewService(db)
	webhookSvc := webhooks.NewService(db)
	docSvc := documents.NewService(db, smtpSvc, webhookSvc, "http://localhost:5173", uploadDir)
	service := NewService(db, docSvc, webhookSvc, smtpSvc, "http://localhost:5173")

	doc := schemas.Document{Name: "Contract", Status: "pending", FileName: "contract.pdf", OwnerID: 1}
	if err := db.Create(&doc).Error; err != nil {
		t.Fatalf("create document: %v", err)
	}
	signer := schemas.Signer{DocumentID: doc.ID, Name: "Alice", Email: "alice@example.com", Role: "signer", Status: "pending", Token: "tok-1"}
	if err := db.Create(&signer).Error; err != nil {
		t.Fatalf("create signer: %v", err)
	}

	fullName := schemas.Field{DocumentID: doc.ID, SignerID: signer.ID, FieldType: "text", Required: true, Label: "Full name"}
	consent := schemas.Field{DocumentID: doc.ID, SignerID: signer.ID, FieldType: "checkbox", Required: true, Label: "I agree"}
	initials := schemas.Field{DocumentID: doc.ID, SignerID: signer.ID, FieldType: "text", Required: false, Label: "Initials"}
	for _, field := range []*schemas.Field{&fullName, &consent, &initials} {
		if err := db.Create(field).Error; err != nil {
			t.Fatalf("create field: %v", err)
		}
	}

	rejections := []struct {
		name    string
		fields  []FieldValue
		missing string
	}{
		{
			name:    "nothing submitted at all",
			fields:  nil,
			missing: "Full name, I agree",
		},
		{
			name:    "whitespace is not a text value",
			fields:  []FieldValue{{FieldID: fullName.ID, Value: "   "}, {FieldID: consent.ID, Value: "true"}},
			missing: "Full name",
		},
		{
			name:    "an unticked checkbox is not consent",
			fields:  []FieldValue{{FieldID: fullName.ID, Value: "Alice Martin"}, {FieldID: consent.ID, Value: "false"}},
			missing: "I agree",
		},
	}

	for _, tc := range rejections {
		t.Run(tc.name, func(t *testing.T) {
			err := service.SubmitSignature(context.Background(), "tok-1", &SubmitSignatureRequest{Fields: tc.fields}, "127.0.0.1", "go-test")
			if err == nil {
				t.Fatal("SubmitSignature() accepted a blank required field")
			}
			if !strings.Contains(err.Error(), tc.missing) {
				t.Fatalf("SubmitSignature() error = %q, want it to name %q", err.Error(), tc.missing)
			}
			var after schemas.Signer
			if loadErr := db.Where("id = ?", signer.ID).First(&after).Error; loadErr != nil {
				t.Fatalf("reload signer: %v", loadErr)
			}
			if after.Status != "pending" {
				t.Fatalf("signer status = %q after a rejected submission, want %q", after.Status, "pending")
			}
		})
	}

	filled := []FieldValue{{FieldID: fullName.ID, Value: "Alice Martin"}, {FieldID: consent.ID, Value: "true"}}
	if err := service.SubmitSignature(context.Background(), "tok-1", &SubmitSignatureRequest{Fields: filled}, "127.0.0.1", "go-test"); err != nil {
		t.Fatalf("SubmitSignature() with every required field filled = %v, want nil", err)
	}

	var signed schemas.Signer
	if err := db.Where("id = ?", signer.ID).First(&signed).Error; err != nil {
		t.Fatalf("reload signer: %v", err)
	}
	if signed.Status != "signed" {
		t.Fatalf("signer status = %q, want %q", signed.Status, "signed")
	}
}
