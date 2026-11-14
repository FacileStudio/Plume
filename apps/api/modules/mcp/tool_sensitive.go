package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/FacileStudio/Plume/apps/api/modules/signers"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	statusSent     = "sent"
	statusReminded = "reminded"
	statusDeleted  = "deleted"
)

var (
	errNotDraft   = errors.Invalid("only draft documents can be sent")
	errNoSigners  = errors.Invalid("the document has nobody to sign or approve yet. Add a signer with plume_add_signer")
	errNotPending = errors.Invalid("only a signer who has not signed yet, on a document out for signature, can be reminded")
)

// sensitive describes one confirmed tool: its kind, its done status and what the model
// reads once the person answered.
type sensitive struct {
	kind, done, doneMessage, retryHint string
}

var (
	sendAction   = sensitive{kindSend, statusSent, "The user confirmed and Plume emailed the signers.", "Call plume_send_document again to ask the user afresh."}
	remindAction = sensitive{kindRemind, statusReminded, "The user confirmed and Plume emailed the signer again.", "Call plume_remind_signer again to ask the user afresh."}
	deleteAction = sensitive{kindDelete, statusDeleted, "The user confirmed and the document is deleted.", "Call plume_delete_document again to ask the user afresh."}
)

// sendDocument is plume_send_document: the draft goes to its signers once the user
// confirms.
func (t *toolset) sendDocument(ctx context.Context, req *sdk.CallToolRequest, in documentInput) (*sdk.CallToolResult, actionOutput, error) {
	if mcpkit.Retrying(req) {
		return t.resume(ctx, req, sendAction)
	}
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, actionOutput{}, err
	}
	view, err := t.documentView(ctx, userID, in.DocumentID)
	if err != nil {
		return nil, actionOutput{}, err
	}
	text, err := sendQuestion(view)
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, "")
	}
	return t.ask(ctx, req, question{kindSend, in.DocumentID, text, actionOutput{DocumentID: in.DocumentID}})
}

// sendQuestion is what the person is asked before a draft goes out.
func sendQuestion(view documentOutput) (string, error) {
	if view.Document.Status != "draft" {
		return "", errNotDraft
	}
	var invited []string
	for _, s := range view.Signers {
		if s.Role != "viewer" {
			invited = append(invited, s.Email)
		}
	}
	if len(invited) == 0 {
		return "", errNoSigners
	}
	return fmt.Sprintf("Send %q for signature? Plume emails a signing link to %s.", view.Document.Name, strings.Join(invited, ", ")), nil
}

// remindSigner is plume_remind_signer: the signer gets their link again once the user
// confirms.
func (t *toolset) remindSigner(ctx context.Context, req *sdk.CallToolRequest, in signerRef) (*sdk.CallToolResult, actionOutput, error) {
	if mcpkit.Retrying(req) {
		return t.resume(ctx, req, remindAction)
	}
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, actionOutput{}, err
	}
	signer, err := t.Signers.OwnedSigner(ctx, userID, in.SignerID)
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, hintSigner)
	}
	doc, err := t.Documents.Get(ctx, id(userID), id(signer.DocumentID))
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, hintDocument)
	}
	if signer.Status != "pending" || doc.Status != "pending" {
		return nil, actionOutput{}, mcpkit.Failure(errNotPending, "")
	}
	out := actionOutput{DocumentID: signer.DocumentID, SignerID: in.SignerID}
	return t.ask(ctx, req, question{kindRemind, in.SignerID, remindQuestion(signer, doc.Name), out})
}

// remindQuestion is what the person is asked before a signer is emailed again.
func remindQuestion(s *signers.SignerResponse, document string) string {
	return fmt.Sprintf("Email %s (%s) the signing link for %q again?", s.Name, s.Email, document)
}

// deleteDocument is plume_delete_document: the document goes once the user confirms.
func (t *toolset) deleteDocument(ctx context.Context, req *sdk.CallToolRequest, in documentInput) (*sdk.CallToolResult, actionOutput, error) {
	if mcpkit.Retrying(req) {
		return t.resume(ctx, req, deleteAction)
	}
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, actionOutput{}, err
	}
	doc, err := t.Documents.Get(ctx, id(userID), id(in.DocumentID))
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, hintDocument)
	}
	text := fmt.Sprintf("Delete the %s document %q, its signers and its file for good? This cannot be undone.", doc.Status, doc.Name)
	return t.ask(ctx, req, question{kindDelete, in.DocumentID, text, actionOutput{DocumentID: in.DocumentID}})
}

// question is one confirmation to open: its kind, target, text and the result fields
// the model reads while it waits.
type question struct {
	kind   string
	target int64
	text   string
	out    actionOutput
}

// ask opens the pending action and returns either the client's question or the
// pending status with its confirm link.
func (t *toolset) ask(ctx context.Context, req *sdk.CallToolRequest, q question) (*sdk.CallToolResult, actionOutput, error) {
	out := q.out
	result, outcome, err := t.confirm.Request(ctx, req, q.kind, id(q.target), q.text)
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, hintDocument)
	}
	if result != nil {
		return result, actionOutput{}, nil
	}
	out.Status, out.ConfirmURL, out.Message = outcome.Status, outcome.ConfirmURL, outcome.Message
	return nil, out, nil
}

// resume handles the retry that carries the person's answer.
func (t *toolset) resume(ctx context.Context, req *sdk.CallToolRequest, s sensitive) (*sdk.CallToolResult, actionOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, actionOutput{}, err
	}
	outcome, err := t.confirm.Resume(ctx, req, s.kind)
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, s.retryHint)
	}
	if outcome.Status == mcpkit.StatusDeclined {
		return nil, actionOutput{Status: mcpkit.StatusDeclined, Message: outcome.Message}, nil
	}
	_, targetID, _ := refOf(outcome.Action)
	out := doneOutput(s, targetID)
	if s.kind == kindSend {
		if view, err := t.documentView(ctx, userID, targetID); err == nil {
			out.Document = &view.Document
		}
	}
	return nil, out, nil
}

// doneOutput is what the model reads once the action ran on targetID.
func doneOutput(s sensitive, targetID int64) actionOutput {
	out := actionOutput{Status: s.done, Message: s.doneMessage}
	switch s.kind {
	case kindRemind:
		out.SignerID = targetID
	case kindDeleteClient:
		out.ClientID = targetID
	case kindDeleteField:
		out.FieldID = targetID
	default:
		out.DocumentID = targetID
	}
	return out
}
