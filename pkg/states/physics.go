package states

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/types"
)

// PhysicsState represents an entity with physics components
type PhysicsState struct {
	// data
	Entity              *ecs.BasicEntity
	Position            *types.Position
	Velocity            *types.Velocity
	Acceleration        *types.Acceleration
	AngularVelocity     *types.Vector3
	AngularAcceleration *types.Vector3
	Orientation         *types.Orientation
	Mass                *types.Mass
	Time                float64

	// Accumulators for forces and moments within a timestep
	AccumulatedForce  types.Vector3
	AccumulatedMoment types.Vector3

	// Inertia Tensors (Body Frame)
	// InertiaTensorBody is the 3x3 inertia tensor in the body frame.
	// InverseInertiaTensorBody is its inverse, also in the body frame.
	InertiaTensorBody        types.Matrix3x3
	InverseInertiaTensorBody types.Matrix3x3

	// Current event detected this timestep
	CurrentEvent types.Event

	// components
	Motor     *components.Motor
	Bodytube  *components.Bodytube
	Nosecone  *components.Nosecone
	Finset    *components.TrapezoidFinset
	Parachute *components.Parachute
}
