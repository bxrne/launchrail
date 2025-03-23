package types

import "github.com/EngoEngine/ecs"

// Orientation represents a 3D orientation
type Orientation struct {
	ecs.BasicEntity
	Quat Quaternion
}

// Integrates the angular velocity to the orientation
func (o *Orientation) Integrate(angularVelocity Vector3, dt float64) {
	// Integrate quaternion
	o.Quat = *o.Quat.Multiply(&Quaternion{
		W: 0,
		X: angularVelocity.X * dt,
		Y: angularVelocity.Y * dt,
		Z: angularVelocity.Z * dt,
	}).Add(&o.Quat).Normalize()
}
