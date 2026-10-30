package pdfutil

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// PageCount returns the number of pages in the PDF at path.
func PageCount(path string) (int, error) {
	configurePdfcpu()
	count, err := api.PageCountFile(path)
	if err != nil {
		return 0, fmt.Errorf("pdfutil: count pages of %q: %w", path, err)
	}
	return count, nil
}
