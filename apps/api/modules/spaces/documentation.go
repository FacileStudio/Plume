package spaces

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

// DeleteSpaceResponse is the payload returned when a space is deleted.
type DeleteSpaceResponse struct {
	Deleted bool `json:"deleted"`
}

// LeaveSpaceResponse is the payload returned when the caller leaves a space.
type LeaveSpaceResponse struct {
	Left bool `json:"left"`
}

// RemoveMemberResponse is the payload returned when a member is removed.
type RemoveMemberResponse struct {
	Removed bool `json:"removed"`
}

// Documentation returns the spaces module's route table for the API reference.
func Documentation() documentation.Module {
	routes := spaceRoutes()
	routes = append(routes, getSpaceRoute()...)
	routes = append(routes, updateSpaceRoute()...)
	routes = append(routes, deleteSpaceRoute()...)
	routes = append(routes, leaveSpaceRoute()...)
	routes = append(routes, memberRoutes()...)
	routes = append(routes, syncRoute()...)
	return documentation.Module{
		Name:        "spaces",
		Description: "Collaborative space management and membership routes.",
		Routes:      routes,
	}
}

// spaceRoutes returns the /spaces collection route entries.
func spaceRoutes() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "POST",
			Path:         "/spaces",
			Summary:      "Create a space",
			Description:  "Creates a new collaborative space. The creator becomes the owner.",
			Auth:         "bearer",
			RequestBody:  CreateSpaceRequest{},
			ResponseBody: SpaceResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body, missing name or invalid color/icon."},
				{Status: 409, Code: "already_exists", Description: "Registre manages this app's spaces; create it there."},
			},
		},
		{
			Method:       "GET",
			Path:         "/spaces",
			Summary:      "List spaces",
			Description:  "Returns all spaces the calling user is a member of.",
			Auth:         "bearer",
			ResponseBody: []SpaceResponse{},
		},
	}
}

// getSpaceRoute returns the GET /spaces/{spaceId} route entry.
func getSpaceRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/spaces/{spaceId}",
			Summary:      "Get a space",
			Description:  "Returns a single space by ID with its member list.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "spaceId", Type: "int", Description: "Space ID"}},
			ResponseBody: SpaceResponse{},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Space not found or caller is not a member."},
			},
		},
	}
}

// updateSpaceRoute returns the PUT /spaces/{spaceId} route entry.
func updateSpaceRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "PUT",
			Path:         "/spaces/{spaceId}",
			Summary:      "Update a space",
			Description:  "Updates a space's name, color, icon or description. Requires owner or admin role.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "spaceId", Type: "int", Description: "Space ID"}},
			RequestBody:  UpdateSpaceRequest{},
			ResponseBody: SpaceResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body."},
				{Status: 403, Code: "permission_denied", Description: "Caller is not an owner or admin."},
				{Status: 404, Code: "not_found", Description: "Space not found."},
			},
		},
	}
}

// deleteSpaceRoute returns the DELETE /spaces/{spaceId} route entry.
func deleteSpaceRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:      "DELETE",
			Path:        "/spaces/{spaceId}",
			Summary:     "Delete a space",
			Description: "Deletes a space and reassigns or cascades its resources. Only the space owner can delete a space.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "spaceId", Type: "int", Description: "Space ID"}},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller is not the space owner."},
				{Status: 404, Code: "not_found", Description: "Space not found."},
			},
		},
	}
}

// leaveSpaceRoute returns the POST /spaces/{spaceId}/leave route entry.
func leaveSpaceRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:      "POST",
			Path:        "/spaces/{spaceId}/leave",
			Summary:     "Leave a space",
			Description: "Removes the calling user from a space. The sole owner cannot leave without transferring ownership first.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "spaceId", Type: "int", Description: "Space ID"}},
			Errors: []documentation.Error{
				{Status: 403, Code: "permission_denied", Description: "Caller is not a member."},
				{Status: 409, Code: "already_exists", Description: "Sole owner cannot leave, or Registre manages the space."},
			},
		},
	}
}

// syncRoute returns the POST /registre/sync route entry, mounted only when
// OIDC is configured.
func syncRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:  "POST",
			Path:    "/registre/sync",
			Summary: "Nudge a Registre sync",
			Description: "Registre posts here after any change to a shared space. The body is ignored and the call " +
				"needs no credentials: it only schedules an authenticated pull of the Registre snapshot. Answers 202. " +
				"A space Registre manages refuses rename, delete, member changes and leave with 409.",
			Status: 202,
		},
	}
}
