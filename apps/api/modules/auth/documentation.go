package auth

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "auth",
	Description: "Authentication routes.",
	Routes: []documentation.Route{
		{
			Method:      "GET",
			Path:        "/auth/config",
			Summary:     "Get the auth configuration",
			Description: "Public endpoint. Returns sso_only, oidc_enabled and, when OIDC is configured, oidc_issuer and oidc_redirect_url, so the login page can render the right options.",
			Auth:        "",
		},
		{
			Method:      "POST",
			Path:        "/auth/logout",
			Summary:     "End the session",
			Description: "Revokes the session and clears the cookie. Idempotent: a caller whose session has already expired still gets 200, because that is exactly when somebody needs to log out. Served by porte.",
			Auth:        "",
		},
		{
			Method:       "POST",
			Path:         "/auth/register",
			Summary:      "Register a new user",
			Description:  "Creates a user account and returns an auth token.",
			Auth:         "",
			RequestBody:  "RegisterRequest",
			ResponseBody: "AuthResponse",
		},
		{
			Method:       "POST",
			Path:         "/auth/login",
			Summary:      "Authenticate a user",
			Description:  "Authenticates credentials and returns an auth token.",
			Auth:         "",
			RequestBody:  "LoginRequest",
			ResponseBody: "AuthResponse",
		},
		{
			Method:       "GET",
			Path:         "/auth/me",
			Summary:      "Get current user profile",
			Description:  "Returns the authenticated user's profile.",
			Auth:         "bearer",
			ResponseBody: "ProfileResponse",
		},
		{
			Method:       "PUT",
			Path:         "/auth/me",
			Summary:      "Update current user profile",
			Description:  "Updates the authenticated user's name and email.",
			Auth:         "bearer",
			RequestBody:  "UpdateProfileRequest",
			ResponseBody: "ProfileResponse",
		},
		{
			Method:       "PUT",
			Path:         "/auth/password",
			Summary:      "Change password",
			Description:  "Replaces the authenticated user's password after verifying the current one, ends the account's other logins and returns the rotated session token. Omitting current_password adds a first password to an account that has none, and is refused with 400 for an account that already has one.",
			Auth:         "bearer",
			RequestBody:  "ChangePasswordRequest",
			ResponseBody: "ChangePasswordResponse",
		},
	},
}
