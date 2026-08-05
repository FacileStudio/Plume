package smtp

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "smtp",
	Description: "Per-user outgoing mail configuration.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/smtp",
			Summary:      "Get the SMTP configuration",
			Description:  "Returns the stored configuration without the password.",
			Auth:         "bearer",
			ResponseBody: "ConfigResponse",
		},
		{
			Method:       "PUT",
			Path:         "/smtp",
			Summary:      "Save the SMTP configuration",
			Auth:         "bearer",
			RequestBody:  "SaveConfigRequest",
			ResponseBody: "ConfigResponse",
		},
		{
			Method:  "DELETE",
			Path:    "/smtp",
			Summary: "Delete the SMTP configuration",
			Auth:    "bearer",
		},
		{
			Method:      "POST",
			Path:        "/smtp/test",
			Summary:     "Send a test email",
			Description: "Sends a test message through the stored configuration to verify it works.",
			Auth:        "bearer",
			RequestBody: "TestRequest",
		},
	},
}
