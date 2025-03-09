package types

import "github.com/EngoEngine/ecs"

// Orientation represents a 3D orientation
type Orientation struct {
	ecs.BasicEntity
	Quat            Quaternion
	AngularVelocity Vector3
}
