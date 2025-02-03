package components

import "github.com/EngoEngine/ecs"

type Position struct {
	ecs.BasicEntity
	X, Y, Z float64
}

type Velocity struct {
	ecs.BasicEntity
	X, Y, Z float64
}

type Acceleration struct {
	ecs.BasicEntity
	X, Y, Z float64
}

type Mass struct {
	ecs.BasicEntity
	Value float64
}
