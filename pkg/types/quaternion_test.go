package types_test

import (
	"math"
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
)

const epsilon = 1e-9

// almostEqual checks if two floats are nearly equal.
func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// quaternionsEqual compares two quaternions component-wise.
func quaternionsEqual(q1, q2 *types.Quaternion) bool {
	return almostEqual(q1.X, q2.X) && almostEqual(q1.Y, q2.Y) &&
		almostEqual(q1.Z, q2.Z) && almostEqual(q1.W, q2.W)
}

func TestNewQuaternion(t *testing.T) {
	q := types.NewQuaternion(1, 2, 3, 4)
	if !almostEqual(q.X, 1) || !almostEqual(q.Y, 2) ||
		!almostEqual(q.Z, 3) || !almostEqual(q.W, 4) {
		t.Errorf("NewQuaternion did not set fields correctly: got %+v", q)
	}
}

func TestIdentityQuaternion(t *testing.T) {
	q := types.IdentityQuaternion()
	expected := &types.Quaternion{X: 0, Y: 0, Z: 0, W: 1}
	if !quaternionsEqual(q, expected) {
		t.Errorf("IdentityQuaternion returned %+v, expected %+v", q, expected)
	}
}

func TestAdd(t *testing.T) {
	q1 := types.NewQuaternion(1, 2, 3, 4)
	q2 := types.NewQuaternion(5, 6, 7, 8)
	result := q1.Add(q2)
	expected := &types.Quaternion{X: 6, Y: 8, Z: 10, W: 12}
	if !quaternionsEqual(result, expected) {
		t.Errorf("Add: got %+v, expected %+v", result, expected)
	}
}

func TestSubtract(t *testing.T) {
	q1 := types.NewQuaternion(5, 7, 9, 11)
	q2 := types.NewQuaternion(1, 2, 3, 4)
	result := q1.Subtract(q2)
	expected := &types.Quaternion{X: 4, Y: 5, Z: 6, W: 7}
	if !quaternionsEqual(result, expected) {
		t.Errorf("Subtract: got %+v, expected %+v", result, expected)
	}
}

func TestMultiplyWithIdentity(t *testing.T) {
	q := types.NewQuaternion(1, 2, 3, 4)
	identity := types.IdentityQuaternion()
	result := q.Multiply(identity)
	if !quaternionsEqual(result, q) {
		t.Errorf("Multiply with identity: got %+v, expected %+v", result, q)
	}
}

func TestMultiply(t *testing.T) {
	// Test a known multiplication:
	// For q1 = (0,1,0,0) and q2 = (0,0,1,0),
	// the expected result is (1,0,0,0)
	q1 := types.NewQuaternion(0, 1, 0, 0)
	q2 := types.NewQuaternion(0, 0, 1, 0)
	result := q1.Multiply(q2)
	expected := &types.Quaternion{X: 1, Y: 0, Z: 0, W: 0}
	if !quaternionsEqual(result, expected) {
		t.Errorf("Multiply: got %+v, expected %+v", result, expected)
	}
}

func TestScale(t *testing.T) {
	q := types.NewQuaternion(1, -2, 3, -4)
	scalar := 2.0
	result := q.Scale(scalar)
	expected := &types.Quaternion{X: 2, Y: -4, Z: 6, W: -8}
	if !quaternionsEqual(result, expected) {
		t.Errorf("Scale: got %+v, expected %+v", result, expected)
	}
}

func TestMagnitude(t *testing.T) {
	q := types.NewQuaternion(1, 2, 3, 4)
	result := q.Magnitude()
	expected := float64(1*1 + 2*2 + 3*3 + 4*4)
	if !almostEqual(result, expected) {
		t.Errorf("Magnitude: got %f, expected %f", result, expected)
	}
}

func TestNormalize(t *testing.T) {
	q := types.NewQuaternion(1, 2, 3, 4)
	normalized := q.Normalize()
	// Correct normalization: each component divided by the square root of the magnitude.
	mag := math.Sqrt(q.Magnitude())
	expected := &types.Quaternion{
		X: q.X / mag,
		Y: q.Y / mag,
		Z: q.Z / mag,
		W: q.W / mag,
	}
	if !quaternionsEqual(normalized, expected) {
		t.Errorf("Normalize: got %+v, expected %+v", normalized, expected)
	}
}
