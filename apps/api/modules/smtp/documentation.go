package smtp

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

// Documentation returns the smtp module's route table for the API reference.
func Documentation() documentation.Module {
	routes := getConfigRoute()
	routes = append(routes, saveConfigRoute()...)
	routes = append(routes, deleteConfigRoute()...)
	routes = append(routes, testConfigRoute()...)
	return documentation.Module{
		Name:        "smtp",
		Description: "Custom SMTP configuration routes for the calling user.",
		Routes:      routes,
	}
}

// getConfigRoute returns the GET /smtp route entry.
func getConfigRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/smtp",
			Summary:      "Get SMTP configuration",
			Description:  "Returns the custom SMTP configuration for the calling user, with the password omitted. Returns null if none is configured.",
			Auth:         "bearer",
			ResponseBody: ConfigResponse{},
		},
	}
}

// saveConfigRoute returns the PUT /smtp route entry.
func saveConfigRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "PUT",
			Path:         "/smtp",
			Summary:      "Save SMTP configuration",
			Description:  "Creates or replaces the custom SMTP configuration for the calling user. Pass an empty host to clear the configuration. When updating without providing a password, the existing password is kept.",
			Auth:         "bearer",
			RequestBody:  SaveConfigRequest{},
			ResponseBody: ConfigResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Port is not between 1 and 65535, or the from_email is not a valid address."},
			},
		},
	}
}

// deleteConfigRoute returns the DELETE /smtp route entry.
func deleteConfigRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:      "DELETE",
			Path:        "/smtp",
			Summary:     "Delete SMTP configuration",
			Description: "Clears the custom SMTP configuration for the calling user.",
			Auth:        "bearer",
		},
	}
}

// testConfigRoute returns the POST /smtp/test route entry.
func testConfigRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:      "POST",
			Path:        "/smtp/test",
			Summary:     "Send a test email",
			Description: "Sends a test email to the target address to verify the SMTP credentials work.",
			Auth:        "bearer",
			RequestBody: TestRequest{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "No SMTP configuration exists, or the destination address is invalid."},
				{Status: 502, Code: "upstream_error", Description: "The SMTP server rejected the connection or credentials."},
			},
		},
	}
}
