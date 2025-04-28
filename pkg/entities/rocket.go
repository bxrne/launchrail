package entities

import (
	"fmt"
	"math"
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
		fmt.Println("Error: Cannot create RocketEntity, OpenRocket data or motor is nil.")
		return nil
	}

	// Validate motor first (needs mass)
	// GetMass() returns the *current* mass. For initial mass, this is fine.
	if motor.GetMass() <= 0 {
		// Use Designation for motor name if Props is not nil and Designation is not empty
		motorName := "unknown"
		if motor.Props != nil && motor.Props.Designation != "" { // Check if empty string
			motorName = string(motor.Props.Designation) // It's already a string (or string alias)
		}
		fmt.Printf("Error: Cannot create RocketEntity, motor '%s' has invalid initial mass.\n", motorName)
		return nil
	}

	initialMass := calculateTotalMass(orkData, motor) // Pass motor here

	// Validate mass before creating entity
	if initialMass <= 0 {
		fmt.Printf("Error: Cannot create RocketEntity, calculated initial mass is invalid (%.4f).\n", initialMass)
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
			Value:       initialMass, // Set mass first
		},
		components: make(map[string]interface{}),
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

// calculateTotalMass sums masses of all components from OpenRocket data.
// It attempts to calculate material mass where possible and adds explicit MassComponent masses.
// Motor mass is handled separately via the components.Motor struct passed to NewRocketEntity.
func calculateTotalMass(orkData *openrocket.RocketDocument, motor *components.Motor) float64 {
	// Access the Rocket document within the OpenrocketDocument
	// The schema confirmed orkData.Rocket exists.
	// Reverting explicit dereference.

	if orkData == nil || len(orkData.Subcomponents.Stages) == 0 {
		fmt.Println("Warning: Cannot calculate mass, OpenRocket data or stages missing.")
		return 0.0
	}

	if motor == nil {
		fmt.Println("Warning: Cannot calculate total mass, motor component is nil.")
		return 0.0
	}

	var totalMass float64
	// Assuming single stage - previously validated
	stage := orkData.Subcomponents.Stages[0]       // Standard access
	sustainerSubs := &stage.SustainerSubcomponents // Pass pointer to avoid copying large struct

	// --- Add mass for standard components --- (Extracted to helper)
	sumStandardComponentMasses(&totalMass, sustainerSubs)

	// --- Add Motor Mass --- (Extracted to helper)
	totalMass += getValidMotorMass(motor)

	// --- Final Validation ---
	if math.IsNaN(totalMass) || totalMass <= 0 {
		fmt.Printf("Warning: Final calculated total mass is invalid or zero (%.4f). Returning 0.\n", totalMass)
		return 0.0
	}
	// fmt.Printf("Final Calculated Total Mass: %.4f\n", totalMass) // Debug
	return totalMass
}

// massProvider defines the interface for components that can provide their mass.
// This allows calculateTotalMass to work with any component implementing GetMass().
type massProvider interface {
	GetMass() float64
}

// sumStandardComponentMasses iterates through standard components and adds their mass.
func sumStandardComponentMasses(
	totalMass *float64,
	sustainer *openrocket.SustainerSubcomponents, // Pass pointer to sustainer subcomponents
) {
	// Derive necessary sub-component structs from the sustainer
	noseSubs := &sustainer.Nosecone.Subcomponents
	bodyTubeSubs := &sustainer.BodyTube.Subcomponents

	// --- Add mass for components with GetMass() ---
	// Top-level components (accessed via sustainer)
	addComponentMass(totalMass, "Nosecone", &sustainer.Nosecone)
	addComponentMass(totalMass, "BodyTube", &sustainer.BodyTube) // Tube material mass
	// Components nested within BodyTube (accessed via bodyTubeSubs)
	addComponentMass(totalMass, "TrapezoidFinset", &bodyTubeSubs.TrapezoidFinset)
	addComponentMass(totalMass, "Parachute", &bodyTubeSubs.Parachute)
	addComponentMass(totalMass, "Shockcord", &bodyTubeSubs.Shockcord)
	addComponentMass(totalMass, "InnerTube", &bodyTubeSubs.InnerTube) // Inner tube material mass
	// CenteringRings (Iterate via bodyTubeSubs)
	for i := range bodyTubeSubs.CenteringRings {
		addComponentMass(totalMass, fmt.Sprintf("CenteringRing[%d]", i), &bodyTubeSubs.CenteringRings[i])
	}

	// --- Add mass for explicit MassComponents ---
	// (Using derived noseSubs)
	addComponentMass(totalMass, "Nosecone.MassComponent", &noseSubs.MassComponent)
}

// addComponentMass validates and adds the mass of a single component to the total mass.
func addComponentMass(totalMass *float64, compName string, comp massProvider) {
	if comp == nil { // Check if the component itself is nil
		// fmt.Printf("Info: Skipping mass for nil component '%s'\n", compName)
		return
	}
	mass := comp.GetMass()
	if math.IsNaN(mass) || mass < 0 {
		fmt.Printf("Warning: Invalid mass (%.4f) calculated for component '%s' (%T), skipping.\n", mass, compName, comp)
		return
	}
	// fmt.Printf("Adding mass for %s (%T): %.4f\n", compName, comp, mass) // Debug
	*totalMass += mass
}

// getValidMotorMass calculates and validates the motor's initial mass.
// It returns the valid mass or 0.0 if invalid, logging a warning.
func getValidMotorMass(motor *components.Motor) float64 {
	if motor == nil {
		fmt.Println("Warning: Motor component is nil in getValidMotorMass.")
		return 0.0
	}

	motorMass := motor.GetMass() // GetMass() returns the current mass (initial mass at t=0)

	if math.IsNaN(motorMass) || motorMass < 0 {
		motorName := "unknown"
		if motor.Props != nil && motor.Props.Designation != "" { // Check if empty string
			motorName = string(motor.Props.Designation)
		}
		fmt.Printf("Warning: Invalid initial mass (%.4f) obtained from motor component '%s', skipping motor mass.\n", motorMass, motorName)
		return 0.0 // Return 0 if mass is invalid
	}

	// fmt.Printf("Adding mass for Motor '%s': %.4f\n", string(motor.Props.Designation), motorMass) // Debug
	return motorMass // Return the valid mass
}

// AddComponent adds a component to the entity
func (r *RocketEntity) GetComponent(name string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.components[name]
}
