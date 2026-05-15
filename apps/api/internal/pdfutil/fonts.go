package pdfutil

import (
	_ "embed"

	"github.com/go-pdf/fpdf"
)

//go:embed fonts/DejaVuSans.ttf
var dejaVuRegular []byte

//go:embed fonts/DejaVuSans-Bold.ttf
var dejaVuBold []byte

//go:embed fonts/DejaVuSans-Oblique.ttf
var dejaVuItalic []byte

//go:embed fonts/DejaVuSans-BoldOblique.ttf
var dejaVuBoldItalic []byte

// UnicodeFontFamily is the registered family name shared across the API.
const UnicodeFontFamily = "DejaVu"

// RegisterUnicodeFonts attaches the embedded DejaVu Sans family to the given
// fpdf instance so that French accents, dashes, bullets and other non-ASCII
// glyphs render correctly. Safe to call multiple times.
func RegisterUnicodeFonts(pdf *fpdf.Fpdf) {
	pdf.AddUTF8FontFromBytes(UnicodeFontFamily, "", dejaVuRegular)
	pdf.AddUTF8FontFromBytes(UnicodeFontFamily, "B", dejaVuBold)
	pdf.AddUTF8FontFromBytes(UnicodeFontFamily, "I", dejaVuItalic)
	pdf.AddUTF8FontFromBytes(UnicodeFontFamily, "BI", dejaVuBoldItalic)
}
