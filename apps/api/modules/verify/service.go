package verify

import (
	"context"
	stderrors "errors"
	"strings"

	"api/internal/errors"
	"api/modules/documents"
	"api/schemas"

	"gorm.io/gorm"
)

type Service struct {
	orm        *gorm.DB
	docService *documents.Service
}

func NewService(orm *gorm.DB, docService *documents.Service) *Service {
	return &Service{orm: orm, docService: docService}
}

func (s *Service) Lookup(ctx context.Context, hash string) (*Response, error) {
	doc, matchedSigned, err := s.docService.FindByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return &Response{Match: false, Hash: hash}, nil
	}

	variant := "original"
	if matchedSigned {
		variant = "signed"
	}

	docDTO := &DocumentDTO{
		Name:      doc.Name,
		FileName:  doc.FileName,
		Status:    doc.Status,
		CreatedAt: doc.CreatedAt,
	}
	if doc.Status == "completed" {
		completedAt := doc.UpdatedAt
		docDTO.CompletedAt = &completedAt
	}

	var signers []schemas.Signer
	if err := s.orm.WithContext(ctx).
		Where("document_id = ?", doc.ID).
		Order("order_num asc, id asc").
		Find(&signers).Error; err != nil && !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Internal("failed to load signers", err)
	}

	signerDTOs := make([]SignerDTO, 0, len(signers))
	for i := range signers {
		signer := signers[i]
		signerDTOs = append(signerDTOs, SignerDTO{
			Name:     signer.Name,
			Email:    maskEmail(signer.Email),
			Status:   signer.Status,
			SignedAt: signer.SignedAt,
		})
	}

	return &Response{
		Match:    true,
		Hash:     hash,
		Variant:  variant,
		Document: docDTO,
		Signers:  signerDTOs,
	}, nil
}

func maskEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return "***"
	}
	local := email[:at]
	domain := email[at+1:]
	return maskPart(local, 1) + "@" + maskPart(domain, 1)
}

func maskPart(part string, keep int) string {
	if len(part) <= keep {
		return strings.Repeat("*", len(part))
	}
	return part[:keep] + strings.Repeat("*", len(part)-keep)
}
