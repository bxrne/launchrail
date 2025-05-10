package entities

import (
	"fmt"
	"math"
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	openrocket "github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
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
// NewRocketEntity creates a new rocket entity from OpenRocket data
func NewRocketEntity(world *ecs.World, orkData *openrocket.RocketDocument, motor *components.Motor, log *logf.Logger) *RocketEntity {
	if orkData == nil || motor == nil {
		return nil
	}

	// --- 1. Create Components First ---
	createdComponents := make(map[string]interface{}) // Use interface{} for flexibility

	// Motor (already created, just validate and add)
	motorName := "unknown"
	if motor.Props != nil && motor.Props.Designation != "" {
		motorName = string(motor.Props.Designation) // Access via Props
	}
	if motor.GetMass() <= 0 { // Check mass using GetMass()
		log.Error(fmt.Sprintf("Cannot create RocketEntity, motor '%s' has invalid initial mass.", motorName)) // Use fmt.Sprintf
		return nil
	}
	createdComponents["motor"] = motor // Add validated motor

	// Bodytube
	bodytube, err := components.NewBodytubeFromORK(ecs.NewBasic(), orkData)
	if err != nil {
		log.Warn(fmt.Sprintf("Error creating Bodytube from ORK: %v", err)) // Use fmt.Sprintf
		return nil
	}
	createdComponents["bodytube"] = bodytube
	log.Info("Created BodyTube component", "id", bodytube.ID.ID(), "mass", bodytube.GetMass())

	// Create Finset component if present in BodyTube subcomponents
	var finset *components.TrapezoidFinset // Correct type
	// Access stages via Subcomponents and check FinCount > 0
	if len(orkData.Subcomponents.Stages) > 0 && 
		len(orkData.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets) > 0 &&
		orkData.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets[0].FinCount > 0 {
		
		orkSpecificFinset := orkData.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets[0]
		// Assuming position from ORK is the X-offset for the finset attachment point (e.g., LE of root chord)
		// Y and Z are assumed to be 0 in this local frame of attachment for now.
		finsetPosition := types.Vector3{X: orkSpecificFinset.Position.Value, Y: 0, Z: 0}
		finsetMaterial := orkSpecificFinset.Material

		createdFinset, err := components.NewTrapezoidFinsetFromORK(&orkSpecificFinset, finsetPosition, finsetMaterial)

		// Check if creation was successful (constructor might return nil on error)
		if err != nil {
			log.Error("Failed to create Finset component from ORK data", "finset_name", orkSpecificFinset.Name, "error", err)
		} else if createdFinset == nil {
			log.Error("Failed to create Finset component from ORK data (nil returned without error)", "finset_name", orkSpecificFinset.Name)
		} else {
			finset = createdFinset
			createdComponents["finset"] = finset
			log.Info("Created Finset component", "id", finset.ID(), "mass", finset.GetMass()) // Use ID()
		}
	}

	// Nosecone
	nosecone := components.NewNoseconeFromORK(ecs.NewBasic(), orkData)
	if nosecone == nil {
		return nil
	}
	createdComponents["nosecone"] = nosecone

	// Parachute
	parachute, err := components.NewParachuteFromORK(ecs.NewBasic(), orkData)
	if err != nil {
		log.Warn(fmt.Sprintf("Error creating Parachute from ORK: %v", err)) // Use fmt.Sprintf
		return nil
	}
	createdComponents["parachute"] = parachute

	// --- 2. Calculate Total Mass from Created Components ---
	initialMass := calculateTotalMassFromComponents(createdComponents, log)

	// Validate mass before creating entity
	if initialMass <= 0 {
		log.Error(fmt.Sprintf("Cannot create RocketEntity, calculated initial mass from components is invalid (%.4f).", initialMass)) // Use fmt.Sprintf
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
	rocket.PhysicsState.Finset = finset // Assign finset if created
	// ... assign other relevant components like finset if needed by physics/aero ...

	return rocket
}

// --- Mass Calculation Helpers ---

// Interface to get mass from a component
type massProvider interface {
	GetMass() float64
}

// calculateTotalMassFromComponents sums masses from a map of created components.
// calculateTotalMassFromComponents sums masses from a map of created components.
func calculateTotalMassFromComponents(components map[string]interface{}, log *logf.Logger) float64 {
	var totalMass float64
	log.Info("Calculating total mass from components...") // Added
	for name, comp := range components {
		if provider, ok := comp.(massProvider); ok {
			mass := provider.GetMass()
			// Added detailed log for each component's mass
			log.Info("Component Mass Contribution", "name", name, "type", fmt.Sprintf("%T", comp), "mass_kg", mass)
			if math.IsNaN(mass) || mass < 0 {
				log.Warn(fmt.Sprintf("Invalid mass (%.4f) from component, skipping.", mass), "component_name", name, "component_type", fmt.Sprintf("%T", comp))
				continue // Skip negative mass components
			}
			totalMass += mass
		} else { // Added
			log.Warn("Component does not implement massProvider", "name", name, "type", fmt.Sprintf("%T", comp)) // Added
		}
	}

	log.Info("Final calculated total mass from components", "total_mass_kg", totalMass) // Added
	if totalMass <= 0 {
		log.Warn(fmt.Sprintf("Final calculated total mass from components is invalid or zero (%.4f). Returning 0.", totalMass))
		return 0.0
	}

	return totalMass
}

// AddComponent adds a component to the entity
func (r *RocketEntity) GetComponent(name string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.components[name]
}
