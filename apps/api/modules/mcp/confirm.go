package mcp

import (
	"context"
	"strconv"
	"time"

	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	"github.com/FacileStudio/Plume/apps/api/modules/signers"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// The sensitive actions: sending a draft emails its signers, a reminder emails one
// signer again, deleting a document removes it and its file for good, and deleting a
// client or a draft's field removes it for good.
const (
	kindSend         = "send_document"
	kindRemind       = "remind_signer"
	kindDelete       = "delete_document"
	kindDeleteClient = "delete_client"
	kindDeleteField  = "delete_field"
)

var errTarget = errors.NotFound("this action names no document, signer, client or field")

// kinds are the sensitive actions. Each fingerprint covers the target as it was when the
// person was asked, so an edit in between closes the action instead of emailing or
// deleting something the person did not see.
func (t *toolset) kinds() map[string]mcpkit.Kind {
	return map[string]mcpkit.Kind{
		kindSend:         {Tool: toolSend, Fingerprint: t.sendFingerprint, Run: t.runSend},
		kindRemind:       {Tool: toolRemind, Fingerprint: t.remindFingerprint, Run: t.runRemind},
		kindDelete:       {Tool: toolDelete, Fingerprint: t.deleteFingerprint, Run: t.runDelete},
		kindDeleteClient: {Tool: toolDeleteClient, Fingerprint: t.deleteClientFingerprint, Run: t.runDeleteClient},
		kindDeleteField:  {Tool: toolDeleteField, Fingerprint: t.deleteFieldFingerprint, Run: t.runDeleteField},
	}
}

// refOf parses an action's user and target id.
func refOf(a mcpkit.Action) (int64, int64, error) {
	return mcpkit.TargetIDs(a, errTarget)
}

// sendFingerprint digests the draft and everyone it would email, through the reach of
// POST /documents/{id}/send.
func (t *toolset) sendFingerprint(ctx context.Context, a mcpkit.Action) (string, error) {
	userID, docID, err := refOf(a)
	if err != nil {
		return "", err
	}
	doc, err := t.Documents.Get(ctx, id(userID), id(docID))
	if err != nil {
		return "", err
	}
	list, err := t.Signers.ListSigners(ctx, id(userID), id(docID))
	if err != nil {
		return "", err
	}
	fields := documentFields(doc)
	for _, s := range list {
		fields = append(fields, id(s.ID), s.Email, s.Role, strconv.Itoa(s.OrderNum))
	}
	return mcpkit.Digest(fields...), nil
}

// deleteFingerprint digests the document, through the reach of DELETE /documents/{id}.
func (t *toolset) deleteFingerprint(ctx context.Context, a mcpkit.Action) (string, error) {
	userID, docID, err := refOf(a)
	if err != nil {
		return "", err
	}
	doc, err := t.Documents.Get(ctx, id(userID), id(docID))
	if err != nil {
		return "", err
	}
	return mcpkit.Digest(documentFields(doc)...), nil
}

// remindFingerprint digests the signer, through the reach of POST /signers/{id}/remind:
// the document's owner only.
func (t *toolset) remindFingerprint(ctx context.Context, a mcpkit.Action) (string, error) {
	userID, signerID, err := refOf(a)
	if err != nil {
		return "", err
	}
	signer, err := t.Signers.OwnedSigner(ctx, userID, signerID)
	if err != nil {
		return "", err
	}
	return digestSigner(signer), nil
}

// documentFields are the parts of a document a confirmation covers.
func documentFields(d *documents.DocumentResponse) []string {
	return []string{id(d.ID), d.Name, d.Status, strconv.FormatBool(d.Sequential), d.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}

// digestSigner fingerprints a signer.
func digestSigner(s *signers.SignerResponse) string {
	reminded := "-"
	if s.LastRemindedAt != nil {
		reminded = s.LastRemindedAt.UTC().Format(time.RFC3339Nano)
	}
	return mcpkit.Digest(id(s.ID), id(s.DocumentID), s.Email, s.Status, reminded)
}
