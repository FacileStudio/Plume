package documents

import (
	"net/mail"

	"github.com/FacileStudio/tronc/errors"
)

const (
	maxSignatureSigners  = 20
	maxClientFieldLength = 200
)

var signerRoles = map[string]bool{"signer": true, "viewer": true, "approver": true}

var signatureFieldTypes = map[string]bool{"signature": true, "text": true, "date": true, "checkbox": true}

func (r *SignatureRequest) validate() error {
	if r.Name == "" {
		return errors.Invalid("name is required")
	}
	if err := r.Link.Validate(); err != nil {
		return err
	}
	if len(r.Signers) == 0 || len(r.Signers) > maxSignatureSigners {
		return errors.Invalid("signers must hold between 1 and 20 entries")
	}
	for i := range r.Signers {
		if err := r.Signers[i].validate(); err != nil {
			return err
		}
	}
	if r.Client != nil {
		if err := r.Client.validate(); err != nil {
			return err
		}
	}
	for i := range r.Fields {
		if err := r.Fields[i].validate(len(r.Signers)); err != nil {
			return err
		}
	}
	return nil
}

func (c *SignatureClient) validate() error {
	if c.Name == "" || len(c.Name) > maxClientFieldLength || len(c.Company) > maxClientFieldLength {
		return errors.Invalid("client needs a name of at most 200 characters")
	}
	if _, err := mail.ParseAddress(c.Email); err != nil {
		return errors.Invalid("client needs a valid email")
	}
	return nil
}

func (s *SignatureSigner) validate() error {
	if s.Name == "" {
		return errors.Invalid("every signer needs a name")
	}
	addr, err := mail.ParseAddress(s.Email)
	if err != nil || addr.Address != s.Email {
		return errors.Invalid("every signer needs a valid email")
	}
	if s.Role == "" {
		s.Role = "signer"
	}
	if !signerRoles[s.Role] {
		return errors.Invalid("role must be one of: signer, viewer, approver")
	}
	return nil
}

func (f *SignatureField) validate(signerCount int) error {
	if f.Signer < 0 || f.Signer >= signerCount {
		return errors.Invalid("field signer must be the index of a signer in the request")
	}
	if !signatureFieldTypes[f.FieldType] {
		return errors.Invalid("field_type must be one of: signature, text, date, checkbox")
	}
	if f.Page < 1 {
		return errors.Invalid("field page starts at 1")
	}
	inRange := func(v float64) bool { return v >= 0 && v <= 100 }
	if !inRange(f.X) || !inRange(f.Y) || !inRange(f.Width) || !inRange(f.Height) || f.Width == 0 || f.Height == 0 {
		return errors.Invalid("field x, y, width and height are percentages of the page between 0 and 100")
	}
	return nil
}

func (r *SignatureRequest) shouldSend() bool {
	return r.Send == nil || *r.Send
}
