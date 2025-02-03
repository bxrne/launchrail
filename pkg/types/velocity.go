package types

import "github.com/EngoEngine/ecs"

// Velocity represents a 3D velocity
type Velocity struct {
	ecs.BasicEntity
	Vec Vector3
}
