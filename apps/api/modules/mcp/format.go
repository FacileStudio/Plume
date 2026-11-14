package mcp

import (
	"strings"

	"github.com/FacileStudio/Plume/apps/api/modules/clients"
	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	"github.com/FacileStudio/Plume/apps/api/modules/fields"
	"github.com/FacileStudio/Plume/apps/api/modules/signers"
	"github.com/FacileStudio/Plume/apps/api/modules/spaces"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

const (
	documentPage = 25
	clientPage   = 50
	notesLimit   = 500
)

// document renders one document with its page in the web app.
func (t *toolset) document(d *documents.DocumentResponse) documentOut {
	return documentOut{
		ID: d.ID, Name: d.Name, Status: d.Status, FileName: d.FileName, Sequential: d.Sequential,
		ClientID: d.ClientID, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt, URL: t.documentURL(d.ID),
	}
}

// signerOf renders one signer with the number of fields they fill.
func signerOf(s signers.SignerResponse, fieldCount map[int64]int) signerOut {
	return signerOut{
		ID: s.ID, Name: s.Name, Email: s.Email, Role: s.Role, Status: s.Status, Order: s.OrderNum,
		Fields: fieldCount[s.ID], ViewedAt: s.ViewedAt, SignedAt: s.SignedAt, LastRemindedAt: s.LastRemindedAt,
	}
}

// fieldsPerSigner counts the fields of each signer.
func fieldsPerSigner(list []fields.FieldResponse) map[int64]int {
	count := make(map[int64]int, len(list))
	for _, f := range list {
		count[f.SignerID]++
	}
	return count
}

// fieldsOf renders the fields placed on a document.
func fieldsOf(list []fields.FieldResponse) []fieldOut {
	out := make([]fieldOut, len(list))
	for i, f := range list {
		out[i] = fieldOut{ID: f.ID, SignerID: f.SignerID, Type: f.FieldType, Page: f.Page, Required: f.Required, Label: f.Label}
	}
	return out
}

// clientOf renders one client, with long notes cut.
func clientOf(c *clients.ClientResponse) clientOut {
	notes, _ := mcpkit.Clip(c.Notes, notesLimit)
	return clientOut{ID: c.ID, Name: c.Name, Email: c.Email, Company: c.Company, Phone: c.Phone, Notes: notes}
}

// spacesOf renders the person's spaces.
func spacesOf(list []spaces.SpaceResponse) []spaceOut {
	out := make([]spaceOut, len(list))
	for i, s := range list {
		out[i] = spaceOut{ID: s.ID, Name: s.Name, Role: s.Role}
	}
	return out
}

// contains reports whether any of values holds query, case ignored. An empty query
// matches everything.
func contains(query string, values ...string) bool {
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return true
	}
	for _, v := range values {
		if strings.Contains(strings.ToLower(v), needle) {
			return true
		}
	}
	return false
}
