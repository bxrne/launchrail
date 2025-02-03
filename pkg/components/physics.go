package components

import "github.com/EngoEngine/ecs"

// Position represents a 3D position
type Position struct {
	ecs.BasicEntity
	X, Y, Z float64
}

// Velocity represents a 3D velocity
type Velocity struct {
	ecs.BasicEntity
	X, Y, Z float64
}

// Acceleration represents a 3D acceleration
type Acceleration struct {
	ecs.BasicEntity
	X, Y, Z float64
}

// Mass represents a mass value
type Mass struct {
	ecs.BasicEntity
	Value float64
}
