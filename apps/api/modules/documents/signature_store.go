package documents

import (
	"context"
	stderrors "errors"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

func (s *Service) insertSignersAndFields(ctx context.Context, docID int64, req *SignatureRequest, pages int) error {
	return s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		signers := make([]schemas.Signer, len(req.Signers))
		for i, in := range req.Signers {
			signers[i] = schemas.Signer{DocumentID: docID, Name: in.Name, Email: in.Email, Role: in.Role, Status: "pending", OrderNum: in.Order}
		}
		if err := tx.Create(&signers).Error; err != nil {
			return errors.Internal("failed to add signers", err)
		}
		fields := buildSignatureFields(docID, req, signers, pages)
		if len(fields) == 0 {
			return nil
		}
		if err := tx.Create(&fields).Error; err != nil {
			return errors.Internal("failed to add fields", err)
		}
		return nil
	})
}

func buildSignatureFields(docID int64, req *SignatureRequest, signers []schemas.Signer, pages int) []schemas.Field {
	fields := make([]schemas.Field, 0, len(req.Fields))
	for _, f := range req.Fields {
		fields = append(fields, schemas.Field{
			DocumentID: docID, SignerID: signers[f.Signer].ID, FieldType: f.FieldType, Page: f.Page,
			X: f.X, Y: f.Y, Width: f.Width, Height: f.Height, Required: f.Required, Label: f.Label,
		})
	}
	if len(req.Fields) > 0 {
		return fields
	}
	row := 0
	for _, signer := range signers {
		if signer.Role == "viewer" {
			continue
		}
		fields = append(fields, DefaultSignatureField(docID, signer.ID, pages, row))
		row++
	}
	return fields
}

// DefaultSignatureField is the required signature box a signer gets when nobody placed
// one: on the last page, stacked by row so up to five signers do not overlap.
func DefaultSignatureField(docID, signerID int64, pages, row int) schemas.Field {
	return schemas.Field{
		DocumentID: docID, SignerID: signerID, FieldType: "signature", Page: pages, Required: true,
		X: 8, Y: float64(70 - 12*(row%5)), Width: 30, Height: 8, Label: "Signature",
	}
}

func (s *Service) findByExternalRef(ctx context.Context, ownerID int64, ref string) (*schemas.Document, error) {
	if ref == "" {
		return nil, nil
	}
	var record schemas.Document
	err := s.orm.WithContext(ctx).Where("owner_id = ? AND external_ref = ?", ownerID, ref).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Internal("failed to look up document", err)
	}
	return &record, nil
}

// LookupByExternalRef returns the document the caller created with the given
// external_ref, with its signers, or a not-found error.
func (s *Service) LookupByExternalRef(ctx context.Context, ownerID int64, ref string) (*SignatureResponse, error) {
	if err := validateRef("external_ref", ref); err != nil || ref == "" {
		return nil, errors.Invalid("external_ref is required and must be valid")
	}
	record, err := s.findByExternalRef(ctx, ownerID, ref)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, errors.NotFound("document not found")
	}
	return s.signatureResponse(ctx, record)
}

func (s *Service) signatureResponse(ctx context.Context, doc *schemas.Document) (*SignatureResponse, error) {
	var signers []schemas.Signer
	if err := s.orm.WithContext(ctx).Where("document_id = ?", doc.ID).Order("order_num asc, id asc").Find(&signers).Error; err != nil {
		return nil, errors.Internal("failed to load signers", err)
	}
	out := make([]SignatureSignerStatus, len(signers))
	for i, sg := range signers {
		out[i] = SignatureSignerStatus{ID: sg.ID, Name: sg.Name, Email: sg.Email, Role: sg.Role, Status: sg.Status, OrderNum: sg.OrderNum, SignedAt: sg.SignedAt}
		if sg.Token != "" && s.domain != "" {
			out[i].SigningURL = s.domain + "/share/" + sg.Token
		}
	}
	return &SignatureResponse{Document: *toResponse(doc), Signers: out}, nil
}
