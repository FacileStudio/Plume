package signers

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

type RemindResponse struct {
	Message string `json:"message"`
}

type DeclineResponse struct {
	Declined bool `json:"declined"`
}

type DeleteSignerResponse struct {
	Deleted bool `json:"deleted"`
}

var Documentation = documentation.Module{
	Name:        "signers",
	Description: "Signer management and signing workflow routes.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/documents/{docId}/signers",
			Summary:      "List signers on a document",
			Description:  "Returns all signers invited onto a document, ordered by their signing sequence.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "docId", Type: "int", Description: "Document ID"}},
			ResponseBody: []SignerResponse{},
		},
		{
			Method:       "POST",
			Path:         "/documents/{docId}/signers",
			Summary:      "Add a signer to a document",
			Description:  "Invites a signer onto a document in draft status. Fails if the document has already been sent.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "docId", Type: "int", Description: "Document ID"}},
			RequestBody:  AddSignerRequest{},
			ResponseBody: SignerResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body, missing name/email, or document is not in draft."},
				{Status: 404, Code: "not_found", Description: "Document not found."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/signers/{id}",
			Summary:     "Remove a signer",
			Description: "Removes a signer and their associated fields from a document in draft status.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "id", Type: "int", Description: "Signer ID"}},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Document is not in draft."},
				{Status: 404, Code: "not_found", Description: "Signer not found."},
			},
		},
		{
			Method:       "POST",
			Path:         "/signers/{id}/remind",
			Summary:      "Send a reminder email to a signer",
			Description:  "Re-sends the signing invitation email to a signer who has not yet signed.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int", Description: "Signer ID"}},
			ResponseBody: RemindResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Signer has already signed or document is not in sent status."},
				{Status: 404, Code: "not_found", Description: "Signer not found."},
			},
		},
		{
			Method:       "GET",
			Path:         "/sign/{token}",
			Summary:      "Get the signing view for a token",
			Description:  "Returns the document, the signer's own fields and the fields already completed by prior signers. Marks the token as viewed on first access.",
			Auth:         "",
			PathParams:   []documentation.Field{{Name: "token", Type: "string", Description: "Signing token"}},
			ResponseBody: SigningView{},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Token not found or document is not in sent status."},
			},
		},
		{
			Method:       "GET",
			Path:         "/sign/{token}/status",
			Summary:      "Get document signing progress for a token",
			Description:  "Returns a token-scoped summary of signing progress across all participants. Safe to call in any non-draft state (including completed) to render the post-signing status screen.",
			Auth:         "",
			PathParams:   []documentation.Field{{Name: "token", Type: "string", Description: "Signing token"}},
			ResponseBody: SigningStatusResponse{},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Token not found or document is still in draft."},
			},
		},
		{
			Method:      "GET",
			Path:        "/sign/{token}/opened.gif",
			Summary:     "Track email open event",
			Description: "1x1 transparent tracking pixel embedded in signing invitation emails.",
			Auth:        "",
			PathParams:  []documentation.Field{{Name: "token", Type: "string", Description: "Signing token"}},
		},
		{
			Method:      "GET",
			Path:        "/sign/{token}/file",
			Summary:     "Download document file for signing",
			Description: "Streams the original PDF for the recipient to inspect.",
			Auth:        "",
			PathParams:  []documentation.Field{{Name: "token", Type: "string", Description: "Signing token"}},
		},
		{
			Method:      "POST",
			Path:        "/sign/{token}",
			Summary:     "Submit a signature",
			Description: "Submits the field values for a signer. Validates all required fields, burns the field values onto the PDF, advances to the next signer or completes the document, and triggers notifications.",
			Auth:        "",
			PathParams:  []documentation.Field{{Name: "token", Type: "string", Description: "Signing token"}},
			RequestBody: SubmitSignatureRequest{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Missing required fields, or the signer has already signed."},
				{Status: 404, Code: "not_found", Description: "Token not found."},
			},
		},
		{
			Method:       "POST",
			Path:         "/sign/{token}/decline",
			Summary:      "Decline to sign",
			Description:  "Marks the signer as declined and marks the document as declined.",
			Auth:         "",
			PathParams:   []documentation.Field{{Name: "token", Type: "string", Description: "Signing token"}},
			ResponseBody: DeclineResponse{},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Token not found."},
			},
		},
	},
}
