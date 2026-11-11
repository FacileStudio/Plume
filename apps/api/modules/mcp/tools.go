package mcp

import (
	"context"
	stderrors "errors"
	"strconv"

	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	toolList         = "plume_list_documents"
	toolGet          = "plume_get_document"
	toolCreate       = "plume_create_document"
	toolUpdate       = "plume_update_document"
	toolAddSigner    = "plume_add_signer"
	toolRemoveSigner = "plume_remove_signer"
	toolSend         = "plume_send_document"
	toolRemind       = "plume_remind_signer"
	toolDelete       = "plume_delete_document"
	toolClients      = "plume_list_clients"
	toolSaveClient   = "plume_save_client"
	hintDocument     = "Call plume_list_documents for the ids of the documents you can reach."
	hintSigner       = "Call plume_get_document for the ids of a document's signers."
	hintClient       = "Call plume_list_clients for the ids of your clients."
	hintSpace        = "plume_list_documents lists the spaces you belong to."
)

var errNoCaller = stderrors.New("this request carries no valid access token. Reconnect Plume")

// toolset holds what the tool handlers share.
type toolset struct {
	Deps
	confirm *mcpkit.Confirmations
	appURL  string
}

// userOf reads the Plume user id from the verified token, refused the way RequireAuth
// refuses it: a deleted account never acts.
func (t *toolset) userOf(ctx context.Context, req *sdk.CallToolRequest) (int64, error) {
	identity, ok := mcpauth.Identity(req.Extra.TokenInfo)
	if !ok {
		return 0, errNoCaller
	}
	if _, err := t.Accounts.Identity(ctx, identity.UserID); err != nil {
		return 0, errNoCaller
	}
	return identity.UserID, nil
}

// id renders a numeric id the way Plume's services take it.
func id(n int64) string {
	return strconv.FormatInt(n, 10)
}

// documentURL is the document's page in the web app.
func (t *toolset) documentURL(docID int64) string {
	return t.appURL + "/documents/" + id(docID)
}
