package documents

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

// signatureRoutes returns the route entries for signature requests and the
// lookup by external_ref.
func signatureRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "POST",
			Path:         "/documents/signature-requests",
			Summary:      "Create a document for signature in one call",
			Description:  "Multipart form with the PDF in `file` and a JSON object in `request` (name, signers, optional fields, send, external_ref, nuage_space_facile_id, nuage_linked_to, client). Creates the document, its signers and fields, and sends it unless send is false. An optional client ({name, email, company}) files the document under the caller's client with that email, creating it when missing. A repeated external_ref for the same caller returns the existing document with 200 instead of creating a second one.",
			Auth:         "bearer",
			RequestBody:  SignatureRequest{},
			ResponseBody: SignatureResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON, signer, field, reference or PDF."},
			},
		},
		{
			Method:       "GET",
			Path:         "/documents/lookup",
			Summary:      "Read a document's status by external_ref",
			Description:  "Query parameter external_ref. Returns the document and its signers, scoped to the caller. For polling when a webhook was missed.",
			Auth:         "bearer",
			ResponseBody: SignatureResponse{},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "No document with this external_ref."},
			},
		},
	}
}
