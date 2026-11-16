// Package secrets protects the three values Plume keeps for later use: the
// signing link token of a signer, an owner's SMTP password and a webhook's
// HMAC secret.
//
// A signing token is looked up by the SHA-256 hex of its value, stored in
// signers.token_hash, whether or not a key is configured. The token itself
// stays in signers.token, sealed with porte's token sealer when
// PORTE_TOKEN_KEY is set, because reminders and the owner's copy-link need it
// back. The SMTP password and the webhook secret are sealed the same way.
// With no key the sealer is nil and those values stay as they were.
package secrets

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/FacileStudio/Plume/apps/api/schemas"

	portepg "github.com/FacileStudio/porte/pg"
	"gorm.io/gorm"
)

const sealedPrefix = "pv1:"

// ErrNoToken is returned for an empty signing token, which matches nothing.
var ErrNoToken = errors.New("secrets: empty signing token")

// HashToken is the SHA-256 hex of a signing token, the value stored in
// signers.token_hash.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Sealed reports whether a stored value carries the sealer's prefix.
func Sealed(stored string) bool {
	return strings.HasPrefix(stored, sealedPrefix)
}

// legacyToken reports whether input could be a plaintext token from a row not
// converted yet. A sealed value read from the database always starts with the
// prefix and contains a colon, so it can never pass.
func legacyToken(input string) bool {
	return !Sealed(input) && !strings.Contains(input, ":")
}

// SealToken returns what to store in signers.token and signers.token_hash for
// a fresh token.
func SealToken(sealer *portepg.TokenSealer, token string) (stored, hash string, err error) {
	stored, err = sealer.Seal(token)
	if err != nil {
		return "", "", err
	}
	return stored, HashToken(token), nil
}

// FindSigner loads the signer whose link token is token. It matches on
// token_hash. Until a row is converted it falls back to the plaintext token
// column, and only for an input that is not a sealed value.
func FindSigner(db *gorm.DB, token string) (schemas.Signer, error) {
	var signer schemas.Signer
	if token == "" {
		return signer, gorm.ErrRecordNotFound
	}
	err := db.Where("token_hash = ?", HashToken(token)).First(&signer).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) || !legacyToken(token) {
		return signer, err
	}
	return signer, db.Where("token = ? AND (token_hash IS NULL OR token_hash = '')", token).First(&signer).Error
}

// OpenToken returns the usable token behind a stored signers.token, or the
// empty string when it is empty or no configured key opens it.
func OpenToken(sealer *portepg.TokenSealer, stored string) string {
	plain, err := sealer.Open(stored)
	if err != nil {
		return ""
	}
	return plain
}

// Reveal replaces the stored token of each signer with the usable one, in
// memory. The revealed value must never be written back to the database.
func Reveal(sealer *portepg.TokenSealer, signers ...*schemas.Signer) {
	for _, signer := range signers {
		signer.Token = OpenToken(sealer, signer.Token)
	}
}
