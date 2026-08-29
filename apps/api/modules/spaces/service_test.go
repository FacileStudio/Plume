package spaces

import (
	"context"
	stderrors "errors"
	"os"
	"testing"
	"time"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/testdb"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// Two owners leaving at the same instant must not both pass. The count and the
// delete are separate statements, so without a lock over the space's
// membership both transactions read two owners and the space ends with none —
// a state no route repairs, because Delete demands an owner, UpdateMemberRole
// refuses to assign one and AddMember refuses to add one.
//
// The interleave is forced rather than raced, since the window is
// sub-millisecond and two goroutines simply serialize. A held transaction takes
// the same rows the service takes and removes the other owner while holding
// them, so Leave either waits and re-reads one owner, or never looked.
func TestTwoOwnersCannotBothLeaveConcurrently(t *testing.T) {
	service := newService(t)
	ctx := context.Background()
	first := createUser(t, "noah@facile.studio")
	second := createUser(t, "iris@facile.studio")
	space := coOwnedSpace(t, service, first, second)

	tx := holdMembership(t, space.ID)
	defer func() { _ = tx.Rollback() }()

	started := make(chan struct{})
	left := make(chan error, 1)
	go func() {
		close(started)
		left <- service.Leave(ctx, second, space.ID)
	}()
	<-started
	waitForALockWait(t)

	if err := tx.Where("space_id = ? AND user_id = ?", space.ID, first).
		Delete(&schemas.SpaceMember{}).Error; err != nil {
		t.Fatalf("remove the first owner: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit the first owner's exit: %v", err)
	}

	leaveErr := <-left
	var owners int64
	if err := testDB.Model(&schemas.SpaceMember{}).
		Where("space_id = ? AND role = ?", space.ID, RoleOwner).Count(&owners).Error; err != nil {
		t.Fatalf("count the remaining owners: %v", err)
	}
	if owners == 0 {
		t.Fatalf("both owners left and the space is ownerless; the second leave returned %v", leaveErr)
	}
	if code := codeOf(leaveErr); code != "already_exists" {
		t.Fatalf("the second leave returned %q, want already_exists: %v", code, leaveErr)
	}
}

// coOwnedSpace is a space two accounts own equally, which is the only shape in
// which more than one member may leave.
func coOwnedSpace(t *testing.T, service *Service, first, second int64) *SpaceResponse {
	t.Helper()
	space, err := service.Create(context.Background(), first, &CreateSpaceRequest{Name: "Contracts"})
	if err != nil {
		t.Fatalf("create the space: %v", err)
	}
	coOwner := schemas.SpaceMember{SpaceID: space.ID, UserID: second, Role: RoleOwner}
	if err := testDB.Create(&coOwner).Error; err != nil {
		t.Fatalf("add the second owner: %v", err)
	}
	return space
}

// holdMembership opens a transaction standing in for the other leaver and takes
// the rows Leave has to take, so the caller decides when the two interleave.
func holdMembership(t *testing.T, spaceID int64) *gorm.DB {
	t.Helper()
	tx := testDB.Begin()
	if tx.Error != nil {
		t.Fatalf("open the competing transaction: %v", tx.Error)
	}
	var held []int64
	if err := tx.Model(&schemas.SpaceMember{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("space_id = ?", spaceID).Pluck("id", &held).Error; err != nil {
		t.Fatalf("hold the membership rows: %v", err)
	}
	return tx
}

// waitForALockWait gives the goroutine time to reach Postgres and then returns
// as soon as a backend is waiting on a lock. The floor is what makes the
// unlocked version fail rather than pass by accident: a Leave that takes no
// lock finishes in under a millisecond, so it is guaranteed to be done before
// the competing transaction commits.
func waitForALockWait(t *testing.T) {
	t.Helper()
	time.Sleep(100 * time.Millisecond)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var waiting int64
		if err := testDB.Raw(`SELECT count(*) FROM pg_stat_activity
			WHERE datname = current_database() AND wait_event_type = 'Lock'`).
			Scan(&waiting).Error; err != nil {
			t.Fatalf("read the lock waits: %v", err)
		}
		if waiting > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
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
