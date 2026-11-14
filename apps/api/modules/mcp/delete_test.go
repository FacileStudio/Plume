package mcp_test

import (
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
)

// freshClient is a client of the owner's to delete, counted as run once it is gone.
func (f *fixture) freshClient(tb testing.TB) (map[string]any, func() int) {
	client := schemas.Client{Name: "Doomed client", OwnerID: f.users["owner"]}
	must(tb, f.db.Create(&client).Error)
	return map[string]any{"client_id": client.ID}, func() int { return boolInt(f.gone(&schemas.Client{}, client.ID)) }
}

// freshField is the signature box of a fresh draft, counted as run once it is gone.
func (f *fixture) freshField(tb testing.TB) (map[string]any, func() int) {
	fieldID := f.boxOf(tb, f.draft(tb, "Field me").Document.ID)
	return map[string]any{"field_id": fieldID}, func() int { return boolInt(f.gone(&schemas.Field{}, fieldID)) }
}

// boxOf is the id of the first field on a document.
func (f *fixture) boxOf(tb testing.TB, docID int64) int64 {
	var field schemas.Field
	must(tb, f.db.Where("document_id = ?", docID).First(&field).Error)
	return field.ID
}

// gone reports whether the row with id no longer exists.
func (f *fixture) gone(model any, id int64) bool {
	var count int64
	f.db.Model(model).Where("id = ?", id).Count(&count)
	return count == 0
}

// TestMCPDeletionsStayInReach lists a draft's fields, and refuses a stranger's client or
// field deletion and a field on a sent document before anything is asked.
func TestMCPDeletionsStayInReach(t *testing.T) {
	f := newFixture(t)
	doc := f.draft(t, "Offer")
	var view struct {
		Fields []struct {
			ID   int64  `json:"id"`
			Type string `json:"type"`
		} `json:"fields"`
	}
	call(t, f.connect(t, "owner"), "plume_get_document", map[string]any{"document_id": doc.Document.ID}, &view)
	if len(view.Fields) != 1 || view.Fields[0].ID != f.boxOf(t, doc.Document.ID) || view.Fields[0].Type != "signature" {
		t.Fatalf("the draft's signature box must be listed with its id, got %+v", view)
	}
	stranger := f.connect(t, "stranger")
	if text := refused(t, stranger, "plume_delete_field", map[string]any{"field_id": view.Fields[0].ID}); !strings.Contains(text, "not found") {
		t.Fatalf("a stranger cannot delete the owner's field, got %q", text)
	}
	args, _ := f.freshClient(t)
	if text := refused(t, stranger, "plume_delete_client", args); !strings.Contains(text, "not found") {
		t.Fatalf("a stranger cannot delete the owner's client, got %q", text)
	}
	sent := f.sent(t, "Out")
	if text := refused(t, f.connect(t, "owner"), "plume_delete_field", map[string]any{"field_id": f.boxOf(t, sent.Document.ID)}); !strings.Contains(text, "draft") {
		t.Fatalf("a sent document keeps its fields, got %q", text)
	}
}

// TestMCPFieldLookupHidesWhichIdsExist answers a stranger the same way for a field id
// that does not exist, a field on the owner's own draft and a field on a space draft.
func TestMCPFieldLookupHidesWhichIdsExist(t *testing.T) {
	f := newFixture(t)
	personal := f.boxOf(t, f.draft(t, "Personal").Document.ID)
	shared := f.draft(t, "Shared")
	must(t, f.db.Model(&schemas.Document{}).Where("id = ?", shared.Document.ID).Update("space_id", f.space.ID).Error)
	stranger := f.connect(t, "stranger")
	want := refused(t, stranger, "plume_delete_field", map[string]any{"field_id": 999_999})
	for name, fieldID := range map[string]int64{"personal": personal, "space": f.boxOf(t, shared.Document.ID)} {
		if got := refused(t, stranger, "plume_delete_field", map[string]any{"field_id": fieldID}); got != want {
			t.Fatalf("a %s field must answer like a missing one: got %q, want %q", name, got, want)
		}
	}
}
