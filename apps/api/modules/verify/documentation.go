package verify

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "verify",
	Description: "Public signature verification by document hash. Rate limited.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/verify",
			Summary:      "Verify an uploaded PDF",
			Description:  "Public endpoint. Takes a multipart upload with a `file` part, hashes it, and reports the matching signed document if there is one.",
			Auth:         "",
			ResponseBody: "Response",
		},
		{
			Method:       "GET",
			Path:         "/verify/{hash}",
			Summary:      "Verify by hash",
			Description:  "Public endpoint. Looks up a signed document by its SHA-256 hash.",
			Auth:         "",
			ResponseBody: "Response",
			PathParams:   []documentation.Field{{Name: "hash", Type: "string", Description: "64-character hex SHA-256 of the document"}},
		},
	},
}
