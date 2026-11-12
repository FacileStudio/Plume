package spaces

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
)

const testIssuer = "https://sso.test"

// linkIdentity gives a user a porte identity under the test issuer.
func linkIdentity(t *testing.T, userID int64, subject string) {
	t.Helper()
	if err := testDB.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)`,
		userID, testIssuer, subject).Error; err != nil {
		t.Fatalf("link identity: %v", err)
	}
}

// addMember puts a user in a space at a role.
func addMember(t *testing.T, spaceID, userID int64, role string) {
	t.Helper()
	if err := testDB.Create(&schemas.SpaceMember{SpaceID: spaceID, UserID: userID, Role: role}).Error; err != nil {
		t.Fatalf("add member: %v", err)
	}
}

// dumpSpaces renders every space and membership, timestamps included, so two
// dumps compare equal only when nothing was written in between.
func dumpSpaces(t *testing.T) string {
	t.Helper()
	var rows []string
	if err := testDB.Raw(`
		SELECT concat_ws('|', s.id, s.name, coalesce(s.registre_space_id, '-'), s.owner_id, s.updated_at,
		                 u.email, m.role, m.joined_at)
		  FROM spaces s JOIN space_members m ON m.space_id = s.id JOIN users u ON u.id = m.user_id
		 ORDER BY s.id, u.email`).Scan(&rows).Error; err != nil {
		t.Fatalf("dump: %v", err)
	}
	var users int64
	testDB.Model(&schemas.User{}).Count(&users)
	return fmt.Sprintf("%d users\n%s", users, strings.Join(rows, "\n"))
}

// A local space is linked by name ignoring case, a stranger is provisioned
// before ever signing in, a local-only member survives, an SSO member Registre
// dropped is removed, and the same snapshot applied twice writes nothing.
func TestApplyLinksProvisionsAndIsIdempotent(t *testing.T) {
	newService(t)
	ctx := context.Background()
	alice := createUser(t, "alice@facile.studio")
	bob := createUser(t, "bob@facile.studio")
	carol := createUser(t, "carol@facile.studio")
	dave := createUser(t, "dave@facile.studio")
	linkIdentity(t, alice, "sub-alice")
	linkIdentity(t, dave, "sub-dave")
	local := schemas.Space{Name: "Contracts", OwnerID: alice}
	if err := testDB.Create(&local).Error; err != nil {
		t.Fatalf("create space: %v", err)
	}
	addMember(t, local.ID, alice, RoleOwner)
	addMember(t, local.ID, carol, RoleMember)
	addMember(t, local.ID, dave, RoleAdmin)

	snapshot := directory.Snapshot{Version: 3, Spaces: []directory.Space{
		{ID: "rs-1", Name: "contracts", Members: []directory.Member{
			{Subject: "sub-alice", Email: "alice@facile.studio", Role: directory.RoleOwner},
			{Subject: "sub-bob", Email: "BOB@facile.studio", Role: directory.RoleAdmin},
			{Subject: "sub-erin", Email: "erin@facile.studio", Name: "Erin", Role: directory.RoleMember,
				Picture: testIssuer + "/avatars/sub-erin"},
		}},
		{ID: "rs-2", Name: "Fresh", Members: []directory.Member{
			{Subject: "sub-erin", Email: "erin@facile.studio", Role: directory.RoleOwner},
		}},
	}}
	registre := NewRegistre(testDB, directory.Client{Issuer: testIssuer}, nil, slog.Default())
	if err := registre.Apply(ctx, snapshot); err != nil {
		t.Fatalf("apply: %v", err)
	}

	var linked schemas.Space
	testDB.Where("registre_space_id = ?", "rs-1").First(&linked)
	if linked.ID != local.ID || linked.Name != "contracts" {
		t.Fatalf("rs-1 linked to %d %q, want %d renamed to contracts", linked.ID, linked.Name, local.ID)
	}
	want := map[string]string{"alice@facile.studio": RoleOwner, "bob@facile.studio": RoleAdmin,
		"erin@facile.studio": RoleMember, "carol@facile.studio": RoleMember}
	if got := rolesIn(t, local.ID); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("members = %v, want %v", got, want)
	}

	var erin schemas.User
	testDB.Where("email = ?", "erin@facile.studio").First(&erin)
	if erin.Name != "Erin" || erin.OIDCPictureURL != testIssuer+"/avatars/sub-erin" {
		t.Fatalf("erin provisioned as %+v", erin)
	}
	for subject, userID := range map[string]int64{"sub-bob": bob, "sub-erin": erin.ID} {
		var owner int64
		testDB.Raw(`SELECT user_id FROM porte_identities WHERE provider = ? AND subject = ?`, testIssuer, subject).Scan(&owner)
		if owner != userID {
			t.Fatalf("identity %s points at %d, want %d", subject, owner, userID)
		}
	}

	var fresh schemas.Space
	testDB.Where("registre_space_id = ?", "rs-2").First(&fresh)
	if fresh.OwnerID != erin.ID || rolesIn(t, fresh.ID)["erin@facile.studio"] != RoleOwner {
		t.Fatalf("rs-2 created as %+v with %v", fresh, rolesIn(t, fresh.ID))
	}

	before := dumpSpaces(t)
	if err := registre.Apply(ctx, snapshot); err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if after := dumpSpaces(t); after != before {
		t.Fatalf("second apply changed state:\n%s\nwant:\n%s", after, before)
	}
}

// rolesIn maps each member's email to their role in a space.
func rolesIn(t *testing.T, spaceID int64) map[string]string {
	t.Helper()
	var members []schemas.SpaceMember
	testDB.Preload("User").Where("space_id = ?", spaceID).Find(&members)
	out := map[string]string{}
	for _, member := range members {
		out[member.User.Email] = member.Role
	}
	return out
}

// A space Registre manages refuses local edits with 409, while an unlinked
// one stays editable and a new space is refused outright.
func TestLinkedSpacesAreManagedInRegistre(t *testing.T) {
	service := newService(t).ManagedBy(testIssuer)
	ctx := context.Background()
	owner := createUser(t, "alice@facile.studio")
	linked := "rs-1"
	managed := schemas.Space{Name: "Managed", OwnerID: owner, RegistreSpaceID: &linked}
	plain := schemas.Space{Name: "Plain", OwnerID: owner}
	testDB.Create(&managed)
	testDB.Create(&plain)
	addMember(t, managed.ID, owner, RoleOwner)
	addMember(t, plain.ID, owner, RoleOwner)

	if _, err := service.Create(ctx, owner, &CreateSpaceRequest{Name: "New"}); codeOf(err) != "already_exists" {
		t.Fatalf("create = %v, want a conflict", err)
	}
	if _, err := service.Update(ctx, owner, managed.ID, &UpdateSpaceRequest{Name: "x"}); codeOf(err) != "already_exists" {
		t.Fatalf("update linked = %v, want a conflict", err)
	}
	if err := service.Delete(ctx, owner, managed.ID); codeOf(err) != "already_exists" {
		t.Fatalf("delete linked = %v, want a conflict", err)
	}
	if _, err := service.Update(ctx, owner, plain.ID, &UpdateSpaceRequest{Name: "Renamed"}); err != nil {
		t.Fatalf("update unlinked: %v", err)
	}
	got, err := service.Get(ctx, owner, managed.ID)
	if err != nil || got.RegistreURL != testIssuer+"/spaces/rs-1" {
		t.Fatalf("get linked = %+v, %v", got, err)
	}
}
