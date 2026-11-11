package mcp

import (
	"context"

	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// runSend sends the draft to its signers.
func (t *toolset) runSend(ctx context.Context, a mcpkit.Action) error {
	userID, docID, err := refOf(a)
	if err != nil {
		return err
	}
	_, err = t.Documents.Send(ctx, id(userID), id(docID))
	return err
}

// runRemind emails the signer their link again.
func (t *toolset) runRemind(ctx context.Context, a mcpkit.Action) error {
	userID, signerID, err := refOf(a)
	if err != nil {
		return err
	}
	_, err = t.Reminders.RemindSigner(ctx, id(userID), id(signerID))
	return err
}

// runDelete deletes the document and its file.
func (t *toolset) runDelete(ctx context.Context, a mcpkit.Action) error {
	userID, docID, err := refOf(a)
	if err != nil {
		return err
	}
	return t.Documents.Delete(ctx, id(userID), id(docID))
}
