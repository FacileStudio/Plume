package mcp

import (
	"context"
	"fmt"
	"strconv"

	"github.com/FacileStudio/Plume/apps/api/modules/clients"
	"github.com/FacileStudio/Plume/apps/api/modules/fields"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var errFieldNotDraft = errors.Invalid("fields can only be deleted from draft documents")

var (
	deleteClientAction = sensitive{kindDeleteClient, statusDeleted, "The user confirmed and the client is deleted. Its documents stay, with no client.", "Call plume_delete_client again to ask the user afresh."}
	deleteFieldAction  = sensitive{kindDeleteField, statusDeleted, "The user confirmed and the field is deleted.", "Call plume_delete_field again to ask the user afresh."}
)

// deleteClient is plume_delete_client: the client goes once the user confirms, and the
// documents filed under it lose their client, as DELETE /clients/{id} does.
func (t *toolset) deleteClient(ctx context.Context, req *sdk.CallToolRequest, in clientRef) (*sdk.CallToolResult, actionOutput, error) {
	if mcpkit.Retrying(req) {
		return t.resume(ctx, req, deleteClientAction)
	}
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, actionOutput{}, err
	}
	client, err := t.Clients.Get(ctx, userID, in.ClientID)
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, hintClient)
	}
	text := fmt.Sprintf("Delete the client %q for good? Its documents stay, with no client. This cannot be undone.", client.Name)
	return t.ask(ctx, req, question{kindDeleteClient, in.ClientID, text, actionOutput{ClientID: in.ClientID}})
}

// deleteField is plume_delete_field: a field leaves its draft once the user confirms, as
// DELETE /documents/{id}/fields/{fieldId} does.
func (t *toolset) deleteField(ctx context.Context, req *sdk.CallToolRequest, in fieldRef) (*sdk.CallToolResult, actionOutput, error) {
	if mcpkit.Retrying(req) {
		return t.resume(ctx, req, deleteFieldAction)
	}
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, actionOutput{}, err
	}
	field, err := t.Fields.Find(ctx, id(userID), in.FieldID)
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, hintField)
	}
	doc, err := t.Documents.Get(ctx, id(userID), id(field.DocumentID))
	if err != nil {
		return nil, actionOutput{}, mcpkit.Failure(err, hintDocument)
	}
	if doc.Status != "draft" {
		return nil, actionOutput{}, mcpkit.Failure(errFieldNotDraft, "")
	}
	text := fmt.Sprintf("Delete the %s field on page %d of the draft %q?", field.FieldType, field.Page, doc.Name)
	return t.ask(ctx, req, question{kindDeleteField, in.FieldID, text, actionOutput{DocumentID: field.DocumentID, FieldID: in.FieldID}})
}

// deleteClientFingerprint digests the client, through the reach of DELETE /clients/{id}:
// the owner only.
func (t *toolset) deleteClientFingerprint(ctx context.Context, a mcpkit.Action) (string, error) {
	userID, clientID, err := refOf(a)
	if err != nil {
		return "", err
	}
	client, err := t.Clients.Get(ctx, userID, clientID)
	if err != nil {
		return "", err
	}
	return digestClient(client), nil
}

// deleteFieldFingerprint digests the field and its document's state, through the reach
// of the document routes.
func (t *toolset) deleteFieldFingerprint(ctx context.Context, a mcpkit.Action) (string, error) {
	userID, fieldID, err := refOf(a)
	if err != nil {
		return "", err
	}
	field, err := t.Fields.Find(ctx, id(userID), fieldID)
	if err != nil {
		return "", err
	}
	doc, err := t.Documents.Get(ctx, id(userID), id(field.DocumentID))
	if err != nil {
		return "", err
	}
	return mcpkit.Digest(append(documentFields(doc), digestField(field)...)...), nil
}

// runDeleteClient deletes the client and unlinks its documents.
func (t *toolset) runDeleteClient(ctx context.Context, a mcpkit.Action) error {
	userID, clientID, err := refOf(a)
	if err != nil {
		return err
	}
	return t.Clients.Delete(ctx, userID, clientID)
}

// runDeleteField deletes the field from its draft.
func (t *toolset) runDeleteField(ctx context.Context, a mcpkit.Action) error {
	userID, fieldID, err := refOf(a)
	if err != nil {
		return err
	}
	field, err := t.Fields.Find(ctx, id(userID), fieldID)
	if err != nil {
		return err
	}
	return t.Fields.Delete(ctx, id(userID), id(field.DocumentID), id(fieldID))
}

// digestClient fingerprints a client.
func digestClient(c *clients.ClientResponse) string {
	return mcpkit.Digest(id(c.ID), c.Name, c.Email, c.Company, c.Phone, c.Notes, c.UpdatedAt)
}

// digestField lists the parts of a field a confirmation covers.
func digestField(f *fields.FieldResponse) []string {
	return []string{
		id(f.ID), id(f.SignerID), f.FieldType, strconv.Itoa(f.Page), strconv.FormatBool(f.Required), f.Label,
		strconv.FormatFloat(f.X, 'g', -1, 64), strconv.FormatFloat(f.Y, 'g', -1, 64),
		strconv.FormatFloat(f.Width, 'g', -1, 64), strconv.FormatFloat(f.Height, 'g', -1, 64),
	}
}
