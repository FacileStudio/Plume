package mcp

import (
	"context"
	stderrors "errors"
	"strings"

	"github.com/FacileStudio/Plume/apps/api/internal/pdfutil"
	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	"github.com/FacileStudio/Plume/apps/api/modules/fields"
	"github.com/FacileStudio/Plume/apps/api/modules/signers"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	statusAdded   = "added"
	statusRemoved = "removed"
)

var errNoChange = stderrors.New("pass at least one of name, sequential or client_id to change")

// updateDocument is plume_update_document: a draft's title, order or client.
func (t *toolset) updateDocument(ctx context.Context, req *sdk.CallToolRequest, in updateInput) (*sdk.CallToolResult, documentOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, documentOutput{}, err
	}
	if strings.TrimSpace(in.Name) == "" && in.Sequential == nil && in.ClientID == nil {
		return nil, documentOutput{}, errNoChange
	}
	if in.ClientID != nil && *in.ClientID > 0 {
		if _, err := t.Clients.Get(ctx, userID, *in.ClientID); err != nil {
			return nil, documentOutput{}, mcpkit.Failure(err, hintClient)
		}
	}
	change := &documents.UpdateRequest{Name: strings.TrimSpace(in.Name), Sequential: in.Sequential, ClientID: in.ClientID}
	if _, err := t.Documents.Update(ctx, id(userID), id(in.DocumentID), change); err != nil {
		return nil, documentOutput{}, mcpkit.Failure(err, hintDocument)
	}
	out, err := t.documentView(ctx, userID, in.DocumentID)
	out.Status = statusUpdated
	return nil, out, err
}

// addSigner is plume_add_signer: a person on a draft, with a signature box on the last
// page unless they only receive a copy or the caller asks for none.
func (t *toolset) addSigner(ctx context.Context, req *sdk.CallToolRequest, in addSignerInput) (*sdk.CallToolResult, signerOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, signerOutput{}, err
	}
	owner, doc := id(userID), id(in.DocumentID)
	added, err := t.Signers.AddSigner(ctx, owner, doc, &signers.AddSignerRequest{
		Name: strings.TrimSpace(in.Name), Email: strings.TrimSpace(in.Email), Role: in.Role, Order: in.Order,
	})
	if err != nil {
		return nil, signerOutput{}, mcpkit.Failure(err, hintDocument)
	}
	placed := 0
	if !in.NoSignatureBox && added.Role != "viewer" {
		if err := t.placeSignatureBox(ctx, owner, in.DocumentID, added.ID); err != nil {
			return nil, signerOutput{}, mcpkit.Failure(err, hintDocument)
		}
		placed = 1
	}
	out := signerOf(*added, map[int64]int{added.ID: placed})
	return nil, signerOutput{Status: statusAdded, DocumentID: in.DocumentID, Signer: &out}, nil
}

// removeSigner is plume_remove_signer: a person off a draft, with the fields placed for
// them, through the same service call as DELETE /signers/{id}.
func (t *toolset) removeSigner(ctx context.Context, req *sdk.CallToolRequest, in signerRef) (*sdk.CallToolResult, signerOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, signerOutput{}, err
	}
	signer, err := t.Signers.OwnedSigner(ctx, userID, in.SignerID)
	if err != nil {
		return nil, signerOutput{}, mcpkit.Failure(err, hintSigner)
	}
	if err := t.Signers.RemoveSigner(ctx, id(userID), id(in.SignerID)); err != nil {
		return nil, signerOutput{}, mcpkit.Failure(err, hintSigner)
	}
	return nil, signerOutput{Status: statusRemoved, DocumentID: signer.DocumentID}, nil
}

// placeSignatureBox adds the default signature box for signerID, below the boxes already
// on the last page, as a signature request without fields does.
func (t *toolset) placeSignatureBox(ctx context.Context, owner string, docID, signerID int64) error {
	path, err := t.Documents.GetFilePath(ctx, owner, id(docID))
	if err != nil {
		return err
	}
	pages, err := pdfutil.PageCount(path)
	if err != nil || pages < 1 {
		return mcpkit.ErrUnavailable
	}
	existing, err := t.Fields.List(ctx, owner, id(docID))
	if err != nil {
		return err
	}
	box := documents.DefaultSignatureField(docID, signerID, pages, boxesOn(existing, pages))
	_, err = t.Fields.Create(ctx, owner, id(docID), &fields.CreateFieldRequest{
		SignerID: signerID, FieldType: box.FieldType, Page: box.Page, X: box.X, Y: box.Y,
		Width: box.Width, Height: box.Height, Required: box.Required, Label: box.Label,
	})
	return err
}

// boxesOn counts the signature boxes already on page.
func boxesOn(list []fields.FieldResponse, page int) int {
	n := 0
	for _, f := range list {
		if f.Page == page && f.FieldType == "signature" {
			n++
		}
	}
	return n
}
