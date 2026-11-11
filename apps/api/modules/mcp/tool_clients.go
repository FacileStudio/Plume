package mcp

import (
	"context"
	stderrors "errors"
	"strings"

	"github.com/FacileStudio/Plume/apps/api/modules/clients"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var errClientName = stderrors.New("a new client needs a name")

// listClients is plume_list_clients: the person's clients, newest first.
func (t *toolset) listClients(ctx context.Context, req *sdk.CallToolRequest, in clientsInput) (*sdk.CallToolResult, clientsOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, clientsOutput{}, err
	}
	all, err := t.Clients.List(ctx, userID)
	if err != nil {
		return nil, clientsOutput{}, mcpkit.Failure(err, "")
	}
	kept := all[:0:0]
	for _, c := range all {
		if contains(in.Query, c.Name, c.Company, c.Email) {
			kept = append(kept, c)
		}
	}
	page, next, err := pageOf(kept, in.Cursor, clientPage)
	if err != nil {
		return nil, clientsOutput{}, err
	}
	out := clientsOutput{Clients: make([]clientOut, len(page)), NextCursor: next}
	for i := range page {
		out.Clients[i] = clientOf(&page[i])
	}
	return nil, out, nil
}

// saveClient is plume_save_client: a new client, or the given fields of one.
func (t *toolset) saveClient(ctx context.Context, req *sdk.CallToolRequest, in saveClientInput) (*sdk.CallToolResult, clientOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, clientOutput{}, err
	}
	if in.ClientID == 0 {
		return t.createClient(ctx, userID, in)
	}
	current, err := t.Clients.Get(ctx, userID, in.ClientID)
	if err != nil {
		return nil, clientOutput{}, mcpkit.Failure(err, hintClient)
	}
	change := &clients.UpdateClientRequest{
		Name: strings.TrimSpace(in.Name), Email: or(in.Email, current.Email), Company: or(in.Company, current.Company),
		Phone: or(in.Phone, current.Phone), Notes: or(in.Notes, current.Notes),
	}
	saved, err := t.Clients.Update(ctx, userID, in.ClientID, change)
	if err != nil {
		return nil, clientOutput{}, mcpkit.Failure(err, hintClient)
	}
	return nil, clientOutput{Status: statusUpdated, Client: clientOf(saved)}, nil
}

// createClient adds a client.
func (t *toolset) createClient(ctx context.Context, userID int64, in saveClientInput) (*sdk.CallToolResult, clientOutput, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, clientOutput{}, errClientName
	}
	saved, err := t.Clients.Create(ctx, userID, &clients.CreateClientRequest{
		Name: strings.TrimSpace(in.Name), Email: or(in.Email, ""), Company: or(in.Company, ""),
		Phone: or(in.Phone, ""), Notes: or(in.Notes, ""),
	})
	if err != nil {
		return nil, clientOutput{}, mcpkit.Failure(err, "")
	}
	return nil, clientOutput{Status: statusCreated, Client: clientOf(saved)}, nil
}

// or is the given value, trimmed, or fallback when it was left out.
func or(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return strings.TrimSpace(*value)
}
