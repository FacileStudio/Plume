package spaces

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

var spaceID = []documentation.Field{{Name: "spaceId", Type: "string", Description: "Space ID"}}

var spaceMember = []documentation.Field{
	{Name: "spaceId", Type: "string", Description: "Space ID"},
	{Name: "memberId", Type: "string", Description: "Member ID"},
}

var Documentation = documentation.Module{
	Name:        "spaces",
	Description: "Shared workspaces and their members.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/spaces",
			Summary:      "Create a space",
			Auth:         "bearer",
			RequestBody:  "CreateSpaceRequest",
			ResponseBody: "SpaceResponse",
		},
		{
			Method:       "GET",
			Path:         "/spaces",
			Summary:      "List spaces",
			Description:  "Returns every space the authenticated user belongs to.",
			Auth:         "bearer",
			ResponseBody: "[]SpaceResponse",
		},
		{
			Method:       "GET",
			Path:         "/spaces/{spaceId}",
			Summary:      "Get a space",
			Auth:         "bearer",
			ResponseBody: "SpaceResponse",
			PathParams:   spaceID,
		},
		{
			Method:       "PUT",
			Path:         "/spaces/{spaceId}",
			Summary:      "Update a space",
			Auth:         "bearer",
			RequestBody:  "UpdateSpaceRequest",
			ResponseBody: "SpaceResponse",
			PathParams:   spaceID,
		},
		{
			Method:     "DELETE",
			Path:       "/spaces/{spaceId}",
			Summary:    "Delete a space",
			Auth:       "bearer",
			PathParams: spaceID,
		},
		{
			Method:      "POST",
			Path:        "/spaces/{spaceId}/leave",
			Summary:     "Leave a space",
			Description: "Removes the authenticated user from the space. Owners cannot leave their own space.",
			Auth:        "bearer",
			PathParams:  spaceID,
		},
		{
			Method:       "GET",
			Path:         "/spaces/{spaceId}/members",
			Summary:      "List space members",
			Auth:         "bearer",
			ResponseBody: "[]MemberResponse",
			PathParams:   spaceID,
		},
		{
			Method:       "POST",
			Path:         "/spaces/{spaceId}/members",
			Summary:      "Add a member",
			Auth:         "bearer",
			RequestBody:  "AddMemberRequest",
			ResponseBody: "MemberResponse",
			PathParams:   spaceID,
		},
		{
			Method:       "PUT",
			Path:         "/spaces/{spaceId}/members/{memberId}",
			Summary:      "Change a member's role",
			Auth:         "bearer",
			RequestBody:  "UpdateMemberRoleRequest",
			ResponseBody: "MemberResponse",
			PathParams:   spaceMember,
		},
		{
			Method:     "DELETE",
			Path:       "/spaces/{spaceId}/members/{memberId}",
			Summary:    "Remove a member",
			Auth:       "bearer",
			PathParams: spaceMember,
		},
	},
}
