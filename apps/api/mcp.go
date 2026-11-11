package main

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/internal/version"
	"github.com/FacileStudio/Plume/apps/api/modules/mcp"
	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// newMCP builds the MCP server when MCP_RESOURCE is set and SSO is on, and nothing
// otherwise. The verifier runs Registre discovery here, so a wrong issuer fails at boot
// like the OIDC kit does. A person's first MCP call creates their account through the
// OIDC kit, as a first web login would.
func newMCP(ctx context.Context, appEnv env.Config, parts authParts, svc *services, appLogger *slog.Logger) (*mcp.Service, error) {
	if appEnv.OIDC == nil || appEnv.OIDC.MCPResource == "" {
		return nil, nil
	}
	verifier, err := mcpauth.Verifier(ctx, mcpauth.Config{
		Issuer:      appEnv.OIDC.Issuer,
		Resource:    appEnv.OIDC.MCPResource,
		ScopePrefix: "plume",
		Identities:  parts.identities,
		Provision:   true,
		Accounts:    parts.oidc,
		Logger:      appLogger,
	})
	if err != nil {
		return nil, err
	}
	appLogger.Info("mcp server on", slog.String("resource", appEnv.OIDC.MCPResource))
	return mcp.NewService(mcp.Deps{
		Documents: svc.documents, Signers: svc.signers, Fields: svc.fields, Reminders: svc.reminders,
		Clients: svc.clients, Spaces: svc.spaces, Accounts: parts.users,
	}, mcp.Config{
		Resource: appEnv.OIDC.MCPResource,
		Issuer:   appEnv.OIDC.Issuer,
		Version:  version.String(),
		Origins:  appEnv.CORSAllowedOrigins,
		StateKey: appEnv.OIDC.MCPStateKey(),
		Verifier: verifier,
		Pending:  mcpkit.PostgresPending{DB: parts.users.DB},
		Logger:   appLogger,
	})
}
