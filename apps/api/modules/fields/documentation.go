package fields

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

type DeleteFieldResponse struct {
	Deleted bool `json:"deleted"`
}

var Documentation = documentation.Module{
	Name:        "fields",
	Description: "Document signing field routes (signature, text, date, checkbox).",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/documents/{docId}/fields",
			Summary:      "List fields on a document",
			Description:  "Returns all fields placed on a document, with their positions, dimensions and assigned signers.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "docId", Type: "int", Description: "Document ID"}},
			ResponseBody: []FieldResponse{},
		},
		{
			Method:       "POST",
			Path:         "/documents/{docId}/fields",
			Summary:      "Add a field to a document",
			Description:  "Places a new signing field on a draft document page. Validates field type, page bounds and coordinates.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "docId", Type: "int", Description: "Document ID"}},
			RequestBody:  CreateFieldRequest{},
			ResponseBody: FieldResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid field type, coordinates out of bounds, or document is not in draft."},
				{Status: 404, Code: "not_found", Description: "Document or signer not found."},
			},
		},
		{
			Method:      "PUT",
			Path:        "/documents/{docId}/fields/{fieldId}",
			Summary:     "Update a field",
			Description: "Updates a field's position, dimensions, label, required flag or assigned signer on a draft document.",
			Auth:        "bearer",
			PathParams: []documentation.Field{
				{Name: "docId", Type: "int", Description: "Document ID"},
				{Name: "fieldId", Type: "int", Description: "Field ID"},
			},
			RequestBody:  UpdateFieldRequest{},
			ResponseBody: FieldResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid coordinates or document is not in draft."},
				{Status: 404, Code: "not_found", Description: "Field not found."},
			},
		},
		{
			Method:      "DELETE",
			Path:        "/documents/{docId}/fields/{fieldId}",
			Summary:     "Delete a field",
			Description: "Removes a field from a draft document.",
			Auth:        "bearer",
			PathParams: []documentation.Field{
				{Name: "docId", Type: "int", Description: "Document ID"},
				{Name: "fieldId", Type: "int", Description: "Field ID"},
			},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Document is not in draft."},
				{Status: 404, Code: "not_found", Description: "Field not found."},
			},
		},
	},
}
