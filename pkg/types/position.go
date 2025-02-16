package types

import "github.com/EngoEngine/ecs"

// Position represents a 3D position
type Position struct {
	ecs.BasicEntity
	Vec Vector3
}
