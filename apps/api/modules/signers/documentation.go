package signers

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "signers",
	Description: "Manage document signers and signing flow.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/documents/{docId}/signers",
			Summary:      "List signers for a document",
			Auth:         "bearer",
			ResponseBody: "[]SignerResponse",
			PathParams:   []documentation.Field{{Name: "docId", Type: "int", Description: "Document ID"}},
		},
		{
			Method:       "POST",
			Path:         "/documents/{docId}/signers",
			Summary:      "Add a signer to a document",
			Auth:         "bearer",
			RequestBody:  "AddSignerRequest",
			ResponseBody: "SignerResponse",
			PathParams:   []documentation.Field{{Name: "docId", Type: "int", Description: "Document ID"}},
		},
		{
			Method:     "DELETE",
			Path:       "/signers/{id}",
			Summary:    "Remove a signer",
			Auth:       "bearer",
			PathParams: []documentation.Field{{Name: "id", Type: "int", Description: "Signer ID"}},
		},
		{
			Method:       "GET",
			Path:         "/sign/{token}",
			Summary:      "Get signing view",
			Description:  "Public endpoint. Returns the document, signer info, and fields for signing.",
			Auth:         "public",
			ResponseBody: "SigningView",
			PathParams:   []documentation.Field{{Name: "token", Type: "string", Description: "Signer token"}},
		},
		{
			Method:      "POST",
			Path:        "/sign/{token}",
			Summary:     "Submit signature",
			Description: "Public endpoint. Submits field values and records the signature.",
			Auth:        "public",
			RequestBody: "SubmitSignatureRequest",
			PathParams:  []documentation.Field{{Name: "token", Type: "string", Description: "Signer token"}},
		},
	},
}
