package openrocket_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/openrocket"
)

func TestLoad(t *testing.T) {
	// Path to the test .ork file
	testFilePath := "../../testdata/openrocket/l1.ork"

	// Call the Load function
	_, err := openrocket.Load(testFilePath)
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
}
