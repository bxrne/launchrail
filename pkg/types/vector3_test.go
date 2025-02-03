package types_test

import (
	"math"
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
)

// Generate tests for Vector3 methods
// TEST: These are just references to help generate your test cases.
func TestVector3Operations(t *testing.T) {
	v1 := types.Vector3{X: 3, Y: 4, Z: 5}
	v2 := types.Vector3{X: -1, Y: 2, Z: -3}

	// Test Add
	result := v1.Add(v2)
	expected := types.Vector3{X: 2, Y: 6, Z: 2}
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}

	// Test Subtract
	result = v1.Subtract(v2)
	expected = types.Vector3{X: 4, Y: 2, Z: 8}
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}

	// Test Magnitude
	resultMag := v1.Magnitude()
	expectedMag := math.Sqrt(50)
	if resultMag != expectedMag {
		t.Errorf("Expected %v but got %v", expectedMag, resultMag)
	}

	// Test MultiplyScalar
	result = v1.MultiplyScalar(2)
	expected = types.Vector3{X: 6, Y: 8, Z: 10}
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}

	// Test DivideScalar
	defer func() {
		if r := recover(); r != nil {
			// Expected panic
			t.Logf("Recovered from panic: %v", r)
		}
	}()
	// Proper case
	result = v1.DivideScalar(2)
	expected = types.Vector3{X: 1.5, Y: 2, Z: 2.5}
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}

	// Round to 2 decimal places
	result = v1.Round(2)
	expected = types.Vector3{X: 3, Y: 4, Z: 5}
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}
	// Division by zero should panic
	v1.DivideScalar(0)
}
