package types

import "math"

// Vector3 represents a 3D vector
type Vector3 struct {
	X, Y, Z float64
}

// Add returns the sum of two vectors
func (v Vector3) Add(other Vector3) Vector3 {
	return Vector3{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

// Subtract returns the difference of two vectors
func (v Vector3) Scale(scalar float64) Vector3 {
	return Vector3{
		X: v.X * scalar,
		Y: v.Y * scalar,
		Z: v.Z * scalar,
	}
}

// Magnitude returns the length of the vector
func (v Vector3) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}
