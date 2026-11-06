package signers

import (
	"context"
	"os"
	"time"

	"github.com/FacileStudio/Plume/apps/api/internal/nuage"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
	"github.com/FacileStudio/Plume/apps/api/schemas"
)

// publishToNuage queues the final PDF for archiving in the owner's Plume
// folder in Nuage. It never fails the caller: Nuage being down must not undo a
// completed signature.
func (s *Service) publishToNuage(doc *schemas.Document) {
	if s.publisher == nil {
		return
	}
	job := nuageJob(doc)
	job.Load = func() ([]byte, error) {
		path, err := s.docService.GetFilePathByDocID(context.Background(), doc.ID)
		if err != nil {
			return nil, err
		}
		return os.ReadFile(path)
	}
	job.OnPublished = func(linkedTo string) {
		s.recordArchived(doc, linkedTo)
	}
	s.publisher.Enqueue(job)
}

// nuageJob describes the publish for a completed document: the owner whose
// folder receives it and the linked_to key the caller asked for, if any.
func nuageJob(doc *schemas.Document) nuage.Job {
	return nuage.Job{DocumentID: doc.ID, OwnerID: doc.OwnerID, Title: doc.Name, LinkedTo: doc.NuageLinkedTo}
}

func (s *Service) recordArchived(doc *schemas.Document, linkedTo string) {
	now := time.Now().UTC()
	if err := s.orm.Model(&schemas.Document{}).Where("id = ?", doc.ID).Update("nuage_published_at", now).Error; err != nil {
		return
	}
	archived := *doc
	archived.NuagePublishedAt = &now
	s.webhookSvc.Dispatch(doc.OwnerID, webhooks.BuildArchivedEvent(&archived, s.domain, doc.NuageSpaceFacileID, linkedTo))
}
