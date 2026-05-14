package auth

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "auth",
	Description: "Authentication routes.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/auth/register",
			Summary:      "Register a new user",
			Description:  "Creates a user account and returns an auth token.",
			Auth:         "public",
			RequestBody:  "RegisterRequest",
			ResponseBody: "AuthResponse",
		},
		{
			Method:       "POST",
			Path:         "/auth/login",
			Summary:      "Authenticate a user",
			Description:  "Authenticates credentials and returns an auth token.",
			Auth:         "public",
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
			Description:  "Changes the authenticated user's password after verifying the current one.",
			Auth:         "bearer",
			RequestBody:  "ChangePasswordRequest",
		},
	},
}
