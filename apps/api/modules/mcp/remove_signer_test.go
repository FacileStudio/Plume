package mcp_test

import (
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
)

// TestMCPRemoveSigner takes a signer off a draft with their fields, and refuses a stranger
// and a document already sent.
func TestMCPRemoveSigner(t *testing.T) {
	f := newFixture(t)
	doc := f.draft(t, "Offer")
	signerID := doc.Signers[0].ID
	var boxes int64
	must(t, f.db.Model(&schemas.Field{}).Where("signer_id = ?", signerID).Count(&boxes).Error)
	if boxes == 0 {
		t.Fatal("the draft's signer must start with a signature box")
	}
	if text := refused(t, f.connect(t, "stranger"), "plume_remove_signer", map[string]any{"signer_id": signerID}); !strings.Contains(text, "not found") {
		t.Fatalf("a stranger cannot remove the owner's signer, got %q", text)
	}
	var removed struct {
		Status     string `json:"status"`
		DocumentID int64  `json:"document_id"`
	}
	call(t, f.connect(t, "owner"), "plume_remove_signer", map[string]any{"signer_id": signerID}, &removed)
	if removed.Status != "removed" || removed.DocumentID != doc.Document.ID {
		t.Fatalf("want removed from the draft, got %+v", removed)
	}
	var signers, fields int64
	must(t, f.db.Model(&schemas.Signer{}).Where("id = ?", signerID).Count(&signers).Error)
	must(t, f.db.Model(&schemas.Field{}).Where("signer_id = ?", signerID).Count(&fields).Error)
	if signers != 0 || fields != 0 {
		t.Fatalf("the signer and their fields must be gone, got signers=%d fields=%d", signers, fields)
	}
	sent := f.sent(t, "Signed offer")
	if text := refused(t, f.connect(t, "owner"), "plume_remove_signer", map[string]any{"signer_id": sent.Signers[0].ID}); !strings.Contains(text, "draft") {
		t.Fatalf("a sent document keeps its signers, got %q", text)
	}
}
