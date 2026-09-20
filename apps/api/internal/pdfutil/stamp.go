package pdfutil

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
)

// drawCheckbox stamps a check mark when the field's value is true.
func drawCheckbox(pdf *fpdf.Fpdf, r rect) {
	if r.value != "true" {
		return
	}
	pdf.SetDrawColor(30, 30, 30)
	lw := r.h * 0.08
	if lw < 1 {
		lw = 1
	}
	pdf.SetLineWidth(lw)
	pdf.Line(r.x+r.w*0.2, r.y+r.h*0.55, r.x+r.w*0.42, r.y+r.h*0.78)
	pdf.Line(r.x+r.w*0.42, r.y+r.h*0.78, r.x+r.w*0.8, r.y+r.h*0.22)
}

// drawSignatureImage stamps a base64 data-URL image of a signature inside the
// field's bounds. A malformed payload is dropped: the overlay must never fail
// the whole flatten because of one bad field. The image name is unique within
// one overlay document, so the counter is created per overlay instead of
// shared across renders and requests.
func drawSignatureImage(pdf *fpdf.Fpdf, dataURL string, r rect, sigSeq *int) {
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

	*sigSeq++
	name := fmt.Sprintf("sig_%d", *sigSeq)

	reader := bytes.NewReader(imgBytes)
	opts := fpdf.ImageOptions{ImageType: imageType}
	pdf.RegisterImageOptionsReader(name, opts, reader)

	padding := r.h * 0.05
	imgX := r.x + padding
	imgY := r.y + padding
	imgW := r.w - padding*2
	imgH := r.h - padding*2

	pdf.ImageOptions(name, imgX, imgY, imgW, imgH, false, opts, 0, "")
}

// drawField stamps a single field's value at the field's bounds on the
// current page.
func drawField(pdf *fpdf.Fpdf, field FieldOverlay, r rect, sigSeq *int) {
	pdf.SetTextColor(0, 0, 0)

	switch field.FieldType {
	case "signature":
		if strings.HasPrefix(field.Value, "data:image/") {
			drawSignatureImage(pdf, field.Value, r, sigSeq)
		} else {
			fontSize := clampFontSize(r.h*0.55, 6, 24)
			pdf.SetFont(UnicodeFontFamily, "I", fontSize)
			pdf.Text(r.x+2, r.y+r.h/2+fontSize/3, field.Value)
		}

	case "checkbox":
		drawCheckbox(pdf, r)

	default:
		fontSize := clampFontSize(r.h*0.55, 6, 14)
		pdf.SetFont(UnicodeFontFamily, "", fontSize)
		text := truncateToFit(pdf, field.Value, r.w-4)
		pdf.Text(r.x+2, r.y+r.h/2+fontSize/3, text)
	}
}

// truncateToFit shortens text with an ellipsis until it fits maxW.
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

// clampFontSize keeps a computed font size inside readable bounds.
func clampFontSize(size, min, max float64) float64 {
	if size < min {
		return min
	}
	if size > max {
		return max
	}
	return size
}
