package signers

import (
	"context"
	stderrors "errors"
	"strconv"
	"time"

	"api/internal/errors"
	"api/modules/documents"
	"api/schemas"

	"gorm.io/gorm"
)

var validRoles = map[string]bool{
	"signer":   true,
	"viewer":   true,
	"approver": true,
}

type Service struct {
	orm        *gorm.DB
	docService *documents.Service
}

func NewService(orm *gorm.DB, docService *documents.Service) *Service {
	return &Service{orm: orm, docService: docService}
}

func (s *Service) ListSigners(ctx context.Context, ownerID string, docID string) ([]SignerResponse, error) {
	if _, err := s.docService.Get(ctx, ownerID, docID); err != nil {
		return nil, err
	}

	did, _ := strconv.ParseInt(docID, 10, 64)
	var records []schemas.Signer
	if err := s.orm.WithContext(ctx).Where("document_id = ?", did).Order("order_num asc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list signers", err)
	}

	out := make([]SignerResponse, len(records))
	for i := range records {
		out[i] = *toSignerResponse(&records[i])
	}
	return out, nil
}

func (s *Service) AddSigner(ctx context.Context, ownerID string, docID string, req *AddSignerRequest) (*SignerResponse, error) {
	doc, err := s.docService.Get(ctx, ownerID, docID)
	if err != nil {
		return nil, err
	}
	if doc.Status != "draft" {
		return nil, errors.Invalid("signers can only be added to draft documents")
	}
	if req.Name == "" {
		return nil, errors.Invalid("name is required")
	}
	if req.Email == "" {
		return nil, errors.Invalid("email is required")
	}
	role := req.Role
	if role == "" {
		role = "signer"
	}
	if !validRoles[role] {
		return nil, errors.Invalid("role must be one of: signer, viewer, approver")
	}

	record := &schemas.Signer{
		DocumentID: doc.ID,
		Name:       req.Name,
		Email:      req.Email,
		Role:       role,
		Status:     "pending",
		OrderNum:   req.Order,
	}
	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, errors.Internal("failed to add signer", err)
	}
	return toSignerResponse(record), nil
}

func (s *Service) RemoveSigner(ctx context.Context, ownerID string, signerID string) error {
	sid, _ := strconv.ParseInt(signerID, 10, 64)

	var signer schemas.Signer
	err := s.orm.WithContext(ctx).Where("id = ?", sid).First(&signer).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NotFound("signer not found")
	}
	if err != nil {
		return errors.Internal("failed to read signer", err)
	}

	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	var doc schemas.Document
	err = s.orm.WithContext(ctx).Where("id = ? AND owner_id = ?", signer.DocumentID, uid).First(&doc).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NotFound("signer not found")
	}
	if err != nil {
		return errors.Internal("failed to verify document ownership", err)
	}
	if doc.Status != "draft" {
		return errors.Invalid("signers can only be removed from draft documents")
	}

	if err := s.orm.WithContext(ctx).Delete(&signer).Error; err != nil {
		return errors.Internal("failed to remove signer", err)
	}
	return nil
}

func (s *Service) GetSigningView(ctx context.Context, token string) (*SigningView, error) {
	var signer schemas.Signer
	err := s.orm.WithContext(ctx).Where("token = ?", token).First(&signer).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("invalid signing link")
	}
	if err != nil {
		return nil, errors.Internal("failed to read signer", err)
	}

	doc, err := s.docService.FindByID(ctx, signer.DocumentID)
	if err != nil {
		return nil, err
	}
	if doc.Status != "pending" {
		return nil, errors.Invalid("document is not available for signing")
	}

	var fields []schemas.Field
	if err := s.orm.WithContext(ctx).Where("document_id = ? AND signer_id = ?", doc.ID, signer.ID).Find(&fields).Error; err != nil {
		return nil, errors.Internal("failed to load fields", err)
	}

	fieldResponses := make([]FieldResponse, len(fields))
	for i := range fields {
		fieldResponses[i] = *toFieldResponse(&fields[i])
	}

	return &SigningView{
		Document: DocumentInfo{
			ID:       doc.ID,
			Name:     doc.Name,
			FileName: doc.FileName,
			Status:   doc.Status,
		},
		Signer: *toSignerResponse(&signer),
		Fields: fieldResponses,
	}, nil
}

func (s *Service) SubmitSignature(ctx context.Context, token string, req *SubmitSignatureRequest, ipAddress string, userAgent string) error {
	var signer schemas.Signer
	err := s.orm.WithContext(ctx).Where("token = ?", token).First(&signer).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NotFound("invalid signing link")
	}
	if err != nil {
		return errors.Internal("failed to read signer", err)
	}
	if signer.Status != "pending" {
		return errors.Invalid("already signed or declined")
	}

	doc, err := s.docService.FindByID(ctx, signer.DocumentID)
	if err != nil {
		return err
	}
	if doc.Status != "pending" {
		return errors.Invalid("document is not available for signing")
	}

	for _, fv := range req.Fields {
		if err := s.orm.WithContext(ctx).Model(&schemas.Field{}).Where("id = ? AND signer_id = ?", fv.FieldID, signer.ID).Update("value", fv.Value).Error; err != nil {
			return errors.Internal("failed to save field value", err)
		}
	}

	now := time.Now().UTC()
	signer.Status = "signed"
	signer.SignedAt = &now
	signer.IPAddress = ipAddress
	signer.UserAgent = userAgent
	if err := s.orm.WithContext(ctx).Save(&signer).Error; err != nil {
		return errors.Internal("failed to update signer status", err)
	}

	var pendingCount int64
	s.orm.WithContext(ctx).Model(&schemas.Signer{}).Where("document_id = ? AND status = ? AND role != ?", doc.ID, "pending", "viewer").Count(&pendingCount)
	if pendingCount == 0 {
		if err := s.docService.UpdateStatus(ctx, doc.ID, "completed"); err != nil {
			return errors.Internal("failed to complete document", err)
		}
	}

	return nil
}

func (s *Service) DeclineSignature(ctx context.Context, token string, ipAddress string, userAgent string) error {
	var signer schemas.Signer
	err := s.orm.WithContext(ctx).Where("token = ?", token).First(&signer).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NotFound("invalid signing link")
	}
	if err != nil {
		return errors.Internal("failed to read signer", err)
	}
	if signer.Status != "pending" {
		return errors.Invalid("already signed or declined")
	}

	doc, err := s.docService.FindByID(ctx, signer.DocumentID)
	if err != nil {
		return err
	}
	if doc.Status != "pending" {
		return errors.Invalid("document is not available for signing")
	}

	now := time.Now().UTC()
	signer.Status = "declined"
	signer.SignedAt = &now
	signer.IPAddress = ipAddress
	signer.UserAgent = userAgent
	if err := s.orm.WithContext(ctx).Save(&signer).Error; err != nil {
		return errors.Internal("failed to update signer status", err)
	}

	if err := s.docService.UpdateStatus(ctx, doc.ID, "declined"); err != nil {
		return errors.Internal("failed to update document status", err)
	}

	return nil
}

func toSignerResponse(record *schemas.Signer) *SignerResponse {
	return &SignerResponse{
		ID:         record.ID,
		DocumentID: record.DocumentID,
		Name:       record.Name,
		Email:      record.Email,
		Role:       record.Role,
		Status:     record.Status,
		OrderNum:   record.OrderNum,
		SignedAt:    record.SignedAt,
		CreatedAt:  record.CreatedAt,
	}
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
