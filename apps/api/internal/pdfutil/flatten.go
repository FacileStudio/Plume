package pdfutil

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/go-pdf/fpdf/contrib/gofpdi"
)

var sigImgCounter int

type FieldOverlay struct {
	Page      int
	X         float64
	Y         float64
	Width     float64
	Height    float64
	FieldType string
	Value     string
}

func FlattenFields(inputPath, outputPath string, fields []FieldOverlay) error {
	pdf := fpdf.New("P", "pt", "A4", "")
	pdf.SetAutoPageBreak(false, 0)

	tr := pdf.UnicodeTranslatorFromDescriptor("")

	imp := gofpdi.NewImporter()

	tpl1 := imp.ImportPage(pdf, inputPath, 1, "/MediaBox")
	sizes := imp.GetPageSizes()
	totalPages := len(sizes)

	fieldsByPage := make(map[int][]FieldOverlay)
	for _, f := range fields {
		if f.Value != "" {
			fieldsByPage[f.Page] = append(fieldsByPage[f.Page], f)
		}
	}

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		pageSize := sizes[pageNum]["/MediaBox"]
		w, h := pageSize["w"], pageSize["h"]

		orient := "P"
		if w > h {
			orient = "L"
		}
		pdf.AddPageFormat(orient, fpdf.SizeType{Wd: w, Ht: h})

		var tplID int
		if pageNum == 1 {
			tplID = tpl1
		} else {
			tplID = imp.ImportPage(pdf, inputPath, pageNum, "/MediaBox")
		}
		imp.UseImportedTemplate(pdf, tplID, 0, 0, w, h)

		for _, field := range fieldsByPage[pageNum] {
			drawField(pdf, field, w, h, tr)
		}
	}

	return pdf.OutputFileAndClose(outputPath)
}

func drawField(pdf *fpdf.Fpdf, field FieldOverlay, pageW, pageH float64, tr func(string) string) {
	x := field.X / 100 * pageW
	y := field.Y / 100 * pageH
	w := field.Width / 100 * pageW
	h := field.Height / 100 * pageH

	pdf.SetTextColor(0, 0, 0)

	switch field.FieldType {
	case "signature":
		if strings.HasPrefix(field.Value, "data:image/") {
			drawSignatureImage(pdf, field.Value, x, y, w, h)
		} else {
			fontSize := clampFontSize(h*0.55, 6, 24)
			pdf.SetFont("Times", "I", fontSize)
			pdf.Text(x+2, y+h/2+fontSize/3, tr(field.Value))
		}

	case "checkbox":
		if field.Value == "true" {
			pdf.SetDrawColor(30, 30, 30)
			lw := h * 0.08
			if lw < 1 {
				lw = 1
			}
			pdf.SetLineWidth(lw)
			pdf.Line(x+w*0.2, y+h*0.55, x+w*0.42, y+h*0.78)
			pdf.Line(x+w*0.42, y+h*0.78, x+w*0.8, y+h*0.22)
		}

	default:
		fontSize := clampFontSize(h*0.55, 6, 14)
		pdf.SetFont("Helvetica", "", fontSize)
		text := truncateToFit(pdf, tr(field.Value), w-4)
		pdf.Text(x+2, y+h/2+fontSize/3, text)
	}
}

func clampFontSize(size, min, max float64) float64 {
	if size < min {
		return min
	}
	if size > max {
		return max
	}
	return size
}

func truncateToFit(pdf *fpdf.Fpdf, text string, maxW float64) string {
	if pdf.GetStringWidth(text) <= maxW {
		return text
	}
	for i := len(text) - 1; i > 0; i-- {
		candidate := text[:i] + "..."
		if pdf.GetStringWidth(candidate) <= maxW {
			return candidate
		}
	}
	return text
}

func drawSignatureImage(pdf *fpdf.Fpdf, dataURL string, x, y, w, h float64) {
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return
	}

	imgBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return
	}

	sigImgCounter++
	name := fmt.Sprintf("sig_%d", sigImgCounter)

	reader := bytes.NewReader(imgBytes)
	opts := fpdf.ImageOptions{ImageType: "png"}
	pdf.RegisterImageOptionsReader(name, opts, reader)

	padding := h * 0.05
	imgX := x + padding
	imgY := y + padding
	imgW := w - padding*2
	imgH := h - padding*2

	pdf.ImageOptions(name, imgX, imgY, imgW, imgH, false, opts, 0, "")
}
