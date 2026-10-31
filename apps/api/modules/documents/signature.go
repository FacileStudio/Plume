package documents

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/FacileStudio/Plume/apps/api/internal/pdfutil"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// CreateSignatureRequest creates a document from a PDF, adds its signers and
// fields and, unless req.Send is false, sends it, all in one call. The bool
// reports whether a document was created: a request whose external_ref is
// already taken by the same owner returns the existing document untouched.
func (s *Service) CreateSignatureRequest(ctx context.Context, ownerID int64, req *SignatureRequest, fileName string, file io.Reader) (*SignatureResponse, bool, error) {
	if err := req.validate(); err != nil {
		return nil, false, err
	}
	if existing, err := s.findByExternalRef(ctx, ownerID, req.ExternalRef); err != nil || existing != nil {
		return s.existingResponse(ctx, existing, err)
	}
	doc, err := s.createLinkedDraft(ctx, ownerID, req, fileName)
	if err != nil {
		if existing, _ := s.findByExternalRef(ctx, ownerID, req.ExternalRef); existing != nil {
			return s.existingResponse(ctx, existing, nil)
		}
		return nil, false, err
	}
	if err := s.fillSignatureRequest(ctx, doc, req, fileName, file); err != nil {
		s.discard(ctx, doc)
		return nil, false, err
	}
	resp, err := s.finishSignatureRequest(ctx, doc, req)
	if err != nil {
		s.discard(ctx, doc)
		return nil, false, err
	}
	return resp, true, nil
}

func (s *Service) createLinkedDraft(ctx context.Context, ownerID int64, req *SignatureRequest, fileName string) (*schemas.Document, error) {
	record := &schemas.Document{
		Name:               req.Name,
		FileName:           fileName,
		Status:             "draft",
		OwnerID:            ownerID,
		Sequential:         req.Sequential,
		SpaceID:            req.SpaceID,
		ExternalRef:        req.ExternalRef,
		NuageSpaceFacileID: req.NuageSpaceFacileID,
		NuageLinkedTo:      req.NuageLinkedTo,
	}
	if req.SpaceID != nil {
		if err := s.requireSpaceMember(ctx, *req.SpaceID, ownerID); err != nil {
			return nil, err
		}
	}
	if req.Client != nil {
		clientID, err := s.clientFor(ctx, ownerID, req.Client)
		if err != nil {
			return nil, err
		}
		record.ClientID = &clientID
	}
	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, errors.Internal("failed to create document", err)
	}
	return record, nil
}

func (s *Service) fillSignatureRequest(ctx context.Context, doc *schemas.Document, req *SignatureRequest, fileName string, file io.Reader) error {
	if _, err := s.storeFile(ctx, file, fileName, doc.OwnerID, doc.ID); err != nil {
		return err
	}
	pages, err := s.pageCount(ctx, doc.ID)
	if err != nil {
		return err
	}
	return s.insertSignersAndFields(ctx, doc.ID, req, pages)
}

func (s *Service) pageCount(ctx context.Context, docID int64) (int, error) {
	stored, err := s.FindByID(ctx, docID)
	if err != nil {
		return 0, err
	}
	pages, err := pdfutil.PageCount(filepath.Join(s.uploadDir, stored.StoragePath))
	if err != nil || pages < 1 {
		return 0, errors.Invalid("uploaded file is not a readable PDF")
	}
	return pages, nil
}

func (s *Service) finishSignatureRequest(ctx context.Context, doc *schemas.Document, req *SignatureRequest) (*SignatureResponse, error) {
	ownerID := strconv.FormatInt(doc.OwnerID, 10)
	go s.webhookSvc.Dispatch(doc.OwnerID, webhooks.BuildDocumentEvent(webhooks.EventDocumentCreated, doc, s.domain))
	if req.shouldSend() {
		if _, err := s.Send(ctx, ownerID, strconv.FormatInt(doc.ID, 10)); err != nil {
			return nil, err
		}
	}
	fresh, err := s.FindByID(ctx, doc.ID)
	if err != nil {
		return nil, err
	}
	return s.signatureResponse(ctx, fresh)
}

func (s *Service) existingResponse(ctx context.Context, doc *schemas.Document, err error) (*SignatureResponse, bool, error) {
	if err != nil {
		return nil, false, err
	}
	resp, err := s.signatureResponse(ctx, doc)
	return resp, false, err
}

func (s *Service) discard(ctx context.Context, doc *schemas.Document) {
	if stored, err := s.FindByID(ctx, doc.ID); err == nil && stored.StoragePath != "" {
		os.Remove(filepath.Join(s.uploadDir, stored.StoragePath))
	}
	s.orm.WithContext(ctx).Where("document_id = ?", doc.ID).Delete(&schemas.Field{})
	s.orm.WithContext(ctx).Where("document_id = ?", doc.ID).Delete(&schemas.Signer{})
	s.orm.WithContext(ctx).Delete(&schemas.Document{}, doc.ID)
}
