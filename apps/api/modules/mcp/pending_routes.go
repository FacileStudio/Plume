package mcp

import (
	"context"
	"net/http"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

const webActor = "plume-web"

// pendingUser is the signed-in person of a confirm page request. Only the person's own
// session passes: a delegated token must never confirm what an agent asked for.
func pendingUser(r *http.Request) (string, error) {
	authenticated, ok := porte.From(r.Context())
	if !ok {
		return "", errors.Unauthorized("missing auth")
	}
	if authenticated.SessionID == 0 || authenticated.Actor != "" {
		return "", mcpkit.ErrDelegated
	}
	return id(authenticated.UserID), nil
}

// describe adds the targeted document or signer, as it is now, to a confirm page item.
func (t *toolset) describe(ctx context.Context, item mcpkit.PendingItem, a mcpkit.Action) PendingItem {
	out := PendingItem{PendingItem: item}
	userID, targetID, err := refOf(a)
	if err != nil {
		return out
	}
	docID := targetID
	if a.Kind == kindRemind {
		signer, err := t.Signers.OwnedSigner(ctx, userID, targetID)
		if err != nil {
			return out
		}
		out.Signer = &PendingSigner{ID: signer.ID, Name: signer.Name, Email: signer.Email}
		docID = signer.DocumentID
	}
	out.Document = t.pendingDocument(ctx, userID, docID)
	return out
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
