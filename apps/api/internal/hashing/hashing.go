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

func SHA256Reader(reader io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func SHA256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return SHA256Reader(file)
}

func IsValidHex(value string) bool {
	return hexPattern.MatchString(value)
}
