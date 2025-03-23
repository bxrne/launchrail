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
				BasicEntity:     basic,
				AngularVelocity: types.Vector3{X: 0, Y: 0, Z: 0},
			},
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
	var totalMass float64

	// Get masses from OpenRocket components
	nosecone := orkData.Subcomponents.Stages[0].SustainerSubcomponents.Nosecone
	bodytube := orkData.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube

	totalMass += nosecone.GetMass() + bodytube.GetMass()
	// Add other component masses...

	return totalMass
}

// AddComponent adds a component to the entity
func (r *RocketEntity) GetComponent(name string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.components[name]
}
