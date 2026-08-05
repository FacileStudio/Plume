package fields

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

var docField = []documentation.Field{
	{Name: "docId", Type: "int", Description: "Document ID"},
	{Name: "fieldId", Type: "int", Description: "Field ID"},
}

var Documentation = documentation.Module{
	Name:        "fields",
	Description: "Signature and input fields placed on a document.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/documents/{docId}/fields",
			Summary:      "List fields on a document",
			Auth:         "bearer",
			ResponseBody: "[]FieldResponse",
			PathParams:   []documentation.Field{{Name: "docId", Type: "int", Description: "Document ID"}},
		},
		{
			Method:       "POST",
			Path:         "/documents/{docId}/fields",
			Summary:      "Add a field to a document",
			Auth:         "bearer",
			RequestBody:  "CreateFieldRequest",
			ResponseBody: "FieldResponse",
			PathParams:   []documentation.Field{{Name: "docId", Type: "int", Description: "Document ID"}},
		},
		{
			Method:       "PUT",
			Path:         "/documents/{docId}/fields/{fieldId}",
			Summary:      "Update a field",
			Description:  "Moves, resizes or relabels a field. Only allowed while the document is a draft.",
			Auth:         "bearer",
			RequestBody:  "UpdateFieldRequest",
			ResponseBody: "FieldResponse",
			PathParams:   docField,
		},
		{
			Method:     "DELETE",
			Path:       "/documents/{docId}/fields/{fieldId}",
			Summary:    "Delete a field",
			Auth:       "bearer",
			PathParams: docField,
		},
	},
}
