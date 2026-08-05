package signing

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "signing",
	Description: "Evidence produced once a document is signed.",
	Routes: []documentation.Route{
		{
			Method:      "GET",
			Path:        "/documents/{docId}/certificate",
			Summary:     "Download the signing certificate",
			Description: "Returns the generated certificate of completion as application/pdf rather than JSON.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "docId", Type: "int", Description: "Document ID"}},
		},
		{
			Method:      "GET",
			Path:        "/documents/{docId}/audit-trail",
			Summary:     "Download the audit trail",
			Description: "Returns the generated audit trail of signing events as application/pdf rather than JSON.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "docId", Type: "int", Description: "Document ID"}},
		},
	},
}
