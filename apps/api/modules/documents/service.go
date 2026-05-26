package documents

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"api/internal/errors"
	"api/internal/hashing"
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

	go s.webhookSvc.Dispatch(uid, webhooks.BuildDocumentEvent(webhooks.EventDocumentCreated, record, s.domain))

	return toResponse(record), nil
}

func (s *Service) UpdateStorage(ctx context.Context, docID int64, path string, originalHash string) error {
	err := s.orm.WithContext(ctx).Model(&schemas.Document{}).Where("id = ?", docID).Updates(map[string]any{
		"storage_path":  path,
		"original_hash": originalHash,
	}).Error
	if err != nil {
		return errors.Internal("failed to update storage", err)
	}
	return nil
}

func (s *Service) GetFilePath(ctx context.Context, ownerID string, docID string) (string, error) {
	record, err := s.findOwned(ctx, ownerID, docID)
	if err != nil {
		return "", err
	}
	return s.resolveFilePath(ctx, record)
}

func (s *Service) GetFilePathByDocID(ctx context.Context, docID int64) (string, error) {
	record, err := s.FindByID(ctx, docID)
	if err != nil {
		return "", err
	}
	return s.resolveFilePath(ctx, record)
}

func (s *Service) resolveFilePath(ctx context.Context, record *schemas.Document) (string, error) {
	if record.StoragePath == "" {
		return "", errors.NotFound("no file uploaded for this document")
	}
	originalPath := filepath.Join(s.uploadDir, record.StoragePath)
	if record.Status == "completed" {
		return s.getOrCreateSignedFile(ctx, record, originalPath)
	}
	return originalPath, nil
}

func (s *Service) getOrCreateSignedFile(ctx context.Context, doc *schemas.Document, originalPath string) (string, error) {
	signedPath := signedFilePath(originalPath)

	if signedInfo, err := os.Stat(signedPath); err == nil {
		if origInfo, origErr := os.Stat(originalPath); origErr == nil && signedInfo.ModTime().After(origInfo.ModTime()) {
			s.fillSignedHash(ctx, doc, signedPath)
			return signedPath, nil
		}
		_ = os.Remove(signedPath)
	}

	var fields []schemas.Field
	if err := s.orm.WithContext(ctx).Where("document_id = ? AND value != ''", doc.ID).Find(&fields).Error; err != nil {
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
		_ = os.Remove(signedPath)
		return "", errors.Internal("failed to generate signed document", err)
	}
	s.fillSignedHash(ctx, doc, signedPath)
	return signedPath, nil
}

func (s *Service) fillSignedHash(ctx context.Context, doc *schemas.Document, signedPath string) {
	hash, err := hashing.SHA256File(signedPath)
	if err != nil {
		return
	}
	if doc.SignedHash == hash {
		return
	}
	if err := s.orm.WithContext(ctx).Model(&schemas.Document{}).Where("id = ?", doc.ID).Update("signed_hash", hash).Error; err != nil {
		return
	}
	doc.SignedHash = hash
}

func signedFilePath(originalPath string) string {
	return strings.TrimSuffix(originalPath, ".pdf") + "_signed.pdf"
}

// InvalidateGeneratedFiles removes cached derived PDFs (flattened signed file,
// audit trail, certificate) so subsequent downloads regenerate them from the
// latest field values and signer state. Safe to call when no cache exists.
func (s *Service) InvalidateGeneratedFiles(ctx context.Context, docID int64) {
	record, err := s.FindByID(ctx, docID)
	if err != nil || record.StoragePath == "" {
		return
	}
	originalPath := filepath.Join(s.uploadDir, record.StoragePath)
	_ = os.Remove(signedFilePath(originalPath))
	_ = os.Remove(filepath.Join(s.uploadDir, "audit", fmt.Sprintf("audit_%d.pdf", docID)))
	_ = os.Remove(filepath.Join(s.uploadDir, "certificates", fmt.Sprintf("cert_%d.pdf", docID)))
	if record.SignedHash != "" {
		s.orm.WithContext(ctx).Model(&schemas.Document{}).Where("id = ?", docID).Update("signed_hash", "")
	}
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
	if req.Sequential != nil {
		record.Sequential = *req.Sequential
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
		os.Remove(signedFilePath(fullPath))
	}
	if err := s.orm.WithContext(ctx).Delete(record).Error; err != nil {
		return errors.Internal("failed to delete document", err)
	}

	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	go s.webhookSvc.Dispatch(uid, webhooks.BuildDocumentEvent(webhooks.EventDocumentDeleted, record, s.domain))

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
		token, err := generateToken()
		if err != nil {
			return nil, errors.Internal("failed to generate signing token", err)
		}
		signers[i].Token = token
		if err := s.orm.WithContext(ctx).Save(&signers[i]).Error; err != nil {
			return nil, errors.Internal("failed to save signer token", err)
		}
	}

	record.Status = "pending"
	if err := s.orm.WithContext(ctx).Save(record).Error; err != nil {
		return nil, errors.Internal("failed to update document status", err)
	}

	uid, _ := strconv.ParseInt(ownerID, 10, 64)
	now := time.Now().UTC()
	dispatchInvitation := func(signer *schemas.Signer) {
		s.orm.WithContext(ctx).Model(&schemas.Signer{}).
			Where("id = ?", signer.ID).
			Update("last_reminded_at", now)
		go s.smtp.SendSigningEmail(uid, signer.Name, signer.Email, record.Name, signer.Token, s.domain)
	}

	if record.Sequential {
		minOrder := 0
		found := false
		for i := range signers {
			if signers[i].Role != "signer" && signers[i].Role != "approver" {
				continue
			}
			if !found || signers[i].OrderNum < minOrder {
				minOrder = signers[i].OrderNum
				found = true
			}
		}
		if found {
			for i := range signers {
				if (signers[i].Role == "signer" || signers[i].Role == "approver") && signers[i].OrderNum == minOrder {
					dispatchInvitation(&signers[i])
				}
			}
		}
	} else {
		for i := range signers {
			if signers[i].Role == "signer" || signers[i].Role == "approver" {
				dispatchInvitation(&signers[i])
			}
		}
	}

	go s.webhookSvc.Dispatch(uid, webhooks.BuildDocumentEvent(webhooks.EventDocumentSent, record, s.domain))

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

func (s *Service) FindByHash(ctx context.Context, hash string) (*schemas.Document, bool, error) {
	var record schemas.Document
	err := s.orm.WithContext(ctx).
		Where("original_hash = ? OR signed_hash = ?", hash, hash).
		First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, errors.Internal("failed to look up document", err)
	}
	matchedSigned := record.SignedHash == hash
	return &record, matchedSigned, nil
}

func (s *Service) UploadDir() string {
	return s.uploadDir
}

func (s *Service) BackfillHashes(ctx context.Context) (int, error) {
	var docs []schemas.Document
	if err := s.orm.WithContext(ctx).
		Where("storage_path <> '' AND (original_hash IS NULL OR original_hash = '')").
		Find(&docs).Error; err != nil {
		return 0, errors.Internal("failed to load documents for backfill", err)
	}

	updated := 0
	for i := range docs {
		doc := &docs[i]
		path := filepath.Join(s.uploadDir, doc.StoragePath)
		hash, err := hashing.SHA256File(path)
		if err != nil {
			continue
		}
		if err := s.orm.WithContext(ctx).Model(&schemas.Document{}).
			Where("id = ?", doc.ID).
			Update("original_hash", hash).Error; err != nil {
			continue
		}
		updated++

		if doc.Status == "completed" {
			signedPath := signedFilePath(path)
			if _, statErr := os.Stat(signedPath); statErr == nil {
				if signedHash, hashErr := hashing.SHA256File(signedPath); hashErr == nil {
					s.orm.WithContext(ctx).Model(&schemas.Document{}).
						Where("id = ?", doc.ID).
						Update("signed_hash", signedHash)
				}
			}
		}
	}
	return updated, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func toResponse(record *schemas.Document) *DocumentResponse {
	return &DocumentResponse{
		ID:           record.ID,
		Name:         record.Name,
		Status:       record.Status,
		FileName:     record.FileName,
		OwnerID:      record.OwnerID,
		Sequential:   record.Sequential,
		OriginalHash: record.OriginalHash,
		SignedHash:   record.SignedHash,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}
}
