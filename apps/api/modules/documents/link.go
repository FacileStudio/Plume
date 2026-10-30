package documents

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/FacileStudio/tronc/errors"
)

const maxRefLength = 200

var facileIDPattern = regexp.MustCompile(`^fac_[0-9a-f]{24}$`)

// Link carries the optional fields an app that asks Plume for signatures
// attaches to a document: its own reference and where the signed PDF is
// archived in Nuage.
type Link struct {
	ExternalRef        string `json:"external_ref,omitempty"`
	NuageSpaceFacileID string `json:"nuage_space_facile_id,omitempty"`
	NuageLinkedTo      string `json:"nuage_linked_to,omitempty"`
}

// Validate checks every set field and leaves unset ones alone.
func (l Link) Validate() error {
	if err := validateRef("external_ref", l.ExternalRef); err != nil {
		return err
	}
	if err := validateRef("nuage_linked_to", l.NuageLinkedTo); err != nil {
		return err
	}
	if l.NuageSpaceFacileID != "" && !facileIDPattern.MatchString(l.NuageSpaceFacileID) {
		return errors.Invalid("nuage_space_facile_id must be fac_ followed by 24 lowercase hex characters")
	}
	return nil
}

func validateRef(name, value string) error {
	if value == "" {
		return nil
	}
	if utf8.RuneCountInString(value) > maxRefLength {
		return errors.Invalid(name + " must be at most 200 characters")
	}
	if strings.TrimSpace(value) != value {
		return errors.Invalid(name + " must not start or end with whitespace")
	}
	for _, r := range value {
		if !unicode.IsPrint(r) {
			return errors.Invalid(name + " must contain printable characters only")
		}
	}
	return nil
}
