package diff

import (
	"crypto/sha256"
	"fmt"
)

// combinedHash returns the SHA256 hash of two byte slices
// NOTE: hash(hash(a) + hash(b))
func CombinedHash(a, b []byte) string {
	hash := sha256.New()
	hash.Write(a)
	hash.Write(b)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
