package entities

import (
	"fmt"
	"math"
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
	openrocket "github.com/bxrne/launchrail/pkg/openrocket"
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
		fmt.Println("Error: Cannot create RocketEntity, OpenRocket data or motor is nil.")
		return nil
	}

	// --- 1. Create Components First ---
	createdComponents := make(map[string]interface{}) // Use interface{} for flexibility

	// Motor (already created, just validate and add)
	if motor.GetMass() <= 0 {
		motorName := "unknown"
		if motor.Props != nil && motor.Props.Designation != "" { 
			motorName = string(motor.Props.Designation)
		}
		fmt.Printf("Error: Cannot create RocketEntity, motor '%s' has invalid initial mass.\n", motorName)
		return nil
	}
	createdComponents["motor"] = motor // Add validated motor

	// Bodytube
	bodytube, err := components.NewBodytubeFromORK(ecs.NewBasic(), orkData)
	if err != nil {
		fmt.Printf("Error creating Bodytube from ORK: %v\n", err) // Log error
		return nil
	}
	createdComponents["bodytube"] = bodytube

	// Nosecone
	nosecone := components.NewNoseconeFromORK(ecs.NewBasic(), orkData)
	if nosecone == nil {
		fmt.Println("Error creating Nosecone from ORK.")
		return nil
	}
	createdComponents["nosecone"] = nosecone

	// Finset
	finset := components.NewTrapezoidFinsetFromORK(ecs.NewBasic(), orkData)
	if finset == nil {
		fmt.Println("Error creating Finset from ORK.")
		return nil
	}
	createdComponents["finset"] = finset

	// Parachute
	parachute, err := components.NewParachuteFromORK(ecs.NewBasic(), orkData)
	if err != nil {
		fmt.Printf("Error creating Parachute from ORK: %v\n", err)
		// Decide if this is fatal - maybe return nil or proceed without parachute?
		// For now, let's treat it as fatal for safety.
		return nil
	}
	createdComponents["parachute"] = parachute

	// --- 2. Calculate Total Mass from Created Components ---
	initialMass := calculateTotalMassFromComponents(createdComponents)

	// Validate mass before creating entity
	if initialMass <= 0 {
		fmt.Printf("Error: Cannot create RocketEntity, calculated initial mass from components is invalid (%.4f).\n", initialMass)
		return nil
	}

	// --- 3. Create Rocket Entity ---
	basic := ecs.NewBasic()
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
				Quat:        *types.IdentityQuaternion(),
			},
			AngularAcceleration: &types.Vector3{},
			AngularVelocity:     &types.Vector3{},
		},
		Mass: &types.Mass{
			BasicEntity: basic,
			Value:       initialMass, // Set mass using calculated value
		},
		components: createdComponents, // Assign the map of created components
	}

	// Assign components to PhysicsState *after* creating PhysicsState
	rocket.PhysicsState.Motor = motor // Assign directly for physics system access
	rocket.PhysicsState.Bodytube = bodytube
	rocket.PhysicsState.Nosecone = nosecone
	// ... assign other relevant components like finset if needed by physics/aero ...

	return rocket
}

// --- Mass Calculation Helpers ---

// Interface to get mass from a component
type massProvider interface {
	GetMass() float64
}

// calculateTotalMassFromComponents sums masses from a map of created components.
func calculateTotalMassFromComponents(components map[string]interface{}) float64 {
	var totalMass float64
	for name, comp := range components {
		if provider, ok := comp.(massProvider); ok {
			mass := provider.GetMass()
			if math.IsNaN(mass) || mass < 0 {
				fmt.Printf("Warning: Invalid mass (%.4f) from component '%s' (%T), skipping.\n", mass, name, comp)
				continue // Skip this component's mass
			}
			totalMass += mass
		} else {
			// This component doesn't provide mass via GetMass()
			// fmt.Printf("Info: Component '%s' (%T) does not provide mass via GetMass().\n", name, comp) 
		}
	}

	if math.IsNaN(totalMass) || totalMass <= 0 {
		fmt.Printf("Warning: Final calculated total mass from components is invalid or zero (%.4f). Returning 0.\n", totalMass)
		return 0.0
	}
	return totalMass
}

// Removed deprecated functions: calculateTotalMass, sumStandardComponentMasses, addComponentMass, getValidMotorMass

// AddComponent adds a component to the entity
func (r *RocketEntity) GetComponent(name string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.components[name]
}
