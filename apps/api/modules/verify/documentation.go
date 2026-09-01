package verify

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "verify",
	Description: "Public verification routes for signed PDF documents.",
	Routes: []documentation.Route{
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
			},
		},
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
			},
		},
	},
}
