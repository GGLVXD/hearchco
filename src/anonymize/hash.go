package anonymize

import (
	"crypto/sha256"
	"encoding/base64"
)

func HashToSHA256B64(orig string) string {
	// hash string with sha256 which returns binary
	hasher := sha256.New()
	hasher.Write([]byte(orig))
	hashedBinary := hasher.Sum(nil)

	// encode binary hash to base64 string
	hashedString := base64.URLEncoding.EncodeToString(hashedBinary)

	return hashedString
}

func HashToSHA256B64Salted(orig string, salt string) string {
	return HashToSHA256B64(orig + salt)
}

func CheckHash(hash string, orig string, salt string) bool {
	return hash == HashToSHA256B64Salted(orig, salt)
}
