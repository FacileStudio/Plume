package spaces

import (
	"context"
	stderrors "errors"
	"log/slog"
	"strings"

	"github.com/FacileStudio/Plume/apps/api/internal/oidcavatar"
	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"

	"gorm.io/gorm"
)

// resolvedMember is a snapshot member matched to a Plume user.
type resolvedMember struct {
	userID int64
	role   string
}

// Apply makes Plume's spaces match the snapshot, in one transaction, and is
// idempotent: applying the same snapshot twice writes nothing the second time.
//
// Each Registre space is linked to the Plume space carrying its id, else to
// the unlinked space with the same name ignoring case, else created. Members
// are upserted at their Registre role, and a member with an identity under
// the issuer who is absent from the snapshot is removed. A member without
// one, a local password account, is never touched, and neither is an
// unlinked space. A linked space missing from the snapshot is left alone and
// logged, because Registre has no space deletion.
func (r *Registre) Apply(ctx context.Context, snapshot directory.Snapshot) error {
	return r.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seen := make([]string, 0, len(snapshot.Spaces))
		for _, space := range snapshot.Spaces {
			if err := r.applySpace(tx, space); err != nil {
				return err
			}
			seen = append(seen, space.ID)
		}
		return r.reportMissing(tx, seen)
	})
}

// applySpace links one Registre space and sets its members.
func (r *Registre) applySpace(tx *gorm.DB, space directory.Space) error {
	members, ownerID, err := r.resolveMembers(tx, space.Members)
	if err != nil {
		return err
	}
	local, err := linkSpace(tx, space, ownerID)
	if err != nil {
		return err
	}
	keep := []int64{0}
	for _, member := range members {
		if err := tx.Exec(upsertMember, local.ID, member.userID, member.role).Error; err != nil {
			return err
		}
		keep = append(keep, member.userID)
	}
	return tx.Exec(removeAbsent, local.ID, r.issuer, keep).Error
}

const upsertMember = `
INSERT INTO space_members (space_id, user_id, role, joined_at) VALUES (?, ?, ?, now())
ON CONFLICT (space_id, user_id) DO UPDATE SET role = excluded.role
 WHERE space_members.role <> excluded.role
`

const removeAbsent = `
DELETE FROM space_members m
 USING porte_identities i
 WHERE m.space_id = ? AND i.user_id = m.user_id AND i.provider = ? AND m.user_id NOT IN ?
`

// linkSpace finds or creates the Plume space for a Registre space and sets its
// name, link and owner, writing only when one of them changed.
func linkSpace(tx *gorm.DB, space directory.Space, ownerID int64) (schemas.Space, error) {
	var local schemas.Space
	err := tx.Where("registre_space_id = ?", space.ID).First(&local).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		err = tx.Where("registre_space_id IS NULL AND lower(name) = lower(?)", space.Name).
			Order("id").First(&local).Error
	}
	if err != nil && !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return local, err
	}
	if ownerID == 0 {
		ownerID = local.OwnerID
	}
	if local.ID != 0 && local.RegistreSpaceID != nil && local.Name == space.Name && local.OwnerID == ownerID {
		return local, nil
	}
	local.Name, local.RegistreSpaceID, local.OwnerID = space.Name, &space.ID, ownerID
	return local, tx.Save(&local).Error
}

// resolveMembers matches every snapshot member to a Plume user, provisioning
// the ones Plume has never seen, and returns the first owner's user id. A
// member with an unknown role or no subject is skipped and logged.
func (r *Registre) resolveMembers(tx *gorm.DB, members []directory.Member) ([]resolvedMember, int64, error) {
	out := make([]resolvedMember, 0, len(members))
	for _, member := range members {
		role, err := normalizeRole(member.Role)
		if err != nil || member.Subject == "" {
			r.logger.Warn("registre: skipping a snapshot member", slog.String("sub", member.Subject))
			continue
		}
		userID, err := r.resolveUser(tx, member)
		if err != nil {
			return nil, 0, err
		}
		if userID != 0 {
			out = append(out, resolvedMember{userID: userID, role: role})
		}
	}
	return out, firstOwner(out), nil
}

// firstOwner is the user id of the first owner, or 0 when there is none.
func firstOwner(members []resolvedMember) int64 {
	for _, member := range members {
		if member.role == RoleOwner {
			return member.userID
		}
	}
	return 0
}

// resolveUser finds the user holding the member's identity under the issuer.
// Failing that it takes the user with the member's email, or provisions one,
// and records the identity, so the member's first sign-in lands on that user.
func (r *Registre) resolveUser(tx *gorm.DB, member directory.Member) (int64, error) {
	var userID int64
	err := tx.Raw(`SELECT user_id FROM porte_identities WHERE provider = ? AND subject = ?`,
		r.issuer, member.Subject).Scan(&userID).Error
	if err != nil || userID != 0 {
		return userID, err
	}
	userID, err = userByEmail(tx, member)
	if err != nil || userID == 0 {
		return userID, err
	}
	return userID, tx.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)
		ON CONFLICT (provider, subject) DO NOTHING`, userID, r.issuer, member.Subject).Error
}

// userByEmail returns the user with the member's email, creating it with the
// member's name and picture when there is none. A member without an email
// resolves to nobody.
func userByEmail(tx *gorm.DB, member directory.Member) (int64, error) {
	email := strings.ToLower(strings.TrimSpace(member.Email))
	if email == "" {
		return 0, nil
	}
	var user schemas.User
	err := tx.Where("lower(email) = ?", email).First(&user).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		user = schemas.User{Email: email, Name: member.Name, OIDCPictureURL: oidcavatar.PhotoURL(member.Picture)}
		err = tx.Create(&user).Error
	}
	return user.ID, err
}

// reportMissing logs the linked spaces the snapshot no longer carries.
func (r *Registre) reportMissing(tx *gorm.DB, seen []string) error {
	var names []string
	query := tx.Model(&schemas.Space{}).Where("registre_space_id IS NOT NULL")
	if len(seen) > 0 {
		query = query.Where("registre_space_id NOT IN ?", seen)
	}
	if err := query.Pluck("name", &names).Error; err != nil {
		return err
	}
	if len(names) > 0 {
		r.logger.Warn("registre: linked spaces missing from the snapshot, left alone", slog.Any("spaces", names))
	}
	return nil
}
