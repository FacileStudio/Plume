package spaces

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	portepg "github.com/FacileStudio/porte/pg"
)

func joinAt(t *testing.T, spaceID, userID int64, role string, daysAgo int) {
	t.Helper()
	member := schemas.SpaceMember{SpaceID: spaceID, UserID: userID, Role: role, JoinedAt: time.Now().AddDate(0, 0, -daysAgo)}
	mustCreate(t, &member)
}

func ownerOf(t *testing.T, name string) int64 {
	t.Helper()
	return count(t, `SELECT owner_id FROM documents WHERE name = ?`, name)
}

func TestErasePersonPassesSpaceDocumentsToTheLongestStandingAdmin(t *testing.T) {
	f := seedErasure(t)
	oldMember := createUser(t, "old-member@facile.studio")
	youngAdmin := createUser(t, "young-admin@facile.studio")
	elderAdmin := createUser(t, "elder-admin@facile.studio")
	onlyMember := createUser(t, "only-member@facile.studio")

	withAdmins := schemas.Space{Name: "Admins", OwnerID: f.gone}
	mustCreate(t, &withAdmins)
	joinAt(t, withAdmins.ID, f.gone, "owner", 100)
	joinAt(t, withAdmins.ID, oldMember, "member", 90)
	joinAt(t, withAdmins.ID, youngAdmin, "admin", 5)
	joinAt(t, withAdmins.ID, elderAdmin, "admin", 30)

	withMembers := schemas.Space{Name: "Members", OwnerID: f.gone}
	mustCreate(t, &withMembers)
	joinAt(t, withMembers.ID, f.gone, "owner", 100)
	joinAt(t, withMembers.ID, onlyMember, "member", 10)
	joinAt(t, withMembers.ID, f.other, "member", 2)

	document(t, f.gone, &withAdmins.ID, "to-admin", "1/10_x.pdf")
	document(t, f.gone, &withMembers.ID, "to-member", "1/11_y.pdf")
	mustCreate(t, &schemas.Document{Name: "taken", OwnerID: elderAdmin, SpaceID: &withAdmins.ID, ExternalRef: "ref-1"})
	mustCreate(t, &schemas.Document{Name: "clash", OwnerID: f.gone, SpaceID: &withAdmins.ID, ExternalRef: "ref-1"})

	for range 2 {
		if err := f.deletions.ErasePerson(context.Background(), testIssuer, "sub-gone"); err != nil {
			t.Fatalf("erase: %v", err)
		}
	}

	for name, want := range map[string]int64{"to-admin": elderAdmin, "clash": elderAdmin, "to-member": onlyMember} {
		if got := ownerOf(t, name); got != want {
			t.Errorf("document %q owner = %d, want %d", name, got, want)
		}
	}
	if n := count(t, `SELECT count(*) FROM documents WHERE name = 'clash' AND external_ref = ''`); n != 1 {
		t.Errorf("a clashing external_ref was kept instead of cleared")
	}
	if n := count(t, `SELECT count(*) FROM documents WHERE name = 'taken' AND external_ref = 'ref-1'`); n != 1 {
		t.Errorf("the successor's own external_ref changed")
	}
	for name, want := range map[string]int64{"Admins": elderAdmin, "Members": onlyMember} {
		if got := count(t, `SELECT owner_id FROM spaces WHERE name = ?`, name); got != want {
			t.Errorf("space %q owner = %d, want %d", name, got, want)
		}
	}
}

func TestEraseIdentityRunsAfterTheAccountIsGone(t *testing.T) {
	f := seedErasure(t)
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	erase := portepg.New(sqlDB).EraseIdentity
	if err := f.deletions.ErasePerson(context.Background(), testIssuer, "sub-gone"); err != nil {
		t.Fatalf("erase person: %v", err)
	}
	if err := erase(context.Background(), testIssuer, "sub-gone"); err != nil {
		t.Fatalf("erase identity: %v", err)
	}
	linkIdentity(t, f.other, "sub-extra")
	if err := erase(context.Background(), testIssuer, "sub-extra"); err != nil {
		t.Fatalf("erase a live identity needs porte_magic_links: %v", err)
	}
}
