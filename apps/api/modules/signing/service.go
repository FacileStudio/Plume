package signing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"api/internal/errors"
	"api/modules/documents"
	"api/schemas"

	"github.com/go-pdf/fpdf"
	"gorm.io/gorm"
)

type Service struct {
	orm        *gorm.DB
	uploadDir  string
	docService *documents.Service
}

func NewService(orm *gorm.DB, uploadDir string, docService *documents.Service) *Service {
	return &Service{orm: orm, uploadDir: uploadDir, docService: docService}
}

func (s *Service) GetOrGenerateCertificate(ctx context.Context, ownerID string, docID string) (string, error) {
	_, err := s.docService.Get(ctx, ownerID, docID)
	if err != nil {
		return "", err
	}

	did, _ := strconv.ParseInt(docID, 10, 64)

	var doc schemas.Document
	if err := s.orm.WithContext(ctx).Where("id = ?", did).First(&doc).Error; err != nil {
		return "", errors.Internal("failed to load document", err)
	}
	if doc.Status != "completed" {
		return "", errors.Invalid("certificate is only available for completed documents")
	}

	certPath := s.certificatePath(did)
	if _, err := os.Stat(certPath); err == nil {
		return certPath, nil
	}

	if err := s.generateCertificate(did, &doc); err != nil {
		return "", errors.Internal("failed to generate certificate", err)
	}
	return certPath, nil
}

func (s *Service) certificatePath(docID int64) string {
	return filepath.Join(s.uploadDir, "certificates", fmt.Sprintf("cert_%d.pdf", docID))
}

func (s *Service) generateCertificate(docID int64, doc *schemas.Document) error {
	var signers []schemas.Signer
	if err := s.orm.Where("document_id = ?", docID).Order("order_num asc").Find(&signers).Error; err != nil {
		return err
	}

	var fields []schemas.Field
	s.orm.Where("document_id = ? AND value != ''", docID).Find(&fields)

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 20)
	pdf.CellFormat(190, 15, "Signature Certificate", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(40, 7, "Document:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(150, 7, doc.Name, "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(40, 7, "File:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(150, 7, doc.FileName, "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(40, 7, "Status:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(150, 7, doc.Status, "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(40, 7, "Completed:", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(150, 7, doc.UpdatedAt.Format(time.RFC3339), "", 1, "L", false, 0, "")

	pdf.Ln(10)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(190, 10, "Signers", "", 1, "L", false, 0, "")
	pdf.Ln(3)

	for _, signer := range signers {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(190, 7, signer.Name, "", 1, "L", false, 0, "")

		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(190, 6, fmt.Sprintf("Email: %s", signer.Email), "", 1, "L", false, 0, "")
		pdf.CellFormat(190, 6, fmt.Sprintf("Role: %s  |  Status: %s", signer.Role, signer.Status), "", 1, "L", false, 0, "")

		if signer.SignedAt != nil {
			pdf.CellFormat(190, 6, fmt.Sprintf("Signed at: %s", signer.SignedAt.Format(time.RFC3339)), "", 1, "L", false, 0, "")
		}
		if signer.IPAddress != "" {
			pdf.CellFormat(190, 6, fmt.Sprintf("IP: %s", signer.IPAddress), "", 1, "L", false, 0, "")
		}
		pdf.Ln(4)
	}

	if len(fields) > 0 {
		pdf.Ln(5)
		pdf.SetDrawColor(200, 200, 200)
		pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
		pdf.Ln(8)

		pdf.SetFont("Helvetica", "B", 14)
		pdf.CellFormat(190, 10, "Field Values", "", 1, "L", false, 0, "")
		pdf.Ln(3)

		for _, field := range fields {
			var signerName string
			for _, sn := range signers {
				if sn.ID == field.SignerID {
					signerName = sn.Name
					break
				}
			}
			pdf.SetFont("Helvetica", "", 10)
			label := fmt.Sprintf("%s (%s): %s", field.FieldType, signerName, field.Value)
			pdf.CellFormat(190, 6, label, "", 1, "L", false, 0, "")
		}
	}

	certDir := filepath.Join(s.uploadDir, "certificates")
	if err := os.MkdirAll(certDir, 0o755); err != nil {
		return err
	}

	return pdf.OutputFileAndClose(s.certificatePath(docID))
}
