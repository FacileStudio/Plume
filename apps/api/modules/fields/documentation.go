package fields

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

// DeleteFieldResponse is the payload returned when a signing field is removed.
type DeleteFieldResponse struct {
	Deleted bool `json:"deleted"`
}

// Documentation returns the fields module's route table for the API reference.
func Documentation() documentation.Module {
	routes := listFieldsRoute()
	routes = append(routes, createFieldRoute()...)
	routes = append(routes, updateFieldRoute()...)
	routes = append(routes, deleteFieldRoute()...)
	return documentation.Module{
		Name:        "fields",
		Description: "Document signing field routes (signature, text, date, checkbox).",
		Routes:      routes,
	}
}

// listFieldsRoute returns the GET /documents/{docId}/fields route entry.
func listFieldsRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/documents/{docId}/fields",
			Summary:      "List fields on a document",
			Description:  "Returns all fields placed on a document, with their positions, dimensions and assigned signers.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "docId", Type: "int", Description: "Document ID"}},
			ResponseBody: []FieldResponse{},
		},
	}
}

// createFieldRoute returns the POST /documents/{docId}/fields route entry.
func createFieldRoute() []documentation.Route {
	return []documentation.Route{
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
	}
}

// updateFieldRoute returns the PUT /documents/{docId}/fields/{fieldId} route entry.
func updateFieldRoute() []documentation.Route {
	return []documentation.Route{
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
	}
}

// deleteFieldRoute returns the DELETE /documents/{docId}/fields/{fieldId} route entry.
func deleteFieldRoute() []documentation.Route {
	return []documentation.Route{
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
	}
}
