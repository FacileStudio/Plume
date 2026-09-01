package spaces

import documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"

type DeleteSpaceResponse struct {
	Deleted bool `json:"deleted"`
}

type LeaveSpaceResponse struct {
	Left bool `json:"left"`
}

type RemoveMemberResponse struct {
	Removed bool `json:"removed"`
}

var Documentation = documentation.Module{
	Name:        "spaces",
	Description: "Collaborative space management and membership routes.",
	Routes: []documentation.Route{
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
		{
			Method:      "POST",
			Path:        "/spaces/{spaceId}/leave",
			Summary:     "Leave a space",
			Description: "Removes the calling user from a space. The sole owner cannot leave without transferring ownership first.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "spaceId", Type: "int", Description: "Space ID"}},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Sole owner cannot leave the space."},
				{Status: 404, Code: "not_found", Description: "Space not found or caller is not a member."},
			},
		},
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
	},
}
