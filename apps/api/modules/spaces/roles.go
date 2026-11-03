package spaces

import (
	"context"
	stderrors "errors"
	"strings"

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

// ManagedBy hands this app's spaces to the Registre at issuer and returns the
// service. Without it every space stays editable here, which is what a Plume
// running without OIDC needs.
func (s *Service) ManagedBy(issuer string) *Service {
	s.registreURL = RegistreSpacesURL(issuer)
	return s
}

// RegistreSpacesURL is the page where the Registre at issuer manages spaces.
func RegistreSpacesURL(issuer string) string {
	return strings.TrimRight(issuer, "/") + "/spaces"
}

// requireUnlinked is requireMinRole that also refuses, with 409 and a pointer
// to Registre, a space Registre owns. The role check runs first so that a
// stranger learns nothing about a space.
func (s *Service) requireUnlinked(ctx context.Context, spaceID, userID int64, minRole string) (string, error) {
	role, err := s.requireMinRole(ctx, spaceID, userID, minRole)
	if err != nil || s.registreURL == "" {
		return role, err
	}
	var linked int64
	if err := s.orm.WithContext(ctx).Model(&schemas.Space{}).
		Where("id = ? AND registre_space_id IS NOT NULL", spaceID).Count(&linked).Error; err != nil {
		return "", errors.Internal("failed to check the space link", err)
	}
	if linked > 0 {
		return "", errors.Conflict("this space is managed in Registre: " + s.registreURL)
	}
	return role, nil
}
