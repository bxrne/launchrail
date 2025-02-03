package types

import (
	"fmt"
	"math"
)

// Vector3 represents a 3D vector
type Vector3 struct {
	X, Y, Z float64
}

// Add returns the sum of two vectors
// INFO: Adding two vectors component-wise.
func (v Vector3) Add(other Vector3) Vector3 {
	return Vector3{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

// Subtract returns the difference of two vectors
// INFO: Subtracting other vector from this vector component-wise.
func (v Vector3) Subtract(other Vector3) Vector3 {
	return Vector3{
		X: v.X - other.X,
		Y: v.Y - other.Y,
		Z: v.Z - other.Z,
	}
}

// Magnitude returns the length of the vector
// INFO: Calculating the magnitude as the Euclidean norm.
func (v Vector3) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// String returns a string representation of the vector
// INFO: Format the vector components to two decimal places for readability.
func (v Vector3) String() string {
	return fmt.Sprintf("Vector3{X: %.2f, Y: %.2f, Z: %.2f}", v.X, v.Y, v.Z)
}

// MultiplyScalar returns the vector multiplied by a scalar
// INFO: Scaling the vector components by the given scalar.
func (v Vector3) MultiplyScalar(scalar float64) Vector3 {
	return Vector3{
		X: v.X * scalar,
		Y: v.Y * scalar,
		Z: v.Z * scalar,
	}
}

// DivideScalar returns the vector divided by a scalar
// INFO: Ensure the scalar is not zero to avoid division by zero.
func (v Vector3) DivideScalar(scalar float64) Vector3 {
	if scalar == 0 {
		// WARN: Avoiding division by zero which would lead to NaN values.
		return v // Return original vector instead of dividing by zero
	}
	return Vector3{
		X: v.X / scalar,
		Y: v.Y / scalar,
		Z: v.Z / scalar,
	}
}

// Round returns the vector with each component rounded to the given precision
// INFO: Rounding the vector components to the supplied precision.
func (v Vector3) Round(precision int) Vector3 {
	prec := math.Pow10(precision)
	return Vector3{
		X: math.Round(v.X*prec) / prec,
		Y: math.Round(v.Y*prec) / prec,
		Z: math.Round(v.Z*prec) / prec,
	}
}
