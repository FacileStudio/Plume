package spaces

// CreateSpaceRequest is the payload for creating a space.
type CreateSpaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateSpaceRequest is the payload for updating a space's name or
// description.
type UpdateSpaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AddMemberRequest is the payload for adding a member to a space.
type AddMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// UpdateMemberRoleRequest is the payload for changing a member's role.
type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}

// SpaceResponse is the API representation of a space and the caller's role in
// it.
type SpaceResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     int64  `json:"owner_id"`
	Role        string `json:"role"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// MemberResponse is the API representation of a space membership.
type MemberResponse struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)
