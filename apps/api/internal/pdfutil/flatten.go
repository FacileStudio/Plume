package pdfutil

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-pdf/fpdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// FieldOverlay describes a single field value to stamp onto a page, in
// percentage-of-page coordinates.
type FieldOverlay struct {
	Page      int
	X         float64
	Y         float64
	Width     float64
	Height    float64
	FieldType string
	Value     string
}

// rect holds a field's absolute bounds on a page, plus the raw value the
// field carries so each field-type painter can decide what to draw.
type rect struct {
	x, y, w, h float64
	value      string
}

// configurePdfcpu runs pdfcpu with built-in defaults rather than an on-disk
// config lookup; the distroless container has no home/config dir to read or
// write.
func configurePdfcpu() {
	model.ConfigPath = "disable"
}

// FlattenFields burns the given field overlays into inputPath and writes the
// result to outputPath. Signature values may be data URLs, stamped as images.
// Without any fields the signed document is identical to the original.
func FlattenFields(inputPath, outputPath string, fields []FieldOverlay) (err error) {
	configurePdfcpu()

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
	if len(dims) == 0 {
		return fmt.Errorf("pdfutil: source PDF %q has no pages", inputPath)
	}

	fieldsByPage := buildFieldsByPage(fields)
	if len(fieldsByPage) == 0 {
		return copyFile(inputPath, outputPath)
	}

	return stampOverlay(inputPath, outputPath, dims, fieldsByPage)
}

// stampOverlay renders the overlay, stamps it onto inputPath and writes the
// result to outputPath.
func stampOverlay(inputPath, outputPath string, dims []types.Dim, fieldsByPage map[int][]FieldOverlay) error {
	overlayPath, err := writeOverlay(filepath.Dir(outputPath), dims, fieldsByPage)
	if err != nil {
		return err
	}
	defer os.Remove(overlayPath)

	wm, err := pdfcpu.ParsePDFWatermarkDetails(overlayPath, "scalefactor:1 abs, position:bl, offset:0 0, rotation:0, opacity:1", true, types.POINTS)
	if err != nil {
		return fmt.Errorf("pdfutil: build stamp: %w", err)
	}

	if err := api.AddWatermarksFile(inputPath, outputPath, nil, wm, nil); err != nil {
		return fmt.Errorf("pdfutil: stamp overlay onto %q: %w", inputPath, err)
	}
	return nil
}

// buildFieldsByPage groups the non-empty overlays by page number.
func buildFieldsByPage(fields []FieldOverlay) map[int][]FieldOverlay {
	fieldsByPage := make(map[int][]FieldOverlay)
	for _, f := range fields {
		if f.Value == "" {
			continue
		}
		fieldsByPage[f.Page] = append(fieldsByPage[f.Page], f)
	}
	return fieldsByPage
}

// writeOverlay renders one fpdf page per source page with the field values
// stamped, writes it to a temp PDF file next to dir, and returns its path. The
// caller owns the file and must remove it.
func writeOverlay(dir string, dims []types.Dim, fieldsByPage map[int][]FieldOverlay) (string, error) {
	pdf := fpdf.New("P", "pt", "A4", "")
	pdf.SetAutoPageBreak(false, 0)
	RegisterUnicodeFonts(pdf)

	sigSeq := 0

	for pageNum := 1; pageNum <= len(dims); pageNum++ {
		d := dims[pageNum-1]
		w, h := d.Width, d.Height
		if w <= 0 || h <= 0 {
			return "", fmt.Errorf("pdfutil: page %d has invalid dimensions (%fx%f)", pageNum, w, h)
		}

		orient := "P"
		if w > h {
			orient = "L"
		}
		pdf.AddPageFormat(orient, fpdf.SizeType{Wd: w, Ht: h})

		for _, field := range fieldsByPage[pageNum] {
			drawField(pdf, field, fieldRect(field, w, h), &sigSeq)
		}
	}

	return writeOverlayFile(dir, pdf)
}

// fieldRect converts a field's percentage-of-page coordinates to absolute
// bounds on a page of the given dimensions.
func fieldRect(field FieldOverlay, pageW, pageH float64) rect {
	return rect{
		x:     field.X / 100 * pageW,
		y:     field.Y / 100 * pageH,
		w:     field.Width / 100 * pageW,
		h:     field.Height / 100 * pageH,
		value: field.Value,
	}
}

// writeOverlayFile writes the rendered overlay PDF to a temp file next to dir
// and returns its path.
func writeOverlayFile(dir string, pdf *fpdf.Fpdf) (string, error) {
	overlayFile, err := os.CreateTemp(dir, "plume-overlay-*.pdf")
	if err != nil {
		return "", fmt.Errorf("pdfutil: create overlay temp file: %w", err)
	}
	overlayPath := overlayFile.Name()
	overlayFile.Close()

	if err := pdf.OutputFileAndClose(overlayPath); err != nil {
		return "", fmt.Errorf("pdfutil: write overlay: %w", err)
	}
	return overlayPath, nil
}

// copyFile is used when no field has a value: the signed document must then be
// identical to the original.
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
