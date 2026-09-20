package clients

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

// DeleteClientResponse is the payload returned when a client contact is deleted.
type DeleteClientResponse struct {
	Deleted bool `json:"deleted"`
}

// Documentation returns the clients module's route table for the API reference.
func Documentation() documentation.Module {
	return documentation.Module{
		Name:        "clients",
		Description: "Client contact management routes.",
		Routes:      append(collectionRoutes(), itemRoutes()...),
	}
}

// collectionRoutes returns the /clients collection route entries.
func collectionRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "POST",
			Path:         "/clients",
			Summary:      "Create a client",
			Description:  "Creates a new client contact for the calling user.",
			Auth:         "bearer",
			RequestBody:  CreateClientRequest{},
			ResponseBody: ClientResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body, missing name or invalid email."},
			},
		},
		{
			Method:       "GET",
			Path:         "/clients",
			Summary:      "List clients",
			Description:  "Returns all client contacts for the calling user, ordered by creation date descending.",
			Auth:         "bearer",
			ResponseBody: []ClientResponse{},
		},
	}
}

// itemRoutes returns the /clients/{clientId} item route entries.
func itemRoutes() []documentation.Route {
	return append(getClientRoute(), append(updateClientRoute(), deleteClientRoute()...)...)
}

// getClientRoute returns the GET /clients/{clientId} route entry.
func getClientRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/clients/{clientId}",
			Summary:      "Get a client",
			Description:  "Returns a single client contact by ID.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "clientId", Type: "int", Description: "Client ID"}},
			ResponseBody: ClientResponse{},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Client not found."},
			},
		},
	}
}

// updateClientRoute returns the PUT /clients/{clientId} route entry.
func updateClientRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "PUT",
			Path:         "/clients/{clientId}",
			Summary:      "Update a client",
			Description:  "Updates a client contact's name, email, company or notes.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "clientId", Type: "int", Description: "Client ID"}},
			RequestBody:  UpdateClientRequest{},
			ResponseBody: ClientResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body, missing name or invalid email."},
				{Status: 404, Code: "not_found", Description: "Client not found."},
			},
		},
	}
}

// deleteClientRoute returns the DELETE /clients/{clientId} route entry.
func deleteClientRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:      "DELETE",
			Path:        "/clients/{clientId}",
			Summary:     "Delete a client",
			Description: "Deletes a client contact. Fails if the client is referenced by existing documents.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "clientId", Type: "int", Description: "Client ID"}},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Client is in use by one or more documents."},
				{Status: 404, Code: "not_found", Description: "Client not found."},
			},
		},
	}
}
