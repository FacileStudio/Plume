package spaces

import (
	"context"
	stderrors "errors"
	"os"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/testdb"

	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	url, ok := testdb.URL()
	if !ok {
		testdb.Announce("sh scripts/postgres-up.sh")
		os.Exit(m.Run())
	}

	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_spaces_test", Migrate: schemas.Migrate})
	if err != nil {
		panic(err)
	}
	testDB = db
	os.Exit(m.Run())
}

// newService hands back the service over an empty users and spaces schema, or
// skips loudly. Postgres is the only database this suite runs on.
func newService(t *testing.T) *Service {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testDB.Exec(`TRUNCATE users, spaces, space_members RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear tables: %v", err)
	}
	return NewService(testDB)
}

// createUser inserts an account for a membership to hang off.
func createUser(t *testing.T, email string) int64 {
	t.Helper()
	user := schemas.User{Email: email, Name: email}
	if err := testDB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user.ID
}

// codeOf reports the tronc error code, which is what decides the HTTP status.
func codeOf(err error) string {
	var typed *errors.Error
	if stderrors.As(err, &typed) {
		return typed.Code
	}
	return ""
}

// Two owners are two exits. The guard that refused every owner left a space
// with two equal owners in a state where neither could leave, even though
// leaving would have left the space owned. Only the last owner is stuck, and
// that is a state conflict rather than a permission the caller lacks.
func TestOnlyTheLastOwnerCannotLeave(t *testing.T) {
	service := newService(t)
	ctx := context.Background()
	first := createUser(t, "noah@facile.studio")
	second := createUser(t, "iris@facile.studio")

	space, err := service.Create(ctx, first, &CreateSpaceRequest{Name: "Contracts"})
	if err != nil {
		t.Fatalf("create the space: %v", err)
	}
	coOwner := schemas.SpaceMember{SpaceID: space.ID, UserID: second, Role: RoleOwner}
	if err := testDB.Create(&coOwner).Error; err != nil {
		t.Fatalf("add the second owner: %v", err)
	}

	if err := service.Leave(ctx, first, space.ID); err != nil {
		t.Fatalf("one of two owners could not leave: %v", err)
	}

	err = service.Leave(ctx, second, space.ID)
	if err == nil {
		t.Fatal("the last owner left and the space is now ownerless")
	}
	if code := codeOf(err); code != "already_exists" {
		t.Fatalf("error code %q, want already_exists: %v", code, err)
	}
}

// A role column holding something nobody wrote on purpose is a refusal, not a
// downgrade. normalizeRole used to answer RoleMember for anything it did not
// recognise, so a corrupt value read as a valid role that grants access to the
// space's documents.
func TestAStoredUnknownRoleIsRefusedRatherThanDowngraded(t *testing.T) {
	service := newService(t)
	ctx := context.Background()
	owner := createUser(t, "noah@facile.studio")
	stranger := createUser(t, "iris@facile.studio")

	space, err := service.Create(ctx, owner, &CreateSpaceRequest{Name: "Contracts"})
	if err != nil {
		t.Fatalf("create the space: %v", err)
	}
	corrupt := schemas.SpaceMember{SpaceID: space.ID, UserID: stranger, Role: "superuser"}
	if err := testDB.Create(&corrupt).Error; err != nil {
		t.Fatalf("seed the corrupt membership: %v", err)
	}

	got, err := service.Get(ctx, stranger, space.ID)
	if err == nil {
		t.Fatalf("an unknown role was accepted and read back as %q", got.Role)
	}
	if code := codeOf(err); code != "internal" {
		t.Fatalf("error code %q, want internal: %v", code, err)
	}
}

// A role a request asks for is the caller's mistake rather than the database's,
// so it is a 400 and never a silent demotion to member.
func TestAnUnknownRequestedRoleIsRefused(t *testing.T) {
	service := newService(t)
	ctx := context.Background()
	owner := createUser(t, "noah@facile.studio")
	createUser(t, "iris@facile.studio")

	space, err := service.Create(ctx, owner, &CreateSpaceRequest{Name: "Contracts"})
	if err != nil {
		t.Fatalf("create the space: %v", err)
	}

	_, err = service.AddMember(ctx, owner, space.ID, &AddMemberRequest{
		Email: "iris@facile.studio", Role: "superuser",
	})
	if code := codeOf(err); code != "invalid_argument" {
		t.Fatalf("error code %q, want invalid_argument: %v", code, err)
	}

	member, err := service.AddMember(ctx, owner, space.ID, &AddMemberRequest{Email: "iris@facile.studio"})
	if err != nil {
		t.Fatalf("an absent role is still the member default: %v", err)
	}
	if member.Role != RoleMember {
		t.Fatalf("an absent role became %q", member.Role)
	}
}
