package mcp_test

import (
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/mcp/mcptest"
)

// documentOut is the part of a document result the tests read.
type documentOut struct {
	Status   string `json:"status"`
	Document struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Status   string `json:"status"`
		ClientID *int64 `json:"client_id"`
		URL      string `json:"url"`
	} `json:"document"`
	Signers []struct {
		ID     int64  `json:"id"`
		Email  string `json:"email"`
		Fields int    `json:"fields"`
	} `json:"signers"`
}

// TestMCPConformance runs tronc's conformance kit over the three confirmed tools.
func TestMCPConformance(t *testing.T) {
	f := newFixture(t)
	mcptest.Run(t, mcptest.Config{
		URL:   f.resource,
		Token: f.token,
		Confirm: []mcptest.ConfirmCase{
			{Tool: "plume_send_document", Setup: f.freshDraft},
			{Tool: "plume_remind_signer", Setup: f.freshSigner},
			{Tool: "plume_delete_document", Setup: f.freshDoomed},
		},
	})
}

// freshDraft is a draft to send, counted as run once it is pending.
func (f *fixture) freshDraft(tb testing.TB) (map[string]any, func() int) {
	doc := f.draft(tb, "Send me")
	return map[string]any{"document_id": doc.Document.ID}, func() int { return boolInt(f.status(doc.Document.ID) == "pending") }
}

// freshDoomed is a draft to delete, counted as run once it is gone.
func (f *fixture) freshDoomed(tb testing.TB) (map[string]any, func() int) {
	doc := f.draft(tb, "Delete me")
	return map[string]any{"document_id": doc.Document.ID}, func() int { return boolInt(f.status(doc.Document.ID) == "gone") }
}

// freshSigner is a signer to remind, counted as run once they were reminded today.
func (f *fixture) freshSigner(tb testing.TB) (map[string]any, func() int) {
	signer := f.sent(tb, "Remind me").Signers[0]
	return map[string]any{"signer_id": signer.ID}, func() int {
		var count int64
		f.db.Model(&schemas.Signer{}).Where("id = ? AND last_reminded_at > now() - interval '1 hour'", signer.ID).Count(&count)
		return int(count)
	}
}

// boolInt is 1 for true.
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// TestMCPDraftFlow walks a model's usual path: upload a draft, add a signer, rename it,
// find it, read it.
func TestMCPDraftFlow(t *testing.T) {
	f := newFixture(t)
	session := f.connect(t, "owner")
	var created documentOut
	call(t, session, "plume_create_document", map[string]any{
		"name": "Lease", "pdf_base64": "data:application/pdf;base64," + encoded(t, 2),
		"signers": []map[string]any{{"name": "Ada", "email": "ada@example.test"}},
	}, &created)
	if created.Status != "created" || created.Document.Status != "draft" || len(created.Signers) != 1 || created.Signers[0].Fields != 1 {
		t.Fatalf("want a draft with one signer and one box, got %+v", created)
	}
	if !strings.HasSuffix(created.Document.URL, "/documents/"+itoa(created.Document.ID)) {
		t.Fatalf("the document must link to its page, got %q", created.Document.URL)
	}
	var added struct {
		Signer struct {
			Fields int `json:"fields"`
		} `json:"signer"`
	}
	call(t, session, "plume_add_signer", map[string]any{"document_id": created.Document.ID, "name": "Bob", "email": "bob@example.test"}, &added)
	if added.Signer.Fields != 1 {
		t.Fatalf("a new signer must get a signature box, got %+v", added)
	}
	var updated documentOut
	call(t, session, "plume_update_document", map[string]any{"document_id": created.Document.ID, "name": "Lease 2026", "sequential": true}, &updated)
	if updated.Document.Name != "Lease 2026" || len(updated.Signers) != 2 {
		t.Fatalf("the rename must keep both signers, got %+v", updated)
	}
	checkListed(t, f, created.Document.ID)
}

// checkListed finds the draft by name and refuses a bad status and an empty edit.
func checkListed(t *testing.T, f *fixture, docID int64) {
	t.Helper()
	session := f.connect(t, "owner")
	var list struct {
		Documents []struct{ ID int64 }    `json:"documents"`
		Counts    struct{ Total int64 }   `json:"counts"`
		Spaces    []struct{ Name string } `json:"spaces"`
	}
	call(t, session, "plume_list_documents", map[string]any{"query": "lease 20"}, &list)
	if len(list.Documents) != 1 || list.Documents[0].ID != docID || list.Counts.Total != 1 || len(list.Spaces) != 1 {
		t.Fatalf("the draft must be found by name, with the counts and the space, got %+v", list)
	}
	if text := refused(t, session, "plume_list_documents", map[string]any{"status": "lost"}); !strings.Contains(text, "status must be") {
		t.Fatalf("an unknown status must say which exist, got %q", text)
	}
	if text := refused(t, session, "plume_update_document", map[string]any{"document_id": docID}); !strings.Contains(text, "at least one") {
		t.Fatalf("an empty edit must say what to pass, got %q", text)
	}
}

// TestMCPReachAndSecrets checks a stranger cannot read the owner's document, and that a
// sent document's private signing links never reach the model.
func TestMCPReachAndSecrets(t *testing.T) {
	f := newFixture(t)
	doc := f.sent(t, "Contract")
	var token string
	must(t, f.db.Model(&schemas.Signer{}).Where("id = ?", doc.Signers[0].ID).Pluck("token", &token).Error)
	if token == "" {
		t.Fatal("a sent document's signer must hold a token")
	}
	result := mcptest.Call(t, f.connect(t, "owner"), "plume_get_document", map[string]any{"document_id": doc.Document.ID})
	if result.IsError || strings.Contains(mcptest.Text(result), token) {
		t.Fatalf("the owner reads the document without the signing link, got %s", mcptest.Text(result))
	}
	text := refused(t, f.connect(t, "stranger"), "plume_get_document", map[string]any{"document_id": doc.Document.ID})
	if !strings.Contains(text, "not found") || !strings.Contains(text, "plume_list_documents") {
		t.Fatalf("a stranger must get not found with a hint, got %q", text)
	}
	text = refused(t, f.connect(t, "stranger"), "plume_remind_signer", map[string]any{"signer_id": doc.Signers[0].ID})
	if !strings.Contains(text, "not found") {
		t.Fatalf("a stranger cannot remind the owner's signer, got %q", text)
	}
	if text := refused(t, f.connect(t, "owner"), "plume_send_document", map[string]any{"document_id": doc.Document.ID}); !strings.Contains(text, "draft") {
		t.Fatalf("a sent document cannot be sent again, got %q", text)
	}
}

// TestMCPClients creates a client, changes one field, finds it and files a draft under it.
func TestMCPClients(t *testing.T) {
	f := newFixture(t)
	session := f.connect(t, "owner")
	var saved struct {
		Status string `json:"status"`
		Client struct {
			ID      int64  `json:"id"`
			Email   string `json:"email"`
			Company string `json:"company"`
		} `json:"client"`
	}
	call(t, session, "plume_save_client", map[string]any{"name": "Acme", "email": "hi@acme.test", "company": "Acme SAS"}, &saved)
	call(t, session, "plume_save_client", map[string]any{"client_id": saved.Client.ID, "company": "Acme Group"}, &saved)
	if saved.Status != "updated" || saved.Client.Email != "hi@acme.test" || saved.Client.Company != "Acme Group" {
		t.Fatalf("a partial edit must keep the other fields, got %+v", saved)
	}
	var list struct {
		Clients []struct{ ID int64 } `json:"clients"`
	}
	call(t, session, "plume_list_clients", map[string]any{"query": "group"}, &list)
	if len(list.Clients) != 1 || list.Clients[0].ID != saved.Client.ID {
		t.Fatalf("the client must be found by company, got %+v", list)
	}
	checkFiled(t, f, saved.Client.ID)
}

// checkFiled files a draft under clientID and refuses the edit to a stranger.
func checkFiled(t *testing.T, f *fixture, clientID int64) {
	t.Helper()
	doc := f.draft(t, "Quote")
	var updated documentOut
	call(t, f.connect(t, "owner"), "plume_update_document", map[string]any{"document_id": doc.Document.ID, "client_id": clientID}, &updated)
	if updated.Document.ClientID == nil || *updated.Document.ClientID != clientID {
		t.Fatalf("the draft must be filed under the client, got %+v", updated)
	}
	if text := refused(t, f.connect(t, "stranger"), "plume_update_document", map[string]any{"document_id": doc.Document.ID, "client_id": 0}); !strings.Contains(text, "not found") {
		t.Fatalf("a stranger cannot edit the owner's draft, got %q", text)
	}
}
