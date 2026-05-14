package fields

import (
	"context"
	stderrors "errors"
	"strconv"

	"api/internal/errors"
	"api/modules/documents"
	"api/schemas"

	"gorm.io/gorm"
)

var validFieldTypes = map[string]bool{
	"signature": true,
	"text":      true,
	"date":      true,
	"checkbox":  true,
}

type Service struct {
	orm        *gorm.DB
	docService *documents.Service
}

func NewService(orm *gorm.DB, docService *documents.Service) *Service {
	return &Service{orm: orm, docService: docService}
}

func (s *Service) List(ctx context.Context, ownerID string, docID string) ([]FieldResponse, error) {
	if _, err := s.docService.Get(ctx, ownerID, docID); err != nil {
		return nil, err
	}

	did, _ := strconv.ParseInt(docID, 10, 64)
	var records []schemas.Field
	if err := s.orm.WithContext(ctx).Where("document_id = ?", did).Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list fields", err)
	}

	out := make([]FieldResponse, len(records))
	for i := range records {
		out[i] = *toFieldResponse(&records[i])
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, ownerID string, docID string, req *CreateFieldRequest) (*FieldResponse, error) {
	doc, err := s.docService.Get(ctx, ownerID, docID)
	if err != nil {
		return nil, err
	}
	if doc.Status != "draft" {
		return nil, errors.Invalid("fields can only be added to draft documents")
	}
	if !validFieldTypes[req.FieldType] {
		return nil, errors.Invalid("field_type must be one of: signature, text, date, checkbox")
	}

	var signer schemas.Signer
	if err := s.orm.WithContext(ctx).Where("id = ? AND document_id = ?", req.SignerID, doc.ID).First(&signer).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("signer not found on this document")
		}
		return nil, errors.Internal("failed to verify signer", err)
	}

	record := &schemas.Field{
		DocumentID: doc.ID,
		SignerID:   req.SignerID,
		FieldType:  req.FieldType,
		Page:       req.Page,
		X:          req.X,
		Y:          req.Y,
		Width:      req.Width,
		Height:     req.Height,
		Required:   req.Required,
	}
	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, errors.Internal("failed to create field", err)
	}
	return toFieldResponse(record), nil
}

func (s *Service) Update(ctx context.Context, ownerID string, docID string, fieldID string, req *UpdateFieldRequest) (*FieldResponse, error) {
	doc, err := s.docService.Get(ctx, ownerID, docID)
	if err != nil {
		return nil, err
	}
	if doc.Status != "draft" {
		return nil, errors.Invalid("fields can only be updated on draft documents")
	}
	if !validFieldTypes[req.FieldType] {
		return nil, errors.Invalid("field_type must be one of: signature, text, date, checkbox")
	}

	fid, _ := strconv.ParseInt(fieldID, 10, 64)
	var record schemas.Field
	if err := s.orm.WithContext(ctx).Where("id = ? AND document_id = ?", fid, doc.ID).First(&record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("field not found")
		}
		return nil, errors.Internal("failed to read field", err)
	}

	record.FieldType = req.FieldType
	record.Page = req.Page
	record.X = req.X
	record.Y = req.Y
	record.Width = req.Width
	record.Height = req.Height
	record.Required = req.Required

	if err := s.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, errors.Internal("failed to update field", err)
	}
	return toFieldResponse(&record), nil
}

func (s *Service) Delete(ctx context.Context, ownerID string, docID string, fieldID string) error {
	doc, err := s.docService.Get(ctx, ownerID, docID)
	if err != nil {
		return err
	}
	if doc.Status != "draft" {
		return errors.Invalid("fields can only be deleted from draft documents")
	}

	fid, _ := strconv.ParseInt(fieldID, 10, 64)
	var record schemas.Field
	if err := s.orm.WithContext(ctx).Where("id = ? AND document_id = ?", fid, doc.ID).First(&record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("field not found")
		}
		return errors.Internal("failed to read field", err)
	}

	if err := s.orm.WithContext(ctx).Delete(&record).Error; err != nil {
		return errors.Internal("failed to delete field", err)
	}
	return nil
}

func toFieldResponse(record *schemas.Field) *FieldResponse {
	return &FieldResponse{
		ID:         record.ID,
		DocumentID: record.DocumentID,
		SignerID:   record.SignerID,
		FieldType:  record.FieldType,
		Page:       record.Page,
		X:          record.X,
		Y:          record.Y,
		Width:      record.Width,
		Height:     record.Height,
		Required:   record.Required,
		Value:      record.Value,
	}
}
