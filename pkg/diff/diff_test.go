package diff_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/bxrne/launchrail/pkg/diff"
)

// TEST: Given two byte slices, when CombinedHash is called, then it returns the hash of the two byte slices
func TestCombinedHash(t *testing.T) {
	// Create two byte slices
	a := []byte("hello")
	b := []byte("world")

	// Call CombinedHash
	got := diff.CombinedHash(a, b)

	// Create the expected hash
	hash := sha256.New()
	hash.Write(a)
	hash.Write(b)
	want := fmt.Sprintf("%x", hash.Sum(nil))

	// Compare the hashes
	if got != want {
		t.Errorf("CombinedHash = %q, want %q", got, want)
	}
}
