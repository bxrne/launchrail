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

// compareQuaternions compares two quaternions, considering the possibility of opposite signs.
func compareQuaternions(q1, q2 types.Quaternion) bool {
	return (almostEqual(q1.X, q2.X) && almostEqual(q1.Y, q2.Y) &&
		almostEqual(q1.Z, q2.Z) && almostEqual(q1.W, q2.W)) ||
		(almostEqual(q1.X, -q2.X) && almostEqual(q1.Y, -q2.Y) &&
			almostEqual(q1.Z, -q2.Z) && almostEqual(q1.W, -q2.W))
}

// vectorsEqual compares two vectors component-wise.
func vectorsEqual(v1, v2 types.Vector3) bool {
	return almostEqual(v1.X, v2.X) && almostEqual(v1.Y, v2.Y) && almostEqual(v1.Z, v2.Z)
}

// TEST: GIVEN a quaternion, WHEN NewQuaternion is called, THEN the quaternion should be created with the correct values.
func TestNewQuaternion(t *testing.T) {
	q := types.NewQuaternion(1, 2, 3, 4)
	if !almostEqual(q.X, 1) || !almostEqual(q.Y, 2) ||
		!almostEqual(q.Z, 3) || !almostEqual(q.W, 4) {
		t.Errorf("NewQuaternion did not set fields correctly: got %+v", q)
	}
}

// TEST: GIVEN a quaternion, WHEN IdentityQuaternion is called, THEN the quaternion should be the identity quaternion.
func TestIdentityQuaternion(t *testing.T) {
	q := types.IdentityQuaternion()
	expected := &types.Quaternion{X: 0, Y: 0, Z: 0, W: 1}
	if !quaternionsEqual(q, expected) {
		t.Errorf("IdentityQuaternion returned %+v, expected %+v", q, expected)
	}
}

// TEST: GIVEN two quaternions, WHEN Add is called, THEN the result should be the sum of the two quaternions.
func TestAdd(t *testing.T) {
	q1 := types.NewQuaternion(1, 2, 3, 4)
	q2 := types.NewQuaternion(5, 6, 7, 8)
	result := q1.Add(q2)
	expected := &types.Quaternion{X: 6, Y: 8, Z: 10, W: 12}
	if !quaternionsEqual(result, expected) {
		t.Errorf("Add: got %+v, expected %+v", result, expected)
	}
}

// TEST: GIVEN two quaternions, WHEN Subtract is called, THEN the result should be the difference of the two quaternions.
func TestSubtract(t *testing.T) {
	q1 := types.NewQuaternion(5, 7, 9, 11)
	q2 := types.NewQuaternion(1, 2, 3, 4)
	result := q1.Subtract(q2)
	expected := &types.Quaternion{X: 4, Y: 5, Z: 6, W: 7}
	if !quaternionsEqual(result, expected) {
		t.Errorf("Subtract: got %+v, expected %+v", result, expected)
	}
}

// TEST: GIVEN a quaternion, WHEN Multiply is called with the identity quaternion, THEN the result should be the same quaternion.
func TestMultiplyWithIdentity(t *testing.T) {
	q := types.NewQuaternion(1, 2, 3, 4)
	identity := types.IdentityQuaternion()
	result := q.Multiply(identity)
	if !quaternionsEqual(result, q) {
		t.Errorf("Multiply with identity: got %+v, expected %+v", result, q)
	}
}

// TEST: GIVEN two quaternions, WHEN Multiply is called, THEN the result should be the product of the two quaternions.
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

// TEST: GIVEN a quaternion and a scalar, WHEN Scale is called, THEN the result should be the quaternion scaled by the scalar.
func TestScale(t *testing.T) {
	q := types.NewQuaternion(1, -2, 3, -4)
	scalar := 2.0
	result := q.Scale(scalar)
	expected := &types.Quaternion{X: 2, Y: -4, Z: 6, W: -8}
	if !quaternionsEqual(result, expected) {
		t.Errorf("Scale: got %+v, expected %+v", result, expected)
	}
}

// TEST: GIVEN a quaternion, WHEN Magnitude is called, THEN the result should be the magnitude of the quaternion.
func TestMagnitude(t *testing.T) {
	q := types.NewQuaternion(1, 2, 3, 4)
	result := q.Magnitude()
	expected := float64(1*1 + 2*2 + 3*3 + 4*4)
	if !almostEqual(result, expected) {
		t.Errorf("Magnitude: got %f, expected %f", result, expected)
	}
}

// TEST: GIVEN a quaternion, WHEN Normalize is called, THEN the result should be the normalized quaternion.
func TestNormalize(t *testing.T) {
	t.Run("Normalize", func(t *testing.T) {
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
	})

	t.Run("Normalize Zero Magnitude", func(t *testing.T) {
		q := &types.Quaternion{X: 0, Y: 0, Z: 0, W: 0}
		expected := &types.Quaternion{X: 0, Y: 0, Z: 0, W: 1} // Expect identity for zero magnitude
		got := q.Normalize()
		if !compareQuaternions(*got, *expected) { // Use compareQuaternions
			t.Errorf("Normalize zero magnitude: got %v, expected %v", got, expected)
		}
	})

	t.Run("Normalize Near-Zero Magnitude", func(t *testing.T) {
		q := &types.Quaternion{X: 1e-15, Y: 1e-15, Z: 1e-15, W: 1e-15}
		expected := &types.Quaternion{X: 0, Y: 0, Z: 0, W: 1} // Expect identity
		got := q.Normalize()
		if !compareQuaternions(*got, *expected) { // Use compareQuaternions
			t.Errorf("Normalize near-zero magnitude: got %v, expected %v", got, expected)
		}
	})

	t.Run("Normalize NaN Magnitude Squared", func(t *testing.T) {
		q := &types.Quaternion{X: math.NaN(), Y: 1, Z: 1, W: 1}
		expected := &types.Quaternion{X: 0, Y: 0, Z: 0, W: 1} // Expect identity for invalid magnitude
		got := q.Normalize()
		if !compareQuaternions(*got, *expected) {
			t.Errorf("Normalize NaN magnitude squared: got %v, expected %v", got, expected)
		}
	})

	t.Run("Normalize Inf Magnitude Squared", func(t *testing.T) {
		q := &types.Quaternion{X: math.Inf(1), Y: 1, Z: 1, W: 1}
		expected := &types.Quaternion{X: 0, Y: 0, Z: 0, W: 1} // Expect identity for invalid magnitude
		got := q.Normalize()
		if !compareQuaternions(*got, *expected) {
			t.Errorf("Normalize Inf magnitude squared: got %v, expected %v", got, expected)
		}
	})
}

// TEST: GIVEN a quaternion and a vector, WHEN RotateVector is called, THEN the result should be the vector rotated by the quaternion.
func TestRotateVector(t *testing.T) {
	// Test a known rotation:
	// For q = (0,0,0,1) and v = (1,0,0),
	// the expected result is (1,0,0)
	q := types.IdentityQuaternion()
	v := &types.Vector3{X: 1, Y: 0, Z: 0} // Use pointer
	result := q.RotateVector(v)
	if !vectorsEqual(*result, *v) { // Dereference for comparison
		t.Errorf("RotateVector: got %+v, expected %+v", result, v)
	}

	t.Run("RotateVector with NaN Quaternion", func(t *testing.T) {
		q := &types.Quaternion{X: math.NaN(), Y: 0, Z: 0, W: 1}
		v := &types.Vector3{X: 1, Y: 0, Z: 0} // Use pointer
		expected := v                         // Expect original vector (pointer)
		got := q.RotateVector(v)
		if !vectorsEqual(*got, *expected) { // Dereference for comparison
			t.Errorf("RotateVector with NaN quaternion: got %v, expected %v", got, expected)
		}
	})

	t.Run("RotateVector with Inf Quaternion", func(t *testing.T) {
		q := &types.Quaternion{X: math.Inf(1), Y: 0, Z: 0, W: 1}
		v := &types.Vector3{X: 1, Y: 0, Z: 0} // Use pointer
		expected := v                         // Expect original vector (pointer)
		got := q.RotateVector(v)
		if !vectorsEqual(*got, *expected) { // Dereference for comparison
			t.Errorf("RotateVector with Inf quaternion: got %v, expected %v", got, expected)
		}
	})

	t.Run("RotateVector with Zero Quaternion", func(t *testing.T) {
		q := &types.Quaternion{X: 0, Y: 0, Z: 0, W: 0} // Normalizes to identity
		v := &types.Vector3{X: 1, Y: 2, Z: 3}          // Use pointer
		expected := v                                  // Expect original vector (pointer)
		got := q.RotateVector(v)
		if !vectorsEqual(*got, *expected) { // Dereference for comparison
			t.Errorf("RotateVector with Zero quaternion: got %v, expected %v", got, expected)
		}
	})
}

// TEST: GIVEN a quaternion and angular velocity, WHEN Integrate is called, THEN the quaternion should be updated correctly.
func TestIntegrate(t *testing.T) {
	// Test a known integration:
	// For q = (0,0,0,1) and w = (0,0,0),
	// the expected result is (0,0,0,1)
	q := types.IdentityQuaternion()
	w := types.Vector3{X: 0, Y: 0, Z: 0}
	dt := 1.0
	result := q.Integrate(w, dt)
	expected := &types.Quaternion{X: 0, Y: 0, Z: 0, W: 1}
	if !quaternionsEqual(result, expected) {
		t.Errorf("Integrate: got %+v, expected %+v", result, expected)
	}

	t.Run("Integrate Near-Zero Angular Velocity", func(t *testing.T) {
		q := types.NewQuaternion(1, 0, 0, 0).Normalize() // Start with identity
		angularVelocity := types.Vector3{X: 1e-15, Y: 1e-15, Z: 1e-15}
		dt := 0.1
		expected := *q // Expect almost no change
		got := q.Integrate(angularVelocity, dt)
		if !compareQuaternions(*got, expected) { // Allow small difference
			t.Errorf("Integrate near-zero angVel: got %v, expected %v", got, expected)
		}
	})
}

// TEST: GIVEN a quaternion, WHEN Conjugate is called, THEN the conjugate should be returned.
func TestConjugate(t *testing.T) {
	q := types.NewQuaternion(1, 2, 3, 4)
	conjugate := q.Conjugate()
	expected := &types.Quaternion{X: -1, Y: -2, Z: -3, W: 4}
	if !quaternionsEqual(conjugate, expected) {
		t.Errorf("Conjugate: got %+v, expected %+v", conjugate, expected)
	}
}
