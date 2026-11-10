package signing

import (
	"context"
	"os"

	"github.com/FacileStudio/Plume/apps/api/schemas"
)

// FreezeAuditTrail writes a document's audit trail to disk when no current
// copy is there. The trail prints the owner's name and email from the users
// row, so erasure calls this before it deletes that row: the file then keeps
// what the record said at the time. A completed or declined document never
// changes again, so its cached trail is never regenerated.
func (s *Service) FreezeAuditTrail(ctx context.Context, docID int64) error {
	var doc schemas.Document
	if err := s.orm.WithContext(ctx).Where("id = ?", docID).First(&doc).Error; err != nil {
		return err
	}
	if info, err := os.Stat(s.auditTrailPath(docID)); err == nil && info.ModTime().After(doc.UpdatedAt) {
		return nil
	}
	return s.generateAuditTrail(docID, &doc)
}
