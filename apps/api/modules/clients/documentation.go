package clients

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "clients",
	Description: "Address book of recurring signing counterparties.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/clients",
			Summary:      "Create a client",
			Auth:         "bearer",
			RequestBody:  "CreateClientRequest",
			ResponseBody: "ClientResponse",
		},
		{
			Method:       "GET",
			Path:         "/clients",
			Summary:      "List clients",
			Description:  "Returns every client owned by the authenticated user.",
			Auth:         "bearer",
			ResponseBody: "[]ClientResponse",
		},
		{
			Method:       "GET",
			Path:         "/clients/{clientId}",
			Summary:      "Get a client",
			Auth:         "bearer",
			ResponseBody: "ClientResponse",
			PathParams:   []documentation.Field{{Name: "clientId", Type: "string", Description: "Client ID"}},
		},
		{
			Method:       "PUT",
			Path:         "/clients/{clientId}",
			Summary:      "Update a client",
			Auth:         "bearer",
			RequestBody:  "UpdateClientRequest",
			ResponseBody: "ClientResponse",
			PathParams:   []documentation.Field{{Name: "clientId", Type: "string", Description: "Client ID"}},
		},
		{
			Method:     "DELETE",
			Path:       "/clients/{clientId}",
			Summary:    "Delete a client",
			Auth:       "bearer",
			PathParams: []documentation.Field{{Name: "clientId", Type: "string", Description: "Client ID"}},
		},
	},
}
