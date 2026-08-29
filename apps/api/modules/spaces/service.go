package spaces

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Service implements the spaces CRUD and membership and role management.
type Service struct {
	orm *gorm.DB
}

// NewService wires the spaces service to its database.
func NewService(orm *gorm.DB) *Service {
	return &Service{orm: orm}
}

func (s *Service) Create(ctx context.Context, ownerID int64, req *CreateSpaceRequest) (*SpaceResponse, error) {
	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}

	space := &schemas.Space{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
	}
	if err := s.orm.WithContext(ctx).Create(space).Error; err != nil {
		return nil, errors.Internal("failed to create space", err)
	}

	member := &schemas.SpaceMember{
		SpaceID: space.ID,
		UserID:  ownerID,
		Role:    RoleOwner,
	}
	if err := s.orm.WithContext(ctx).Create(member).Error; err != nil {
		return nil, errors.Internal("failed to add owner as member", err)
	}

	return toSpaceResponse(space, RoleOwner), nil
}

func (s *Service) List(ctx context.Context, userID int64) ([]SpaceResponse, error) {
	var members []schemas.SpaceMember
	if err := s.orm.WithContext(ctx).Where("user_id = ?", userID).Find(&members).Error; err != nil {
		return nil, errors.Internal("failed to list spaces", err)
	}

	if len(members) == 0 {
		return []SpaceResponse{}, nil
	}

	spaceIDs := make([]int64, len(members))
	roleMap := make(map[int64]string, len(members))
	for i, m := range members {
		spaceIDs[i] = m.SpaceID
		roleMap[m.SpaceID] = m.Role
	}

	var spaces []schemas.Space
	if err := s.orm.WithContext(ctx).Where("id IN ?", spaceIDs).Order("created_at desc").Find(&spaces).Error; err != nil {
		return nil, errors.Internal("failed to load spaces", err)
	}

	out := make([]SpaceResponse, len(spaces))
	for i := range spaces {
		out[i] = *toSpaceResponse(&spaces[i], roleMap[spaces[i].ID])
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, userID int64, spaceID int64) (*SpaceResponse, error) {
	role, err := s.getMemberRole(ctx, spaceID, userID)
	if err != nil {
		return nil, err
	}

	var space schemas.Space
	if err := s.orm.WithContext(ctx).Where("id = ?", spaceID).First(&space).Error; err != nil {
		return nil, errors.NotFound("space not found")
	}

	return toSpaceResponse(&space, role), nil
}

func (s *Service) Update(ctx context.Context, userID int64, spaceID int64, req *UpdateSpaceRequest) (*SpaceResponse, error) {
	role, err := s.requireMinRole(ctx, spaceID, userID, RoleAdmin)
	if err != nil {
		return nil, err
	}

	var space schemas.Space
	if err := s.orm.WithContext(ctx).Where("id = ?", spaceID).First(&space).Error; err != nil {
		return nil, errors.NotFound("space not found")
	}

	if req.Name != "" {
		space.Name = req.Name
	}
	space.Description = req.Description

	if err := s.orm.WithContext(ctx).Save(&space).Error; err != nil {
		return nil, errors.Internal("failed to update space", err)
	}

	return toSpaceResponse(&space, role), nil
}

func (s *Service) Delete(ctx context.Context, userID int64, spaceID int64) error {
	if _, err := s.requireMinRole(ctx, spaceID, userID, RoleOwner); err != nil {
		return err
	}

	if err := s.orm.WithContext(ctx).Where("id = ?", spaceID).Delete(&schemas.Space{}).Error; err != nil {
		return errors.Internal("failed to delete space", err)
	}
	return nil
}

func (s *Service) ListMembers(ctx context.Context, userID int64, spaceID int64) ([]MemberResponse, error) {
	if _, err := s.getMemberRole(ctx, spaceID, userID); err != nil {
		return nil, err
	}

	var members []schemas.SpaceMember
	if err := s.orm.WithContext(ctx).Preload("User").Where("space_id = ?", spaceID).Order("joined_at asc").Find(&members).Error; err != nil {
		return nil, errors.Internal("failed to list members", err)
	}

	out := make([]MemberResponse, len(members))
	for i := range members {
		out[i] = toMemberResponse(&members[i])
	}
	return out, nil
}

func (s *Service) AddMember(ctx context.Context, userID int64, spaceID int64, req *AddMemberRequest) (*MemberResponse, error) {
	if _, err := s.requireMinRole(ctx, spaceID, userID, RoleAdmin); err != nil {
		return nil, err
	}

	if req.Email == "" {
		return nil, errors.Invalid("email is required")
	}
	role := RoleMember
	if req.Role != "" {
		requested, err := normalizeRole(req.Role)
		if err != nil {
			return nil, err
		}
		role = requested
	}
	if role == RoleOwner {
		return nil, errors.Invalid("cannot add a member as owner")
	}

	var target schemas.User
	if err := s.orm.WithContext(ctx).Where("email = ?", req.Email).First(&target).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Internal("failed to find user", err)
	}

	var existing schemas.SpaceMember
	err := s.orm.WithContext(ctx).Where("space_id = ? AND user_id = ?", spaceID, target.ID).First(&existing).Error
	if err == nil {
		return nil, errors.Conflict("user is already a member of this space")
	}

	member := &schemas.SpaceMember{
		SpaceID: spaceID,
		UserID:  target.ID,
		Role:    role,
	}
	if err := s.orm.WithContext(ctx).Create(member).Error; err != nil {
		return nil, errors.Internal("failed to add member", err)
	}
	member.User = target

	resp := toMemberResponse(member)
	return &resp, nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, userID int64, spaceID int64, memberID int64, req *UpdateMemberRoleRequest) (*MemberResponse, error) {
	if _, err := s.requireMinRole(ctx, spaceID, userID, RoleAdmin); err != nil {
		return nil, err
	}

	role, err := normalizeRole(req.Role)
	if err != nil {
		return nil, err
	}
	if role == RoleOwner {
		return nil, errors.Invalid("cannot assign owner role")
	}

	var member schemas.SpaceMember
	if err := s.orm.WithContext(ctx).Preload("User").Where("id = ? AND space_id = ?", memberID, spaceID).First(&member).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("member not found")
		}
		return nil, errors.Internal("failed to find member", err)
	}

	if member.Role == RoleOwner {
		return nil, errors.Forbidden("cannot change the owner's role")
	}

	member.Role = role
	if err := s.orm.WithContext(ctx).Save(&member).Error; err != nil {
		return nil, errors.Internal("failed to update member role", err)
	}

	resp := toMemberResponse(&member)
	return &resp, nil
}

func (s *Service) RemoveMember(ctx context.Context, userID int64, spaceID int64, memberID int64) error {
	if _, err := s.requireMinRole(ctx, spaceID, userID, RoleAdmin); err != nil {
		return err
	}

	var member schemas.SpaceMember
	if err := s.orm.WithContext(ctx).Where("id = ? AND space_id = ?", memberID, spaceID).First(&member).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("member not found")
		}
		return errors.Internal("failed to find member", err)
	}

	if member.Role == RoleOwner {
		return errors.Forbidden("cannot remove the space owner")
	}

	if err := s.orm.WithContext(ctx).Delete(&member).Error; err != nil {
		return errors.Internal("failed to remove member", err)
	}
	return nil
}

func (s *Service) Leave(ctx context.Context, userID int64, spaceID int64) error {
	var member schemas.SpaceMember
	if err := s.orm.WithContext(ctx).Where("space_id = ? AND user_id = ?", spaceID, userID).First(&member).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("you are not a member of this space")
		}
		return errors.Internal("failed to check membership", err)
	}

	if member.Role == RoleOwner {
		var owners int64
		if err := s.orm.WithContext(ctx).Model(&schemas.SpaceMember{}).
			Where("space_id = ? AND role = ?", spaceID, RoleOwner).Count(&owners).Error; err != nil {
			return errors.Internal("failed to count the space owners", err)
		}
		if owners <= 1 {
			return errors.Conflict("the last owner cannot leave the space — transfer ownership or delete the space")
		}
	}

	if err := s.orm.WithContext(ctx).Delete(&member).Error; err != nil {
		return errors.Internal("failed to leave space", err)
	}
	return nil
}

func toSpaceResponse(s *schemas.Space, role string) *SpaceResponse {
	return &SpaceResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		OwnerID:     s.OwnerID,
		Role:        role,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}
}

func toMemberResponse(m *schemas.SpaceMember) MemberResponse {
	return MemberResponse{
		ID:       m.ID,
		UserID:   m.UserID,
		Email:    m.User.Email,
		Name:     m.User.Name,
		Role:     m.Role,
		JoinedAt: m.JoinedAt.Format(time.RFC3339),
	}
}
