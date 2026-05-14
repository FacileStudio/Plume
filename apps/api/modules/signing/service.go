package signing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"api/internal/errors"
	"api/internal/hashing"
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

func (s *Service) GetOrGenerateAuditTrail(ctx context.Context, ownerID string, docID string) (string, error) {
	_, err := s.docService.Get(ctx, ownerID, docID)
	if err != nil {
		return "", err
	}

	did, _ := strconv.ParseInt(docID, 10, 64)

	var doc schemas.Document
	if err := s.orm.WithContext(ctx).Where("id = ?", did).First(&doc).Error; err != nil {
		return "", errors.Internal("failed to load document", err)
	}

	trailPath := s.auditTrailPath(did)
	if _, err := os.Stat(trailPath); err == nil {
		return trailPath, nil
	}

	if err := s.generateAuditTrail(did, &doc); err != nil {
		return "", errors.Internal("failed to generate audit trail", err)
	}
	return trailPath, nil
}

func (s *Service) auditTrailPath(docID int64) string {
	return filepath.Join(s.uploadDir, "audit", fmt.Sprintf("audit_%d.pdf", docID))
}

func (s *Service) generateAuditTrail(docID int64, doc *schemas.Document) error {
	var signers []schemas.Signer
	if err := s.orm.Where("document_id = ?", docID).Order("order_num asc").Find(&signers).Error; err != nil {
		return err
	}

	var fields []schemas.Field
	s.orm.Where("document_id = ?", docID).Find(&fields)

	var owner schemas.User
	s.orm.Where("id = ?", doc.OwnerID).First(&owner)

	fileHash := s.hashDocumentFile(doc)

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 25)
	pdf.SetMargins(15, 15, 15)

	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(150, 150, 150)
		pdf.CellFormat(0, 8, fmt.Sprintf("Plume Audit Trail — Document #%d — Page %d", docID, pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 22)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(180, 12, "Audit Trail", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(180, 5, fmt.Sprintf("Generated %s", time.Now().UTC().Format("January 2, 2006 at 15:04 UTC")), "", 1, "L", false, 0, "")
	pdf.Ln(8)

	s.drawSectionLine(pdf)
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(180, 8, "Document Information", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	s.drawInfoRow(pdf, "Name", doc.Name)
	s.drawInfoRow(pdf, "File", doc.FileName)
	s.drawInfoRow(pdf, "Status", strings.ToUpper(doc.Status))
	s.drawInfoRow(pdf, "Owner", fmt.Sprintf("%s (%s)", owner.Name, owner.Email))
	s.drawInfoRow(pdf, "Created", doc.CreatedAt.Format("January 2, 2006 at 15:04 UTC"))
	s.drawInfoRow(pdf, "Last updated", doc.UpdatedAt.Format("January 2, 2006 at 15:04 UTC"))
	if fileHash != "" {
		s.drawInfoRow(pdf, "Document SHA-256", fileHash)
	}

	pdf.Ln(6)
	s.drawSectionLine(pdf)
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(180, 8, "Event Timeline", "", 1, "L", false, 0, "")
	pdf.Ln(4)

	s.drawEvent(pdf, doc.CreatedAt, "Document created", fmt.Sprintf("Created by %s", owner.Email))

	var sentTime *time.Time
	for _, signer := range signers {
		if signer.Token != "" && signer.CreatedAt.Before(doc.UpdatedAt) {
			sentTime = &doc.UpdatedAt
			break
		}
	}
	if doc.Status != "draft" && sentTime != nil {
		recipientList := make([]string, 0, len(signers))
		for _, sn := range signers {
			if sn.Role != "viewer" {
				recipientList = append(recipientList, fmt.Sprintf("%s (%s)", sn.Name, sn.Email))
			}
		}
		s.drawEvent(pdf, *sentTime, "Document sent for signing", fmt.Sprintf("Sent to %s", strings.Join(recipientList, ", ")))
	}

	for _, signer := range signers {
		if signer.SignedAt == nil {
			continue
		}
		action := "Signed"
		if signer.Status == "declined" {
			action = "Declined"
		}
		detail := fmt.Sprintf("%s by %s (%s)", action, signer.Name, signer.Email)
		if signer.IPAddress != "" {
			detail += fmt.Sprintf("\nIP: %s", signer.IPAddress)
		}
		if signer.UserAgent != "" {
			ua := signer.UserAgent
			if len(ua) > 80 {
				ua = ua[:80] + "..."
			}
			detail += fmt.Sprintf("\nUser Agent: %s", ua)
		}
		s.drawEvent(pdf, *signer.SignedAt, fmt.Sprintf("Signer %s", strings.ToLower(action)), detail)
	}

	if doc.Status == "completed" {
		s.drawEvent(pdf, doc.UpdatedAt, "Document completed", "All signers have completed their signatures")
	} else if doc.Status == "declined" {
		s.drawEvent(pdf, doc.UpdatedAt, "Document declined", "A signer has declined to sign")
	}

	pdf.Ln(4)
	s.drawSectionLine(pdf)
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "B", 13)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(180, 8, "Signers", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	for i, signer := range signers {
		if i > 0 {
			pdf.Ln(2)
		}
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(180, 6, fmt.Sprintf("%d. %s", i+1, signer.Name), "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(80, 80, 80)
		pdf.CellFormat(180, 5, fmt.Sprintf("Email: %s", signer.Email), "", 1, "L", false, 0, "")
		pdf.CellFormat(180, 5, fmt.Sprintf("Role: %s  •  Status: %s", signer.Role, signer.Status), "", 1, "L", false, 0, "")
		if signer.SignedAt != nil {
			pdf.CellFormat(180, 5, fmt.Sprintf("Responded: %s", signer.SignedAt.Format("January 2, 2006 at 15:04 UTC")), "", 1, "L", false, 0, "")
		}
		if signer.IPAddress != "" {
			pdf.CellFormat(180, 5, fmt.Sprintf("IP address: %s", signer.IPAddress), "", 1, "L", false, 0, "")
		}
		if signer.UserAgent != "" {
			ua := signer.UserAgent
			if len(ua) > 100 {
				ua = ua[:100] + "..."
			}
			pdf.CellFormat(180, 5, fmt.Sprintf("User agent: %s", ua), "", 1, "L", false, 0, "")
		}
	}

	completedFields := make([]schemas.Field, 0)
	for _, f := range fields {
		if f.Value != "" {
			completedFields = append(completedFields, f)
		}
	}

	if len(completedFields) > 0 {
		pdf.Ln(4)
		s.drawSectionLine(pdf)
		pdf.Ln(6)

		pdf.SetFont("Helvetica", "B", 13)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(180, 8, "Field Values", "", 1, "L", false, 0, "")
		pdf.Ln(2)

		signerMap := make(map[int64]string)
		for _, sn := range signers {
			signerMap[sn.ID] = sn.Name
		}

		for _, f := range completedFields {
			pdf.SetFont("Helvetica", "B", 9)
			pdf.SetTextColor(0, 0, 0)

			label := f.Label
			if label == "" {
				label = strings.ToUpper(f.FieldType[:1]) + f.FieldType[1:]
			}
			pdf.CellFormat(180, 5, fmt.Sprintf("%s (%s — %s)", label, f.FieldType, signerMap[f.SignerID]), "", 1, "L", false, 0, "")

			pdf.SetFont("Helvetica", "", 9)
			pdf.SetTextColor(80, 80, 80)
			val := f.Value
			if strings.HasPrefix(val, "data:image/") {
				val = "[drawn signature image]"
			}
			if f.FieldType == "checkbox" {
				if val == "true" {
					val = "Checked"
				} else {
					val = "Unchecked"
				}
			}
			pdf.CellFormat(180, 5, fmt.Sprintf("Value: %s", val), "", 1, "L", false, 0, "")
			pdf.Ln(1)
		}
	}

	auditDir := filepath.Join(s.uploadDir, "audit")
	if err := os.MkdirAll(auditDir, 0o755); err != nil {
		return err
	}

	return pdf.OutputFileAndClose(s.auditTrailPath(docID))
}

func (s *Service) drawSectionLine(pdf *fpdf.Fpdf) {
	pdf.SetDrawColor(220, 220, 220)
	pdf.SetLineWidth(0.3)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
}

func (s *Service) drawInfoRow(pdf *fpdf.Fpdf, label, value string) {
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(35, 6, label, "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(30, 30, 30)
	pdf.CellFormat(145, 6, value, "", 1, "L", false, 0, "")
}

func (s *Service) drawEvent(pdf *fpdf.Fpdf, t time.Time, title, detail string) {
	y := pdf.GetY()
	pdf.SetFillColor(30, 30, 30)
	pdf.Circle(19, y+3, 2.5, "F")

	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.3)
	pdf.Line(19, y+6, 19, y+18)

	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(130, 130, 130)
	pdf.SetXY(25, y)
	pdf.CellFormat(40, 4, t.Format("2006-01-02 15:04 UTC"), "", 0, "L", false, 0, "")

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetXY(25, y+4)
	pdf.CellFormat(155, 5, title, "", 1, "L", false, 0, "")

	if detail != "" {
		lines := strings.Split(detail, "\n")
		for _, line := range lines {
			pdf.SetFont("Helvetica", "", 8)
			pdf.SetTextColor(100, 100, 100)
			pdf.SetX(25)
			pdf.CellFormat(155, 4, line, "", 1, "L", false, 0, "")
		}
	}
	pdf.Ln(3)
}

func (s *Service) hashDocumentFile(doc *schemas.Document) string {
	if doc.OriginalHash != "" {
		return doc.OriginalHash
	}
	if doc.StoragePath == "" {
		return ""
	}
	hash, err := hashing.SHA256File(filepath.Join(s.uploadDir, doc.StoragePath))
	if err != nil {
		return ""
	}
	return hash
}
