package spaces

import (
	"context"
	stderrors "errors"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

func (s *Service) getMemberRole(ctx context.Context, spaceID int64, userID int64) (string, error) {
	var member schemas.SpaceMember
	err := s.orm.WithContext(ctx).Where("space_id = ? AND user_id = ?", spaceID, userID).First(&member).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return "", errors.Forbidden("you are not a member of this space")
	}
	if err != nil {
		return "", errors.Internal("failed to check membership", err)
	}
	role, err := normalizeRole(member.Role)
	if err != nil {
		return "", errors.Internal("the membership carries an unknown role", err)
	}
	return role, nil
}

func (s *Service) requireMinRole(ctx context.Context, spaceID int64, userID int64, minRole string) (string, error) {
	role, err := s.getMemberRole(ctx, spaceID, userID)
	if err != nil {
		return "", err
	}
	if roleLevel(role) < roleLevel(minRole) {
		return "", errors.Forbidden("insufficient permissions")
	}
	return role, nil
}

func roleLevel(role string) int {
	switch role {
	case RoleOwner:
		return 3
	case RoleAdmin:
		return 2
	case RoleMember:
		return 1
	default:
		return 0
	}
}

// normalizeRole refuses a role it does not know rather than answering
// RoleMember, so a corrupt role column is a refusal instead of a silent
// downgrade to a role that grants access. An absent role in a request body is
// the caller's business to default; this function is not told the difference.
func normalizeRole(role string) (string, error) {
	switch role {
	case RoleOwner, RoleAdmin, RoleMember:
		return role, nil
	default:
		return "", errors.Invalid("unknown role")
	}
}
