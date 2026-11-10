package verify

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

// Documentation returns the verify module's route table for the API reference.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "verify",
		Description: "Public verification routes for signed PDF documents.",
		Routes:      append(hashRoutes(), uploadRoutes()...),
	}
}

// hashRoutes returns the GET /verify/{hash} route entry.
func hashRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/verify/{hash}",
			Summary:      "Verify document status by hash",
			Description:  "Returns public verification metadata for a document: name, hash, signature status, and signer details (with sensitive info omitted).",
			Auth:         "",
			PathParams:   []documentation.Field{{Name: "hash", Type: "string", Description: "Document SHA-256 hash"}},
			ResponseBody: Response{},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Document not found or not yet signed."},
				{Status: 429, Code: "rate_limited", Description: "More than 30 verify requests from this address in a minute. Retry-After gives the wait in seconds."},
			},
		},
	}
}

// uploadRoutes returns the POST /verify route entry.
func uploadRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "POST",
			Path:         "/verify",
			Summary:      "Verify an uploaded PDF file",
			Description:  "Accepts a multipart PDF upload, computes its SHA-256 hash, and verifies it against completed documents in the system.",
			Auth:         "",
			ResponseBody: Response{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Missing file or file is not a valid PDF."},
				{Status: 404, Code: "not_found", Description: "No matching signed document found for this file hash."},
				{Status: 429, Code: "rate_limited", Description: "More than 30 verify requests from this address in a minute. Retry-After gives the wait in seconds."},
			},
		},
	}
}
