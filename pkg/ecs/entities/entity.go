package entities

import "github.com/bxrne/launchrail/pkg/ecs/components"

// Entity represents a unique identifier for game objects
type Entity interface {
	String() string   // full Key=val breakdown
	Describe() string // short description
	AddComponent(c components.Component)
	RemoveComponent(c components.Component)
	Update(dt float64) error
}
