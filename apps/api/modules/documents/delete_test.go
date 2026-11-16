package documents

import (
	"context"
	"strconv"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/schemas"
)

// TestDeleteHonoursTheStatus deletes a draft and a declined document, and keeps a
// sent (pending) and a completed one with an invalid_argument error.
func TestDeleteHonoursTheStatus(t *testing.T) {
	service := newService(t)
	owner := createUser(t, "owner@facile.studio")
	ownerID := strconv.FormatInt(owner, 10)
	for status, deletable := range map[string]bool{"draft": true, "declined": true, "pending": false, "completed": false} {
		doc := schemas.Document{Name: status, OwnerID: owner, Status: status}
		if err := testDB.Create(&doc).Error; err != nil {
			t.Fatalf("create the %s document: %v", status, err)
		}
		err := service.Delete(context.Background(), ownerID, strconv.FormatInt(doc.ID, 10))
		var left int64
		testDB.Model(&schemas.Document{}).Where("id = ?", doc.ID).Count(&left)
		switch {
		case deletable && (err != nil || left != 0):
			t.Fatalf("a %s document must be deleted: err=%v left=%d", status, err, left)
		case !deletable && (codeOf(err) != "invalid_argument" || left != 1):
			t.Fatalf("a %s document must be refused: err=%v left=%d", status, err, left)
		}
	}
}
