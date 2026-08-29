package spaces

import (
	"context"
	stderrors "errors"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Leave removes the caller's own membership. The last owner is refused, and
// the count and the delete share one transaction so that two owners leaving at
// once cannot both pass.
func (s *Service) Leave(ctx context.Context, userID int64, spaceID int64) error {
	return s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockMembers(tx, spaceID); err != nil {
			return errors.Internal("failed to lock the space membership", err)
		}
		return leaveLocked(tx, userID, spaceID)
	})
}

// lockMembers takes the space's membership rows for update, so that the owner
// count and the delete that follows it see the same set. Without it, two owners
// leaving at the same instant both count two owners, both pass, and the space
// ends with none — which no route can repair. SQLite has no row locks and
// serializes writes anyway, so the clause is Postgres-only.
func lockMembers(tx *gorm.DB, spaceID int64) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	var ids []int64
	return tx.Model(&schemas.SpaceMember{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("space_id = ?", spaceID).
		Pluck("id", &ids).Error
}

func leaveLocked(tx *gorm.DB, userID int64, spaceID int64) error {
	var member schemas.SpaceMember
	if err := tx.Where("space_id = ? AND user_id = ?", spaceID, userID).First(&member).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("you are not a member of this space")
		}
		return errors.Internal("failed to check membership", err)
	}

	if member.Role == RoleOwner {
		var owners int64
		if err := tx.Model(&schemas.SpaceMember{}).
			Where("space_id = ? AND role = ?", spaceID, RoleOwner).Count(&owners).Error; err != nil {
			return errors.Internal("failed to count the space owners", err)
		}
		if owners <= 1 {
			return errors.Conflict("the last owner cannot leave the space — transfer ownership or delete the space")
		}
	}

	if err := tx.Delete(&member).Error; err != nil {
		return errors.Internal("failed to leave space", err)
	}
	return nil
}
