package types

import "github.com/EngoEngine/ecs"

// Mass represents a mass value
type Mass struct {
	ecs.BasicEntity
	Value float64
}
