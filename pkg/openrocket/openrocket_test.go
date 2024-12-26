package openrocket_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/openrocket"
)

// TEST: GIVEN a valid OpenRocket file WHEN Load is called THEN no error is returned
func TestLoad(t *testing.T) {
	// Path to the test .ork file
	testFilePath := "../../testdata/openrocket/l1.ork"

	// Call the Load function
	_, err := openrocket.Load(testFilePath, "23.09")
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
}

// TEST: GIVEN an invalid OpenRocket file WHEN Load is called THEN an error is returned
func TestLoadInvalidFile(t *testing.T) {
	// Path to a non-existent file
	testFilePath := "nonexistent.ork"

	// Call the Load function
	_, err := openrocket.Load(testFilePath, "23.09")
	if err == nil {
		t.Fatalf("Load did not return an error")
	}
}

// TEST: GIVEN an invalid OpenRocket version WHEN Load is called THEN an error is returned
func TestLoadInvalidVersion(t *testing.T) {
	// Path to the test .ork file
	testFilePath := "../../testdata/openrocket/l1.ork"

	// Call the Load function with an invalid version
	_, err := openrocket.Load(testFilePath, "invalid")
	if err == nil {
		t.Fatalf("Load did not return an error")
	}
}
