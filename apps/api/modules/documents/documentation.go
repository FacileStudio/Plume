package documents

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

type CreateRequest struct {
	Name     string `json:"name"`
	FileName string `json:"file_name,omitempty"`
}

type DeleteDocumentResponse struct {
	Deleted bool `json:"deleted"`
}

var Documentation = documentation.Module{
	Name:        "documents",
	Description: "Document lifecycle routes: upload, edit, send, download, stats and deletion.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/documents",
			Summary:      "Upload a document",
			Description:  "Uploads a PDF document via multipart form. The file is saved to storage and a draft document record is created.",
			Auth:         "bearer",
			RequestBody:  CreateRequest{},
			ResponseBody: DocumentResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Missing file, file is not a PDF, or missing name."},
			},
		},
		{
			Method:       "GET",
			Path:         "/documents",
			Summary:      "List documents",
			Description:  "Returns all documents owned by the calling user, optionally filtered by status, search query or tag, with pagination.",
			Auth:         "bearer",
			ResponseBody: []DocumentResponse{},
		},
		{
			Method:       "GET",
			Path:         "/documents/stats",
			Summary:      "Get document statistics",
			Description:  "Returns aggregate counts of documents grouped by status (draft, sent, completed, declined, voided).",
			Auth:         "bearer",
			ResponseBody: StatsResponse{},
		},
		{
			Method:       "GET",
			Path:         "/documents/{id}",
			Summary:      "Get a document",
			Description:  "Returns a single document with its signers, fields, custom branding and audit trail.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int", Description: "Document ID"}},
			ResponseBody: DocumentResponse{},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Document not found."},
			},
		},
		{
			Method:      "GET",
			Path:        "/documents/{id}/file",
			Summary:     "Download document file",
			Description: "Streams the original or signed PDF document file.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "id", Type: "int", Description: "Document ID"}},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Document or file not found."},
			},
		},
		{
			Method:       "PUT",
			Path:         "/documents/{id}",
			Summary:      "Update a document",
			Description:  "Updates document metadata (name, message, expiration date, branding, tags).",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int", Description: "Document ID"}},
			RequestBody:  UpdateRequest{},
			ResponseBody: DocumentResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body."},
				{Status: 404, Code: "not_found", Description: "Document not found."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/documents/{id}",
			Summary:     "Delete a document",
			Description: "Deletes a document and its associated storage files. Only draft or voided documents can be deleted.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "id", Type: "int", Description: "Document ID"}},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Document is sent or completed and cannot be deleted."},
				{Status: 404, Code: "not_found", Description: "Document not found."},
			},
		},
		{
			Method:       "POST",
			Path:         "/documents/{id}/send",
			Summary:      "Send a document for signing",
			Description:  "Transitions a draft document to sent status. Validates that at least one signer and one signature field exist, generates signing tokens, and dispatches invitation emails.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int", Description: "Document ID"}},
			ResponseBody: DocumentResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Document has no signers, missing required fields, or is not in draft."},
				{Status: 404, Code: "not_found", Description: "Document not found."},
			},
		},
	},
}
