package mcp

import (
	"context"
	stderrors "errors"

	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	documentStates = map[string]bool{"": true, "draft": true, "pending": true, "completed": true, "declined": true}
	errStatus      = stderrors.New("status must be draft, pending, completed or declined")
)

// listDocuments is plume_list_documents: one scope's documents, newest first, with its
// counts and the spaces the person can switch to.
func (t *toolset) listDocuments(ctx context.Context, req *sdk.CallToolRequest, in listInput) (*sdk.CallToolResult, listOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, listOutput{}, err
	}
	if !documentStates[in.Status] {
		return nil, listOutput{}, errStatus
	}
	space, client := optional(in.SpaceID), optional(in.ClientID)
	all, err := t.Documents.List(ctx, id(userID), in.Status, space, client)
	if err != nil {
		return nil, listOutput{}, mcpkit.Failure(err, hintSpace)
	}
	out, err := t.scopeOf(ctx, userID, space)
	if err != nil {
		return nil, listOutput{}, err
	}
	page, next, err := mcpkit.Page(matching(all, in.Query), in.Cursor, documentPage)
	if err != nil {
		return nil, listOutput{}, err
	}
	out.NextCursor = next
	for i := range page {
		out.Documents = append(out.Documents, t.document(&page[i]))
	}
	return nil, out, nil
}

// scopeOf is the counts of one scope and the person's spaces.
func (t *toolset) scopeOf(ctx context.Context, userID int64, space string) (listOutput, error) {
	stats, err := t.Documents.Stats(ctx, id(userID), space)
	if err != nil {
		return listOutput{}, mcpkit.Failure(err, hintSpace)
	}
	list, err := t.Spaces.List(ctx, userID)
	if err != nil {
		return listOutput{}, mcpkit.Failure(err, "")
	}
	return listOutput{
		Documents: []documentOut{}, Spaces: spacesOf(list),
		Counts: countsOut{Total: stats.Total, Pending: stats.Pending, Completed: stats.Completed},
	}, nil
}

// matching keeps the documents whose name holds query.
func matching(all []documents.DocumentResponse, query string) []documents.DocumentResponse {
	kept := all[:0:0]
	for _, d := range all {
		if contains(query, d.Name, d.FileName) {
			kept = append(kept, d)
		}
	}
	return kept
}

// optional renders an optional id filter: empty when it is not set.
func optional(n int64) string {
	if n <= 0 {
		return ""
	}
	return id(n)
}

// getDocument is plume_get_document.
func (t *toolset) getDocument(ctx context.Context, req *sdk.CallToolRequest, in documentInput) (*sdk.CallToolResult, documentOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, documentOutput{}, err
	}
	out, err := t.documentView(ctx, userID, in.DocumentID)
	return nil, out, err
}

// documentView reads a document the person can reach, with its signers and client.
func (t *toolset) documentView(ctx context.Context, userID, docID int64) (documentOutput, error) {
	owner, doc := id(userID), id(docID)
	record, err := t.Documents.Get(ctx, owner, doc)
	if err != nil {
		return documentOutput{}, mcpkit.Failure(err, hintDocument)
	}
	list, err := t.Signers.ListSigners(ctx, owner, doc)
	if err != nil {
		return documentOutput{}, mcpkit.Failure(err, hintDocument)
	}
	placed, err := t.Fields.List(ctx, owner, doc)
	if err != nil {
		return documentOutput{}, mcpkit.Failure(err, hintDocument)
	}
	count := fieldsPerSigner(placed)
	out := documentOutput{Document: t.document(record), Signers: make([]signerOut, len(list)), Client: t.clientFor(ctx, userID, record.ClientID)}
	for i, s := range list {
		out.Signers[i] = signerOf(s, count)
	}
	out.Fields = fieldsOf(placed)
	return out, nil
}

// clientFor is the person's client a document is filed under, or nil when it has none
// or the client belongs to someone else in a shared space.
func (t *toolset) clientFor(ctx context.Context, userID int64, clientID *int64) *clientOut {
	if clientID == nil {
		return nil
	}
	client, err := t.Clients.Get(ctx, userID, *clientID)
	if err != nil {
		return nil
	}
	out := clientOf(client)
	return &out
}
