package auth

import (
	documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"
	"github.com/FacileStudio/porte"
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

// credentialRoutes documents the local registration and login routes, which
// are only mounted when SSO-only mode is disabled.
func credentialRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "POST",
			Path:         "/auth/register",
			Summary:      "Register a new user",
			Description:  "Creates a user account and returns an auth token.",
			Auth:         "",
			RequestBody:  RegisterRequest{},
			ResponseBody: AuthResponse{},
		},
		{
			Method:       "POST",
			Path:         "/auth/login",
			Summary:      "Authenticate a user",
			Description:  "Authenticates credentials and returns an auth token.",
			Auth:         "",
			RequestBody:  LoginRequest{},
			ResponseBody: AuthResponse{},
		},
	}
}

// profileRoutes documents the authenticated self-service routes for the
// current user's profile and password.
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
		{
			Method:       "PUT",
			Path:         "/auth/password",
			Summary:      "Change password",
			Description:  "Replaces the authenticated user's password after verifying the current one, ends the account's other logins and returns the rotated session token. Omitting current_password adds a first password to an account that has none, and is refused with 400 for an account that already has one.",
			Auth:         "bearer",
			RequestBody:  ChangePasswordRequest{},
			ResponseBody: ChangePasswordResponse{},
		},
	}
}

// Documentation returns the auth module's route table for the API reference.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "auth",
		Description: "Authentication routes.",
		Routes:      append(configRoutes(), append(credentialRoutes(), profileRoutes()...)...),
	}
}
