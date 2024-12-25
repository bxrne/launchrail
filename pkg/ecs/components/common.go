package components

import "github.com/bxrne/launchrail/pkg/ecs/types"

// TransformComponent handles position, rotation, and scale
type TransformComponent struct {
	Position types.Vector3
	Rotation types.Vector3
	Scale    types.Vector3
}

// Type returns the type of the component
func (t *TransformComponent) Type() string { return "Transform" }

// PhysicsComponent handles physical properties
type PhysicsComponent struct {
	Mass      float64         // kg
	Velocity  types.Vector3   // m/s
	Forces    []types.Vector3 // N
	Drag      float64         // Coefficient
	Thrust    float64         // N
	MotorType string
}

func (p *PhysicsComponent) Type() string { return "Physics" }

// AerodynamicsComponent handles aerodynamic properties
type AerodynamicsComponent struct {
	CenterOfPressure types.Vector3
	CenterOfMass     types.Vector3
	CrossSection     float64 // m²
	StabilityMargin  float64 // calibers
}

// Type returns the type of the component
func (a *AerodynamicsComponent) Type() string { return "Aerodynamics" }
