package documents

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	stderrors "errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"api/internal/errors"
	"api/internal/pdfutil"
	"api/modules/smtp"
	"api/modules/webhooks"
	"api/schemas"

	"gorm.io/gorm"
)

type Service struct {
	orm        *gorm.DB
	smtp       *smtp.Service
	webhookSvc *webhooks.Service
	domain     string
	uploadDir  string
}

func NewService(orm *gorm.DB, smtpService *smtp.Service, webhookSvc *webhooks.Service, domain string, uploadDir string) *Service {
	return &Service{orm: orm, smtp: smtpService, webhookSvc: webhookSvc, domain: domain, uploadDir: uploadDir}
}

func (s *Service) Create(ctx context.Context, ownerID string, name string, fileName string) (*DocumentResponse, error) {
	if name == "" {
		return nil, errors.Invalid("name is required")
	}

	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	record := &schemas.Document{
		Name:     name,
		FileName: fileName,
		Status:   "draft",
		OwnerID:  uid,
	}
	if err := s.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, errors.Internal("failed to create document", err)
	}
	return toResponse(record), nil
}

func (s *Service) UpdateStoragePath(ctx context.Context, docID int64, path string) error {
	err := s.orm.WithContext(ctx).Model(&schemas.Document{}).Where("id = ?", docID).Update("storage_path", path).Error
	if err != nil {
		return errors.Internal("failed to update storage path", err)
	}
	return nil
}

func (s *Service) GetFilePath(ctx context.Context, ownerID string, docID string) (string, error) {
	record, err := s.findOwned(ctx, ownerID, docID)
	if err != nil {
		return "", err
	}
	if record.StoragePath == "" {
		return "", errors.NotFound("no file uploaded for this document")
	}
	originalPath := filepath.Join(s.uploadDir, record.StoragePath)
	if record.Status == "completed" {
		return s.getOrCreateSignedFile(ctx, record.ID, originalPath)
	}
	return originalPath, nil
}

func (s *Service) GetFilePathByDocID(ctx context.Context, docID int64) (string, error) {
	record, err := s.FindByID(ctx, docID)
	if err != nil {
		return "", err
	}
	if record.StoragePath == "" {
		return "", errors.NotFound("no file uploaded for this document")
	}
	originalPath := filepath.Join(s.uploadDir, record.StoragePath)
	if record.Status == "completed" {
		return s.getOrCreateSignedFile(ctx, record.ID, originalPath)
	}
	return originalPath, nil
}

func (s *Service) getOrCreateSignedFile(ctx context.Context, docID int64, originalPath string) (string, error) {
	signedPath := strings.TrimSuffix(originalPath, ".pdf") + "_signed.pdf"
	if _, err := os.Stat(signedPath); err == nil {
		return signedPath, nil
	}

	var fields []schemas.Field
	if err := s.orm.WithContext(ctx).Where("document_id = ? AND value != ''", docID).Find(&fields).Error; err != nil {
		return "", errors.Internal("failed to load fields", err)
	}

	overlays := make([]pdfutil.FieldOverlay, len(fields))
	for i, f := range fields {
		overlays[i] = pdfutil.FieldOverlay{
			Page:      f.Page,
			X:         f.X,
			Y:         f.Y,
			Width:     f.Width,
			Height:    f.Height,
			FieldType: f.FieldType,
			Value:     f.Value,
		}
	}

	if err := pdfutil.FlattenFields(originalPath, signedPath, overlays); err != nil {
		return "", errors.Internal("failed to generate signed document", err)
	}
	return signedPath, nil
}

func (s *Service) List(ctx context.Context, ownerID string, status string) ([]DocumentResponse, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	query := s.orm.WithContext(ctx).Where("owner_id = ?", uid)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var records []schemas.Document
	if err := query.Order("created_at desc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list documents", err)
	}

	out := make([]DocumentResponse, len(records))
	for i := range records {
		out[i] = *toResponse(&records[i])
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, ownerID string, docID string) (*DocumentResponse, error) {
	record, err := s.findOwned(ctx, ownerID, docID)
	if err != nil {
		return nil, err
	}
	return toResponse(record), nil
}

func (s *Service) Update(ctx context.Context, ownerID string, docID string, req *UpdateRequest) (*DocumentResponse, error) {
	record, err := s.findOwned(ctx, ownerID, docID)
	if err != nil {
		return nil, err
	}
	if record.Status != "draft" {
		return nil, errors.Invalid("only draft documents can be updated")
	}
	if req.Name != "" {
		record.Name = req.Name
	}
	if req.FileName != "" {
		record.FileName = req.FileName
	}
	if err := s.orm.WithContext(ctx).Save(record).Error; err != nil {
		return nil, errors.Internal("failed to update document", err)
	}
	return toResponse(record), nil
}

func (s *Service) Delete(ctx context.Context, ownerID string, docID string) error {
	record, err := s.findOwned(ctx, ownerID, docID)
	if err != nil {
		return err
	}
	if record.StoragePath != "" {
		fullPath := filepath.Join(s.uploadDir, record.StoragePath)
		os.Remove(fullPath)
		signedPath := strings.TrimSuffix(fullPath, ".pdf") + "_signed.pdf"
		os.Remove(signedPath)
	}
	if err := s.orm.WithContext(ctx).Delete(record).Error; err != nil {
		return errors.Internal("failed to delete document", err)
	}
	return nil
}

func (s *Service) Send(ctx context.Context, ownerID string, docID string) (*DocumentResponse, error) {
	record, err := s.findOwned(ctx, ownerID, docID)
	if err != nil {
		return nil, err
	}
	if record.Status != "draft" {
		return nil, errors.Invalid("only draft documents can be sent")
	}

	var signers []schemas.Signer
	if err := s.orm.WithContext(ctx).Where("document_id = ?", record.ID).Find(&signers).Error; err != nil {
		return nil, errors.Internal("failed to load signers", err)
	}
	if len(signers) == 0 {
		return nil, errors.Invalid("document must have at least one signer")
	}

	for i := range signers {
		signers[i].Token = generateToken()
		if err := s.orm.WithContext(ctx).Save(&signers[i]).Error; err != nil {
			return nil, errors.Internal("failed to generate signer token", err)
		}
	}

	record.Status = "pending"
	if err := s.orm.WithContext(ctx).Save(record).Error; err != nil {
		return nil, errors.Internal("failed to update document status", err)
	}

	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	for i := range signers {
		if signers[i].Role == "signer" || signers[i].Role == "approver" {
			go s.smtp.SendSigningEmail(uid, signers[i].Name, signers[i].Email, record.Name, signers[i].Token, s.domain)
		}
	}

	sentEvent := webhooks.EventPayload{
		EventType: "document.sent",
		Document:  webhooks.EventDocument{ID: record.ID, Name: record.Name, Status: record.Status, FileName: record.FileName},
	}
	go s.webhookSvc.Dispatch(uid, sentEvent)

	return toResponse(record), nil
}

func (s *Service) Stats(ctx context.Context, ownerID string) (*StatsResponse, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	var total, pending, completed int64
	s.orm.WithContext(ctx).Model(&schemas.Document{}).Where("owner_id = ?", uid).Count(&total)
	s.orm.WithContext(ctx).Model(&schemas.Document{}).Where("owner_id = ? AND status = ?", uid, "pending").Count(&pending)
	s.orm.WithContext(ctx).Model(&schemas.Document{}).Where("owner_id = ? AND status = ?", uid, "completed").Count(&completed)
	return &StatsResponse{Total: total, Pending: pending, Completed: completed}, nil
}

func (s *Service) FindByID(ctx context.Context, docID int64) (*schemas.Document, error) {
	var record schemas.Document
	err := s.orm.WithContext(ctx).Where("id = ?", docID).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("document not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to read document", err)
	}
	return &record, nil
}

func (s *Service) findOwned(ctx context.Context, ownerID string, docID string) (*schemas.Document, error) {
	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	did, _ := strconv.ParseInt(docID, 10, 64)

	var record schemas.Document
	err := s.orm.WithContext(ctx).Where("id = ? AND owner_id = ?", did, uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("document not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to read document", err)
	}
	return &record, nil
}

func (s *Service) UpdateStatus(ctx context.Context, docID int64, status string) error {
	return s.orm.WithContext(ctx).Model(&schemas.Document{}).Where("id = ?", docID).Update("status", status).Error
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func toResponse(record *schemas.Document) *DocumentResponse {
	return &DocumentResponse{
		ID:        record.ID,
		Name:      record.Name,
		Status:    record.Status,
		FileName:  record.FileName,
		OwnerID:   record.OwnerID,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}
