package documents

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "documents",
	Description: "Manage documents for signing.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/documents",
			Summary:      "Create a document",
			Description:  "Creates a new document in draft status.",
			Auth:         "bearer",
			RequestBody:  "CreateRequest",
			ResponseBody: "DocumentResponse",
		},
		{
			Method:       "GET",
			Path:         "/documents",
			Summary:      "List documents",
			Description:  "Returns all documents owned by the authenticated user. Supports ?status= filter.",
			Auth:         "bearer",
			ResponseBody: "[]DocumentResponse",
		},
		{
			Method:       "GET",
			Path:         "/documents/{id}",
			Summary:      "Get a document",
			Auth:         "bearer",
			ResponseBody: "DocumentResponse",
			PathParams:   []documentation.Field{{Name: "id", Type: "int", Description: "Document ID"}},
		},
		{
			Method:      "PUT",
			Path:        "/documents/{id}",
			Summary:     "Update a document",
			Auth:        "bearer",
			RequestBody: "UpdateRequest",
			PathParams:  []documentation.Field{{Name: "id", Type: "int", Description: "Document ID"}},
		},
		{
			Method:     "DELETE",
			Path:       "/documents/{id}",
			Summary:    "Delete a document",
			Auth:       "bearer",
			PathParams: []documentation.Field{{Name: "id", Type: "int", Description: "Document ID"}},
		},
		{
			Method:       "POST",
			Path:         "/documents/{id}/send",
			Summary:      "Send document for signing",
			Description:  "Sets status to pending and generates tokens for all signers.",
			Auth:         "bearer",
			ResponseBody: "DocumentResponse",
			PathParams:   []documentation.Field{{Name: "id", Type: "int", Description: "Document ID"}},
		},
	},
}
