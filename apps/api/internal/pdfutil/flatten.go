package pdfutil

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log/slog"
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
	RegisterUnicodeFonts(pdf)

	imp := gofpdi.NewImporter()

	// First ImportPage call parses the source file and populates page sizes.
	firstTpl := imp.ImportPage(pdf, inputPath, 1, "/MediaBox")
	sizes := imp.GetPageSizes()
	totalPages := len(sizes)
	if totalPages == 0 {
		return fmt.Errorf("pdfutil: source PDF %q has no pages", inputPath)
	}

	fieldsByPage := make(map[int][]FieldOverlay)
	for _, f := range fields {
		if f.Value == "" {
			continue
		}
		fieldsByPage[f.Page] = append(fieldsByPage[f.Page], f)
	}

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		pageBox, ok := sizes[pageNum]
		if !ok {
			return fmt.Errorf("pdfutil: missing MediaBox for page %d in %q", pageNum, inputPath)
		}
		mediaBox, ok := pageBox["/MediaBox"]
		if !ok {
			return fmt.Errorf("pdfutil: page %d has no MediaBox", pageNum)
		}
		w, h := mediaBox["w"], mediaBox["h"]
		if w <= 0 || h <= 0 {
			return fmt.Errorf("pdfutil: page %d has invalid dimensions (%fx%f)", pageNum, w, h)
		}

		orient := "P"
		if w > h {
			orient = "L"
		}
		pdf.AddPageFormat(orient, fpdf.SizeType{Wd: w, Ht: h})

		var tplID int
		if pageNum == 1 {
			tplID = firstTpl
		} else {
			tplID = imp.ImportPage(pdf, inputPath, pageNum, "/MediaBox")
		}
		imp.UseImportedTemplate(pdf, tplID, 0, 0, w, h)

		for _, field := range fieldsByPage[pageNum] {
			drawField(pdf, field, w, h)
		}
	}

	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		slog.Error("pdfutil.FlattenFields output", "err", err, "path", outputPath)
		return err
	}
	return nil
}

func drawField(pdf *fpdf.Fpdf, field FieldOverlay, pageW, pageH float64) {
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
			pdf.SetFont(UnicodeFontFamily, "I", fontSize)
			pdf.Text(x+2, y+h/2+fontSize/3, field.Value)
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
		pdf.SetFont(UnicodeFontFamily, "", fontSize)
		text := truncateToFit(pdf, field.Value, w-4)
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
	runes := []rune(text)
	for i := len(runes) - 1; i > 0; i-- {
		candidate := string(runes[:i]) + "..."
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

	imageType := "png"
	if strings.HasPrefix(parts[0], "data:image/jpeg") || strings.HasPrefix(parts[0], "data:image/jpg") {
		imageType = "jpg"
	}

	sigImgCounter++
	name := fmt.Sprintf("sig_%d", sigImgCounter)

	reader := bytes.NewReader(imgBytes)
	opts := fpdf.ImageOptions{ImageType: imageType}
	pdf.RegisterImageOptionsReader(name, opts, reader)

	padding := h * 0.05
	imgX := x + padding
	imgY := y + padding
	imgW := w - padding*2
	imgH := h - padding*2

	pdf.ImageOptions(name, imgX, imgY, imgW, imgH, false, opts, 0, "")
}
