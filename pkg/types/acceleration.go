package types

import "github.com/EngoEngine/ecs"

// Acceleration represents a 3D acceleration as a vector
type Acceleration struct {
	ecs.BasicEntity
	Vec Vector3
}
