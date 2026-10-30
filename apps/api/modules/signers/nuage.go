package signers

import (
	"context"
	"os"
	"time"

	"github.com/FacileStudio/Plume/apps/api/internal/nuage"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
	"github.com/FacileStudio/Plume/apps/api/schemas"
)

// publishToNuage queues the final PDF for archiving in Nuage. It never fails
// the caller: Nuage being down must not undo a completed signature.
func (s *Service) publishToNuage(ctx context.Context, doc *schemas.Document) {
	if s.publisher == nil {
		return
	}
	var space *schemas.Space
	if doc.SpaceID != nil {
		var found schemas.Space
		if err := s.orm.WithContext(ctx).First(&found, *doc.SpaceID).Error; err == nil {
			space = &found
		}
	}
	job := nuageJob(doc, space)
	job.Load = func() ([]byte, error) {
		path, err := s.docService.GetFilePathByDocID(context.Background(), doc.ID)
		if err != nil {
			return nil, err
		}
		return os.ReadFile(path)
	}
	job.OnPublished = func(spaceFacileID, linkedTo string) {
		s.recordArchived(doc, spaceFacileID, linkedTo)
	}
	s.publisher.Enqueue(job)
}

// nuageJob selects the publish target for a completed document. The space
// and the linked_to key the caller asked for win; each falls back on its own
// to the Plume space (then the publisher's default space) and to
// plume:document:<id>.
func nuageJob(doc *schemas.Document, space *schemas.Space) nuage.Job {
	job := nuage.Job{DocumentID: doc.ID, Title: doc.Name, LinkedTo: doc.NuageLinkedTo}
	switch {
	case doc.NuageSpaceFacileID != "":
		job.SpaceFacileID = doc.NuageSpaceFacileID
	case space != nil:
		job.SpaceFacileID = space.FacileID
		job.SpaceName = space.Name
	}
	return job
}

func (s *Service) recordArchived(doc *schemas.Document, spaceFacileID, linkedTo string) {
	now := time.Now().UTC()
	if err := s.orm.Model(&schemas.Document{}).Where("id = ?", doc.ID).Update("nuage_published_at", now).Error; err != nil {
		return
	}
	archived := *doc
	archived.NuagePublishedAt = &now
	s.webhookSvc.Dispatch(doc.OwnerID, webhooks.BuildArchivedEvent(&archived, s.domain, spaceFacileID, linkedTo))
}
