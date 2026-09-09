package documents

import (
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/Plume/apps/api/modules/smtp"
	"github.com/FacileStudio/Plume/apps/api/modules/spaces"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
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

	db, err := testdb.Open(url, testdb.Config{Prefix: "plume_documents_test", Migrate: schemas.Migrate})
	if err != nil {
		panic(err)
	}
	testDB = db
	os.Exit(m.Run())
}

// newService hands back the documents service over an empty schema, or skips
// loudly. Postgres is the only database this suite runs on.
func newService(t *testing.T) *Service {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testDB.Exec(`TRUNCATE users, spaces, space_members, documents RESTART IDENTITY CASCADE`).Error; err != nil {
		t.Fatalf("clear tables: %v", err)
	}
	return NewService(testDB, smtp.NewService(testDB), webhooks.NewService(testDB), "http://plume.test", "/tmp")
}

func createUser(t *testing.T, email string) int64 {
	t.Helper()
	user := schemas.User{Email: email, Name: email}
	if err := testDB.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user.ID
}

func codeOf(err error) string {
	var typed *errors.Error
	if stderrors.As(err, &typed) {
		return typed.Code
	}
	return ""
}

// seedSpace makes a space owned by alice holding a document uploaded by her,
// with bob as a member and carol out of it entirely.
func seedSpace(t *testing.T) (spaceID, owner, member, stranger int64) {
	t.Helper()
	owner = createUser(t, "alice@facile.studio")
	member = createUser(t, "bob@facile.studio")
	stranger = createUser(t, "carol@facile.studio")

	spaceService := spaces.NewService(testDB)
	space, err := spaceService.Create(context.Background(), owner, &spaces.CreateSpaceRequest{Name: "Contracts"})
	if err != nil {
		t.Fatalf("create the space: %v", err)
	}
	if err := testDB.Create(&schemas.SpaceMember{SpaceID: space.ID, UserID: member, Role: spaces.RoleAdmin}).Error; err != nil {
		t.Fatalf("add the member: %v", err)
	}
	return space.ID, owner, member, stranger
}

// The bug was that a space scoped a member's own documents and nothing else:
// only the uploader ever saw them, so an admin sharing the space saw an empty
// list. Anyone in the space must see every document it holds; outsiders must
// not.
func TestSpaceMemberSeesSharedDocuments(t *testing.T) {
	service := newService(t)
	ctx := context.Background()
	spaceID, owner, member, stranger := seedSpace(t)

	doc, err := service.Create(ctx, fmt.Sprintf("%d", owner), "Contract", "contract.pdf", fmt.Sprintf("%d", spaceID))
	if err != nil {
		t.Fatalf("create the document: %v", err)
	}

	got, err := service.List(ctx, fmt.Sprintf("%d", member), "", fmt.Sprintf("%d", spaceID), "")
	if err != nil {
		t.Fatalf("the member could not list the space: %v", err)
	}
	if len(got) != 1 || got[0].ID != doc.ID {
		t.Fatalf("the member saw %d documents, want the one owned by their teammate", len(got))
	}

	personal, err := service.List(ctx, fmt.Sprintf("%d", member), "", "", "")
	if err != nil {
		t.Fatalf("the member could not list their personal documents: %v", err)
	}
	if len(personal) != 0 {
		t.Fatalf("the space document leaked into the member's personal list")
	}

	if got, err := service.Get(ctx, fmt.Sprintf("%d", member), fmt.Sprintf("%d", doc.ID)); err != nil {
		t.Fatalf("the member could not open the shared document: %v", err)
	} else if got.ID != doc.ID {
		t.Fatalf("opened the wrong document: %v", got.ID)
	}

	_, err = service.List(ctx, fmt.Sprintf("%d", stranger), "", fmt.Sprintf("%d", spaceID), "")
	if code := codeOf(err); code != "permission_denied" {
		t.Fatalf("the stranger's listing returned %q, want permission_denied: %v", code, err)
	}

	_, err = service.Get(ctx, fmt.Sprintf("%d", stranger), fmt.Sprintf("%d", doc.ID))
	if code := codeOf(err); code != "permission_denied" {
		t.Fatalf("the stranger's read returned %q, want permission_denied: %v", code, err)
	}
}

// Stats must obey the same boundary, or the dashboard leaks counts the space
// view does not.
func TestSpaceMemberStatsAreSharedToo(t *testing.T) {
	service := newService(t)
	ctx := context.Background()
	spaceID, owner, member, stranger := seedSpace(t)

	if _, err := service.Create(ctx, fmt.Sprintf("%d", owner), "Contract", "contract.pdf", fmt.Sprintf("%d", spaceID)); err != nil {
		t.Fatalf("create the document: %v", err)
	}

	stats, err := service.Stats(ctx, fmt.Sprintf("%d", member), fmt.Sprintf("%d", spaceID))
	if err != nil {
		t.Fatalf("the member could not read the space stats: %v", err)
	}
	if stats.Total != 1 {
		t.Fatalf("the member's stats reported %d total, want 1", stats.Total)
	}

	_, err = service.Stats(ctx, fmt.Sprintf("%d", stranger), fmt.Sprintf("%d", spaceID))
	if code := codeOf(err); code != "permission_denied" {
		t.Fatalf("the stranger's stats returned %q, want permission_denied: %v", code, err)
	}
}

// Uploading into a space is a write across the same boundary, so it must be
// closed too, not just the reads.
func TestUploadingIntoASpaceRequiresMembership(t *testing.T) {
	service := newService(t)
	ctx := context.Background()
	spaceID, owner, _, stranger := seedSpace(t)

	if _, err := service.Create(ctx, fmt.Sprintf("%d", stranger), "Contract", "contract.pdf", fmt.Sprintf("%d", spaceID)); err == nil {
		t.Fatal("an outsider uploaded into the space")
	} else if code := codeOf(err); code != "permission_denied" {
		t.Fatalf("the outsider's upload returned %q, want permission_denied: %v", code, err)
	}

	got, err := service.List(ctx, fmt.Sprintf("%d", owner), "", fmt.Sprintf("%d", spaceID), "")
	if err != nil {
		t.Fatalf("the owner could not list the space: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("the rejected upload left %d documents behind", len(got))
	}
}