package auth

import (
	documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/kit"
)

// configRoutes documents the public routes that porte itself serves: reading
// the auth configuration and ending the session.
func configRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/auth/config",
			Summary:      "Get the auth configuration",
			Description:  "Public endpoint. Returns sso_only, oidc_enabled and, when OIDC is configured, oidc_issuer and oidc_redirect_url, so the login page can render the right options.",
			Auth:         "",
			ResponseBody: porte.ConfigResponse{},
		},
		{
			Method:       "POST",
			Path:         "/auth/logout",
			Summary:      "End the session",
			Description:  "Revokes the session and clears the cookie. Idempotent: a caller whose session has already expired still gets 200, because that is exactly when somebody needs to log out. Served by porte.",
			Auth:         "",
			ResponseBody: porte.LogoutResponse{},
		},
	}
}

// profileRoutes documents the authenticated self-service routes for the
// current user's profile.
func profileRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/auth/me",
			Summary:      "Get current user profile",
			Description:  "Returns the authenticated user's profile.",
			Auth:         "bearer",
			ResponseBody: ProfileResponse{},
		},
		{
			Method:       "PUT",
			Path:         "/auth/me",
			Summary:      "Update current user profile",
			Description:  "Updates the authenticated user's name and email.",
			Auth:         "bearer",
			RequestBody:  UpdateProfileRequest{},
			ResponseBody: ProfileResponse{},
		},
	}
}

// Documentation returns the auth module's route table for the API reference.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "auth",
		Description: "Authentication routes.",
		Routes:      append(configRoutes(), append(profileRoutes(), append(apiTokenRoutes(), invitationsRoute())...)...),
	}
}

// invitationsRoute documents GET /me/invitations, which porte/kit serves when
// OIDC is configured.
func invitationsRoute() documentation.Route {
	route := kit.InvitationsRoute()
	route.ResponseBody = kit.InvitationsResponse{}
	return route
}
