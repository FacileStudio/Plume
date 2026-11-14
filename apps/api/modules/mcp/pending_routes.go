package mcp

import (
	"context"
	"net/http"

	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

const webActor = "plume-web"

// pendingUser is the signed-in person of a confirm page request. Only the person's own
// browser session passes (porte mcpauth.RequireOwnSession): a bearer, an API token, a
// CLI session or a delegated token must never confirm what an agent asked for.
func pendingUser(r *http.Request) (string, error) {
	identity, err := mcpauth.RequireOwnSession(r)
	if err != nil {
		return "", err
	}
	return id(identity.UserID), nil
}

// describe adds the targeted document, signer, client or field, as it is now, to a
// confirm page item.
func (t *toolset) describe(ctx context.Context, item mcpkit.PendingItem, a mcpkit.Action) PendingItem {
	out := PendingItem{PendingItem: item}
	userID, targetID, err := refOf(a)
	if err != nil {
		return out
	}
	switch a.Kind {
	case kindRemind:
		t.describeSigner(ctx, &out, userID, targetID)
	case kindDeleteClient:
		if client, err := t.Clients.Get(ctx, userID, targetID); err == nil {
			out.Client = &PendingClient{ID: client.ID, Name: client.Name, Email: client.Email, Company: client.Company}
		}
	case kindDeleteField:
		t.describeField(ctx, &out, userID, targetID)
	default:
		out.Document = t.pendingDocument(ctx, userID, targetID)
	}
	return out
}

// describeSigner adds the signer a reminder targets and their document.
func (t *toolset) describeSigner(ctx context.Context, out *PendingItem, userID, signerID int64) {
	signer, err := t.Signers.OwnedSigner(ctx, userID, signerID)
	if err != nil {
		return
	}
	out.Signer = &PendingSigner{ID: signer.ID, Name: signer.Name, Email: signer.Email}
	out.Document = t.pendingDocument(ctx, userID, signer.DocumentID)
}

// describeField adds the field a deletion targets and its draft.
func (t *toolset) describeField(ctx context.Context, out *PendingItem, userID, fieldID int64) {
	field, err := t.Fields.Find(ctx, id(userID), fieldID)
	if err != nil {
		return
	}
	out.Field = &PendingField{ID: field.ID, Type: field.FieldType, Page: field.Page, Label: field.Label}
	out.Document = t.pendingDocument(ctx, userID, field.DocumentID)
}

// pendingDocument is the document an action targets, or nil once it is gone.
func (t *toolset) pendingDocument(ctx context.Context, userID, docID int64) *PendingDocument {
	view, err := t.documentView(ctx, userID, docID)
	if err != nil {
		return nil
	}
	out := &PendingDocument{ID: docID, Name: view.Document.Name, Status: view.Document.Status, Signers: []string{}}
	for _, s := range view.Signers {
		out.Signers = append(out.Signers, s.Name+" <"+s.Email+">")
	}
	return out
}

// done is the confirm page's answer once an action ran, with the document's page when it
// still exists.
func (t *toolset) done(_ context.Context, o mcpkit.Outcome) mcpkit.PendingResult {
	_, targetID, _ := refOf(o.Action)
	switch o.Action.Kind {
	case kindSend:
		return mcpkit.PendingResult{Status: statusSent, URL: t.documentURL(targetID)}
	case kindRemind:
		return mcpkit.PendingResult{Status: statusReminded}
	}
	return mcpkit.PendingResult{Status: statusDeleted}
}
