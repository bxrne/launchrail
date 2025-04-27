package entities

import (
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
)

// RocketEntity represents a complete rocket with all its components
type RocketEntity struct {
	*ecs.BasicEntity
	*states.PhysicsState
	*types.Mass
	components map[string]interface{}
	mu         sync.RWMutex
}

// NewRocketEntity creates a new rocket entity from OpenRocket data
func NewRocketEntity(world *ecs.World, orkData *openrocket.RocketDocument, motor *components.Motor) *RocketEntity {
	if orkData == nil || motor == nil {
		return nil
	}

	basic := ecs.NewBasic()

	// Create base rocket entity with non-zero initial values
	rocket := &RocketEntity{
		BasicEntity: &basic,
		PhysicsState: &states.PhysicsState{
			Position: &types.Position{
				BasicEntity: basic,
				Vec:         types.Vector3{X: 0, Y: 0, Z: 0},
			},
			Velocity: &types.Velocity{
				BasicEntity: basic,
				Vec:         types.Vector3{X: 0, Y: 0, Z: 0},
			},
			Acceleration: &types.Acceleration{
				BasicEntity: basic,
				Vec:         types.Vector3{X: 0, Y: -9.81, Z: 0}, // Initialize with gravity
			},
			Orientation: &types.Orientation{
				BasicEntity: basic,
				Quat:        *types.IdentityQuaternion(), // Initialize with identity quaternion
			},
			AngularAcceleration: &types.Vector3{}, // Initialize angular acceleration
			AngularVelocity:     &types.Vector3{}, // Initialize angular velocity
		},
		Mass: &types.Mass{
			BasicEntity: basic,
			Value:       calculateTotalMass(orkData), // Set mass first
		},
		components: make(map[string]interface{}),
	}

	// Validate mass
	if rocket.Mass.Value <= 0 {
		return nil
	}

	// Store components with proper error handling
	rocket.components["motor"] = motor

	bodytube, err := components.NewBodytubeFromORK(ecs.NewBasic(), orkData)
	if err != nil {
		return nil
	}
	rocket.components["bodytube"] = bodytube

	nosecone := components.NewNoseconeFromORK(ecs.NewBasic(), orkData)
	if nosecone == nil {
		return nil
	}
	rocket.components["nosecone"] = nosecone

	finset := components.NewTrapezoidFinsetFromORK(ecs.NewBasic(), orkData)
	if finset == nil {
		return nil
	}
	rocket.components["finset"] = finset

	parachute, err := components.NewParachuteFromORK(ecs.NewBasic(), orkData)
	if err != nil {
		panic(err)
	}
	rocket.components["parachute"] = parachute

	return rocket
}

// Calculate total mass by summing all components from OpenRocket data
func calculateTotalMass(orkData *openrocket.RocketDocument) float64 {
	if orkData == nil || len(orkData.Subcomponents.Stages) == 0 {
		// Consider logging an error here
		return 0.0
	}

	var totalMass float64
	// Assuming single stage - validated in OpenrocketDocument.Validate
	stage := orkData.Subcomponents.Stages[0]

	// Helper for components with pointer receiver GetMass()
	addMass := func(component interface{ GetMass() float64 }) {
		// component is expected to be a non-nil pointer to the struct
		if component != nil {
			totalMass += component.GetMass()
			// TODO: Potentially check for NaN/negative mass from GetMass() implementations
		}
	}

	// --- Add mass for components with GetMass() ---
	sustainerSubs := stage.SustainerSubcomponents // Value struct

	// Pass addresses (&) to satisfy pointer receivers for GetMass
	addMass(&sustainerSubs.Nosecone)
	addMass(&sustainerSubs.BodyTube)

	// Process subcomponents of BodyTube
	bodyTubeSubs := sustainerSubs.BodyTube.Subcomponents // Value struct
	addMass(&bodyTubeSubs.TrapezoidFinset)

	// --- Add mass for components with direct Mass field ---
	noseSubs := sustainerSubs.Nosecone.Subcomponents
	totalMass += noseSubs.MassComponent.Mass // Assumes MassComponent always exists if NoseSubcomponents does

	// --- Components currently omitted due to missing GetMass() or data in schema ---
	// - Parachute: No GetMass() in schema_parachute.go
	// - Shockcord: No GetMass() in schema_common.go
	// - InnerTube: No GetMass() in schema_airframe.go
	// - CenteringRing: No GetMass() in schema_common.go (would need iteration if added)
	// - MotorMount/Motor: No GetMass() in schema_motor.go (Motor mass handled dynamically)
	// NOTE: This means the calculated mass will be an UNDERESTIMATE.
	// To fix this, GetMass() methods or manual calculations based on
	// material/dimensions need to be added to the respective schema types.

	// NOTE: Calculation also relies on the correctness of existing GetMass methods.
	// As observed, Nosecone.GetMass and BodyTube.GetMass implementations might be inaccurate.

	return totalMass
}

// AddComponent adds a component to the entity
func (r *RocketEntity) GetComponent(name string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.components[name]
}
