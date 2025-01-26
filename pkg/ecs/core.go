package ecs

import "github.com/bxrne/launchrail/pkg/ecs/types"

// Component is the base interface for all components
type Component interface {
	Type() string
	Update(dt float64) error
}

// System is the base interface for all systems
type System interface {
	Update(world *World, dt float64) error
	Priority() int
}

// Common component interfaces for type safety
type PhysicsComponent interface {
	Component
	AddForce(force types.Vector3)
	GetVelocity() types.Vector3
	GetMass() float64
}

type MotorComponent interface {
	Component
	GetThrust() float64
	GetMass() float64
}

type AerodynamicsComponent interface {
	Component
	CalculateDrag(velocity types.Vector3) types.Vector3
}

// Registry for component types
const (
	ComponentPhysics      = "Physics"
	ComponentMotor        = "Motor"
	ComponentAerodynamics = "Aerodynamics"
	ComponentNosecone     = "Nosecone"
)
