package mcp

import (
	"bytes"
	"context"
	"encoding/base64"
	stderrors "errors"
	"strings"

	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	statusCreated = "created"
	statusUpdated = "updated"
	maxPDFBytes   = 3 << 20
)

var (
	errPDF      = stderrors.New("pdf_base64 must be a PDF file encoded in base64")
	errPDFLarge = stderrors.New("the PDF is over 3 MB. Ask the user to upload it in Plume instead")
)

// createDocument is plume_create_document: the PDF becomes a draft with its signers and a
// signature box for each, through the route Ardoise and the API use, never sent.
func (t *toolset) createDocument(ctx context.Context, req *sdk.CallToolRequest, in createInput) (*sdk.CallToolResult, documentOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, documentOutput{}, err
	}
	pdf, err := decodePDF(in.PDFBase64)
	if err != nil {
		return nil, documentOutput{}, err
	}
	if in.ClientID > 0 {
		if _, err := t.Clients.Get(ctx, userID, in.ClientID); err != nil {
			return nil, documentOutput{}, mcpkit.Failure(err, hintClient)
		}
	}
	created, _, err := t.Documents.CreateSignatureRequest(ctx, userID, draftRequest(in), fileNameOf(in), bytes.NewReader(pdf))
	if err != nil {
		return nil, documentOutput{}, mcpkit.Failure(err, hintSpace)
	}
	docID := created.Document.ID
	if in.ClientID > 0 {
		if _, err := t.Documents.Update(ctx, id(userID), id(docID), &documents.UpdateRequest{ClientID: &in.ClientID}); err != nil {
			return nil, documentOutput{}, mcpkit.Failure(err, hintClient)
		}
	}
	out, err := t.documentView(ctx, userID, docID)
	out.Status = statusCreated
	return nil, out, err
}

// draftRequest is the signature request of a draft: nothing is sent.
func draftRequest(in createInput) *documents.SignatureRequest {
	send := false
	request := &documents.SignatureRequest{Name: strings.TrimSpace(in.Name), Sequential: in.Sequential, Send: &send}
	if in.SpaceID > 0 {
		space := in.SpaceID
		request.SpaceID = &space
	}
	for _, s := range in.Signers {
		request.Signers = append(request.Signers, documents.SignatureSigner{
			Name: strings.TrimSpace(s.Name), Email: strings.TrimSpace(s.Email), Role: s.Role, Order: s.Order,
		})
	}
	return request
}

// fileNameOf is the stored file name: the one given, else the title, ending in .pdf.
func fileNameOf(in createInput) string {
	name := strings.TrimSpace(in.FileName)
	if name == "" {
		name = strings.TrimSpace(in.Name)
	}
	if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
		name += ".pdf"
	}
	return name
}

// decodePDF reads the base64 PDF, with or without a data: prefix or padding.
func decodePDF(encoded string) ([]byte, error) {
	_, data, found := strings.Cut(encoded, ";base64,")
	if !found {
		data = encoded
	}
	data = strings.TrimRight(strings.Join(strings.Fields(data), ""), "=")
	if base64.RawStdEncoding.DecodedLen(len(data)) > maxPDFBytes {
		return nil, errPDFLarge
	}
	raw, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil || !bytes.HasPrefix(raw, []byte("%PDF-")) {
		return nil, errPDF
	}
	return raw, nil
}
