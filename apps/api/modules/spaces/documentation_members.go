package spaces

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

// memberRoutes returns the /spaces/{spaceId}/members route entries.
func memberRoutes() []documentation.Route {
	routes := listMembersRoute()
	routes = append(routes, addMemberRoute()...)
	routes = append(routes, updateMemberRoleRoute()...)
	routes = append(routes, removeMemberRoute()...)
	return routes
}

// listMembersRoute returns the GET /spaces/{spaceId}/members route entry.
func listMembersRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "GET",
			Path:         "/spaces/{spaceId}/members",
			Summary:      "List space members",
			Description:  "Returns all members of a space with their roles.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "spaceId", Type: "int", Description: "Space ID"}},
			ResponseBody: []MemberResponse{},
			Errors: []documentation.Error{
				{Status: 404, Code: "not_found", Description: "Space not found or caller is not a member."},
			},
		},
	}
}

// addMemberRoute returns the POST /spaces/{spaceId}/members route entry.
func addMemberRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:       "POST",
			Path:         "/spaces/{spaceId}/members",
			Summary:      "Add a member to a space",
			Description:  "Adds a user to the space by email. Requires owner or admin role.",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "spaceId", Type: "int", Description: "Space ID"}},
			RequestBody:  AddMemberRequest{},
			ResponseBody: MemberResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid email, invalid role, or user is already a member."},
				{Status: 403, Code: "permission_denied", Description: "Caller is not an owner or admin."},
				{Status: 404, Code: "not_found", Description: "Space or user not found."},
			},
		},
	}
}

// updateMemberRoleRoute returns the PUT /spaces/{spaceId}/members/{memberId}
// route entry.
func updateMemberRoleRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:      "PUT",
			Path:        "/spaces/{spaceId}/members/{memberId}",
			Summary:     "Update a member's role",
			Description: "Changes a space member's role (admin, member, viewer). Requires owner or admin role.",
			Auth:        "bearer",
			PathParams: []documentation.Field{
				{Name: "spaceId", Type: "int", Description: "Space ID"},
				{Name: "memberId", Type: "int", Description: "Member ID"},
			},
			RequestBody:  UpdateMemberRoleRequest{},
			ResponseBody: MemberResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid role or attempting to demote the sole owner."},
				{Status: 403, Code: "permission_denied", Description: "Caller is not an owner or admin."},
				{Status: 404, Code: "not_found", Description: "Space or member not found."},
			},
		},
	}
}

// removeMemberRoute returns the DELETE /spaces/{spaceId}/members/{memberId}
// route entry.
func removeMemberRoute() []documentation.Route {
	return []documentation.Route{
		{
			Method:      "DELETE",
			Path:        "/spaces/{spaceId}/members/{memberId}",
			Summary:     "Remove a member from a space",
			Description: "Removes a user from a space. Requires owner or admin role.",
			Auth:        "bearer",
			PathParams: []documentation.Field{
				{Name: "spaceId", Type: "int", Description: "Space ID"},
				{Name: "memberId", Type: "int", Description: "Member ID"},
			},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Cannot remove the space owner."},
				{Status: 403, Code: "permission_denied", Description: "Caller is not an owner or admin."},
				{Status: 404, Code: "not_found", Description: "Space or member not found."},
			},
		},
	}
}
