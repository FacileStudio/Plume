package hashing

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"regexp"
)

const HexLength = 64

var hexPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

// SHA256Reader returns the hex-encoded SHA-256 digest of everything read
// from reader.
func SHA256Reader(reader io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// SHA256File returns the hex-encoded SHA-256 digest of the file at path.
func SHA256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return SHA256Reader(file)
}

// IsValidHex reports whether value is a 64-character lowercase hex string.
func IsValidHex(value string) bool {
	return hexPattern.MatchString(value)
}
