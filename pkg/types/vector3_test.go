package types_test

import (
	"math"
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
)

// TEST: GIVEN two Vector3 instances, WHEN Add is called, THEN the result should be the sum of the two vectors.
func TestVector3Add(t *testing.T) {
	v1 := types.Vector3{X: 1, Y: 2, Z: 3}
	v2 := types.Vector3{X: 4, Y: 5, Z: 6}
	expected := types.Vector3{X: 5, Y: 7, Z: 9}

	result := v1.Add(v2)
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}

// TEST: GIVEN two Vector3 instances, WHEN Subtract is called, THEN the result should be the difference of the two vectors.
func TestVector3Subtract(t *testing.T) {
	v1 := types.Vector3{X: 4, Y: 5, Z: 6}
	v2 := types.Vector3{X: 1, Y: 2, Z: 3}
	expected := types.Vector3{X: 3, Y: 3, Z: 3}

	result := v1.Subtract(v2)
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}

// TEST: GIVEN a Vector3 instance, WHEN Magnitude is called, THEN the result should be the length of the vector.
func TestVector3Magnitude(t *testing.T) {
	v := types.Vector3{X: 3, Y: 4, Z: 0}
	expected := 5.0 // 3-4-5 triangle

	result := v.Magnitude()
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}

// TEST: GIVEN a Vector3 instance, WHEN String is called, THEN the result should be a string representation of the vector.
func TestVector3String(t *testing.T) {
	v := types.Vector3{X: 1.234567, Y: 2.345678, Z: 3.456789}
	expected := "Vector3{X: 1.23, Y: 2.35, Z: 3.46}"

	result := v.String()
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}

// TEST: GIVEN a Vector3 instance and a scalar, WHEN MultiplyScalar is called, THEN the result should be the vector scaled by the scalar.
func TestVector3MultiplyScalar(t *testing.T) {
	v := types.Vector3{X: 1, Y: 2, Z: 3}
	scalar := 2.0
	expected := types.Vector3{X: 2, Y: 4, Z: 6}

	result := v.MultiplyScalar(scalar)
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}

// TEST: GIVEN a Vector3 instance and a scalar, WHEN DivideScalar is called, THEN the result should be the vector divided by the scalar.
func TestVector3DivideScalar(t *testing.T) {
	v := types.Vector3{X: 2, Y: 4, Z: 6}
	scalar := 2.0
	expected := types.Vector3{X: 1, Y: 2, Z: 3}

	result := v.DivideScalar(scalar)
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}

// TEST: GIVEN a Vector3 instance and a zero scalar, WHEN DivideScalar is called, THEN the result should be the vector divided by the smallest non-zero float.
func TestVector3DivideScalarZero(t *testing.T) {
	v := types.Vector3{X: 2, Y: 4, Z: 6}
	scalar := 0.0

	smallestNonzero := float64(math.SmallestNonzeroFloat64)
	expected := types.Vector3{
		X: 2 / smallestNonzero,
		Y: 4 / smallestNonzero,
		Z: 6 / smallestNonzero,
	}

	result := v.DivideScalar(scalar)
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}

// TEST: GIVEN a Vector3 instance and a scalar, WHEN Round is called, THEN the result should be the vector rounded to the specified number of decimal places.
func TestVector3Round(t *testing.T) {
	v := types.Vector3{X: 1.234567, Y: 2.345678, Z: 3.456789}
	decimalPlaces := 2
	expected := types.Vector3{X: 1.23, Y: 2.35, Z: 3.46}

	result := v.Round(decimalPlaces)
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}

// TEST: GIVEN a Vector3 instance, WHEN Round is called with zero decimal places, THEN the result should be the vector rounded to zero decimal places.
func TestVector3RoundZero(t *testing.T) {
	v := types.Vector3{X: 1.234567, Y: 2.345678, Z: 3.456789}
	decimalPlaces := 0
	expected := types.Vector3{X: 1, Y: 2, Z: 3}

	result := v.Round(decimalPlaces)
	if result != expected {
		t.Errorf("Expected %v but got %v", expected, result)
	}
}
