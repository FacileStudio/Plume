package pdfutil

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

var sigImgCounter int

func init() {
	// Run with built-in defaults; never touch disk for config (distroless has no
	// home/config dir to read or write).
	model.ConfigPath = "disable"
}

type FieldOverlay struct {
	Page      int
	X         float64
	Y         float64
	Width     float64
	Height    float64
	FieldType string
	Value     string
}

// FlattenFields burns the given field overlays into inputPath and writes the
// result to outputPath. It builds a transparent overlay PDF with fpdf (one page
// per source page, matching dimensions) and stamps it onto the source with
// pdfcpu, which reads arbitrary PDFs robustly — unlike fpdf's page importer,
// which panics on unsupported content-stream filters (e.g. /ASCII85Decode).
func FlattenFields(inputPath, outputPath string, fields []FieldOverlay) (err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("pdfutil.FlattenFields panic", "recover", r, "input", inputPath)
			err = fmt.Errorf("pdfutil: failed to flatten %q: %v", inputPath, r)
		}
	}()

	dims, err := api.PageDimsFile(inputPath)
	if err != nil {
		return fmt.Errorf("pdfutil: read page dimensions of %q: %w", inputPath, err)
	}
	totalPages := len(dims)
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

	// Without any fields the signed document is identical to the original.
	if len(fieldsByPage) == 0 {
		return copyFile(inputPath, outputPath)
	}

	pdf := fpdf.New("P", "pt", "A4", "")
	pdf.SetAutoPageBreak(false, 0)
	RegisterUnicodeFonts(pdf)

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		d := dims[pageNum-1]
		w, h := d.Width, d.Height
		if w <= 0 || h <= 0 {
			return fmt.Errorf("pdfutil: page %d has invalid dimensions (%fx%f)", pageNum, w, h)
		}

		orient := "P"
		if w > h {
			orient = "L"
		}
		pdf.AddPageFormat(orient, fpdf.SizeType{Wd: w, Ht: h})

		for _, field := range fieldsByPage[pageNum] {
			drawField(pdf, field, w, h)
		}
	}

	overlayFile, err := os.CreateTemp(filepath.Dir(outputPath), "plume-overlay-*.pdf")
	if err != nil {
		return fmt.Errorf("pdfutil: create overlay temp file: %w", err)
	}
	overlayPath := overlayFile.Name()
	overlayFile.Close()
	defer os.Remove(overlayPath)

	if err := pdf.OutputFileAndClose(overlayPath); err != nil {
		return fmt.Errorf("pdfutil: write overlay: %w", err)
	}

	// Multi-stamp mode (PdfPageNrSrc == 0) maps overlay page N onto source page N,
	// placed bottom-left at 1:1 scale so fpdf's coordinates line up exactly.
	wm, err := pdfcpu.ParsePDFWatermarkDetails(overlayPath, "scalefactor:1 abs, position:bl, offset:0 0, rotation:0, opacity:1", true, types.POINTS)
	if err != nil {
		return fmt.Errorf("pdfutil: build stamp: %w", err)
	}

	if err := api.AddWatermarksFile(inputPath, outputPath, nil, wm, nil); err != nil {
		return fmt.Errorf("pdfutil: stamp overlay onto %q: %w", inputPath, err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("pdfutil: read %q: %w", src, err)
	}
	if err := os.WriteFile(dst, in, 0o644); err != nil {
		return fmt.Errorf("pdfutil: write %q: %w", dst, err)
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
