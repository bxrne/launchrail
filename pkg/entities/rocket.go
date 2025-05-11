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

// initComponentsFromORK creates and initializes rocket components from OpenRocket data.
func initComponentsFromORK(orkData *openrocket.RocketDocument, motor *components.Motor, log *logf.Logger) (map[string]interface{}, *components.Bodytube, *components.Nosecone, *components.TrapezoidFinset, *components.Parachute, error) {
	createdComponents := make(map[string]interface{}) // Use interface{} for flexibility

	// Motor (already created, just validate and add)
	motorName := "unknown"
	if motor.Props != nil && motor.Props.Designation != "" {
		motorName = string(motor.Props.Designation) // Access via Props
	}
	if motor.GetMass() <= 0 { // Check mass using GetMass()
		return nil, nil, nil, nil, nil, fmt.Errorf("motor '%s' has invalid initial mass (%.4f)", motorName, motor.GetMass())
	}
	createdComponents["motor"] = motor // Add validated motor

	// Populate motor dimensions from OpenRocket data
	if len(orkData.Subcomponents.Stages) > 0 {
		orkDefinitionMotor := orkData.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.InnerTube.MotorMount.Motor
		if orkDefinitionMotor.Designation != "" {
			if m, ok := createdComponents["motor"].(*components.Motor); ok {
				m.Length = orkDefinitionMotor.Length
				m.Diameter = orkDefinitionMotor.Diameter
				log.Info("Updated motor dimensions from ORK data", "name", motorName, "length", m.Length, "diameter", m.Diameter)
			} else {
				log.Warn("Motor component in createdComponents is not of type *components.Motor, cannot update dimensions")
			}
		} else {
			log.Warn("Motor definition in OpenRocket data seems incomplete (empty designation); dimensions not updated.")
		}
	} else {
		log.Warn("No stages found in OpenRocket data; cannot update motor dimensions.")
	}

	// Bodytube
	bodytube, err := components.NewBodytubeFromORK(ecs.NewBasic(), orkData)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("error creating Bodytube from ORK: %w", err)
	}
	createdComponents["bodytube"] = bodytube
	log.Info("Created BodyTube component", "id", bodytube.ID.ID(), "mass", bodytube.GetMass())

	// Finset (optional)
	var finset *components.TrapezoidFinset
	if len(orkData.Subcomponents.Stages) > 0 &&
		len(orkData.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets) > 0 &&
		orkData.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets[0].FinCount > 0 {

		orkSpecificFinset := orkData.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets[0]
		finsetPosition := types.Vector3{X: orkSpecificFinset.Position.Value, Y: 0, Z: 0}
		finsetMaterial := orkSpecificFinset.Material

		createdFinset, finsetErr := components.NewTrapezoidFinsetFromORK(&orkSpecificFinset, finsetPosition, finsetMaterial)
		if finsetErr != nil {
			log.Error("Failed to create Finset component from ORK data", "finset_name", orkSpecificFinset.Name, "error", finsetErr)
			// Not returning an error, as finset might be optional for some configurations
		} else if createdFinset == nil {
			log.Error("Failed to create Finset component (nil returned without error)", "finset_name", orkSpecificFinset.Name)
		} else {
			finset = createdFinset
			createdComponents["finset"] = finset
			log.Info("Created Finset component", "id", finset.ID(), "mass", finset.GetMass()) // Use ID()
		}
	}

	// Nosecone
	nosecone := components.NewNoseconeFromORK(ecs.NewBasic(), orkData)
	if nosecone == nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("error creating Nosecone from ORK (nil returned)")
	}
	createdComponents["nosecone"] = nosecone

	// Parachute
	parachute, err := components.NewParachuteFromORK(ecs.NewBasic(), orkData)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("error creating Parachute from ORK: %w", err)
	}
	createdComponents["parachute"] = parachute

	return createdComponents, bodytube, nosecone, finset, parachute, nil
}

// calculateCGAndInertia calculates the overall rocket center of mass (CG) and aggregate inertia tensor.
func calculateCGAndInertia(initialMass float64, createdComponents map[string]interface{}, rLogger *logf.Logger) (types.Vector3, types.Matrix3x3, types.Matrix3x3) {
	var overallRocketCG types.Vector3
	var totalInertiaTensorBody types.Matrix3x3
	var inverseTotalInertiaTensorBody types.Matrix3x3

	if initialMass > 1e-9 { // Avoid division by zero or tiny mass issues
		// First pass: Calculate overall rocket Center of Mass (CG)
		var sumMass float64            // To verify against initialMass
		sumWeightedPosition := types.Vector3{X: 0, Y: 0, Z: 0}

		for _, compGeneric := range createdComponents {
			var massC float64
			var posCRefGlobal types.Vector3 // Position of component's reference point in rocket global axes
			var posCCmLocal types.Vector3   // Position of component's CM relative to its reference point, in component local axes (aligned with rocket body axes)

			switch comp := compGeneric.(type) {
			case *components.Motor:
				massC = comp.GetMass()
				posCRefGlobal = comp.GetPosition()
				posCCmLocal = comp.GetCenterOfMassLocal()
			case *components.Bodytube:
				massC = comp.GetMass()
				posCRefGlobal = comp.Position
				posCCmLocal = comp.GetCenterOfMass()
			case *components.Nosecone:
				massC = comp.GetMass()
				posCRefGlobal = comp.GetPosition()
				posCCmLocal = comp.GetCenterOfMassLocal()
			case *components.TrapezoidFinset:
				massC = comp.GetMass()
				posCRefGlobal = comp.Position
				posCCmLocal = comp.GetCenterOfMassLocal()
			// NOTE: Parachute components are currently skipped in CG and Inertia calculations
			// as they don't typically have GetPosition, GetCenterOfMassLocal, or GetInertiaTensorLocal methods.
			// If these are added in the future, the case for *components.Parachute can be re-added here.
			default:
				rLogger.Warn("Unknown or unhandled component type in CG calculation", "type", fmt.Sprintf("%T", compGeneric))
				continue
			}

			if massC > 1e-9 {
				posCCmGlobal := posCRefGlobal.Add(posCCmLocal) // CM of component_c in rocket global coordinates
				sumMass += massC
				sumWeightedPosition = sumWeightedPosition.Add(posCCmGlobal.MultiplyScalar(massC))
			}
		}

		if sumMass > 1e-9 {
			overallRocketCG = sumWeightedPosition.DivideScalar(sumMass)
		} else {
			rLogger.Warn("Sum of component masses for CG calculation is zero or negative. CG will be at origin.")
			overallRocketCG = types.Vector3{X: 0, Y: 0, Z: 0}
		}

		// Second pass: Calculate aggregate inertia tensor about overallRocketCG (Parallel Axis Theorem)
		for _, compGeneric := range createdComponents {
			var massC float64
			var posCRefGlobal types.Vector3
			var posCCmLocal types.Vector3
			var icLocalCm types.Matrix3x3 // Inertia tensor of component about its own CM, in rocket body axes

			switch comp := compGeneric.(type) {
			case *components.Motor:
				massC = comp.GetMass()
				posCRefGlobal = comp.GetPosition()
				posCCmLocal = comp.GetCenterOfMassLocal()
				icLocalCm = comp.GetInertiaTensorLocal()
			case *components.Bodytube:
				massC = comp.GetMass()
				posCRefGlobal = comp.Position
				posCCmLocal = comp.GetCenterOfMass()
				icLocalCm = comp.GetInertiaTensor()
			case *components.Nosecone:
				massC = comp.GetMass()
				posCRefGlobal = comp.GetPosition()
				posCCmLocal = comp.GetCenterOfMassLocal()
				icLocalCm = comp.GetInertiaTensorLocal()
			case *components.TrapezoidFinset:
				massC = comp.GetMass()
				posCRefGlobal = comp.Position
				posCCmLocal = comp.GetCenterOfMassLocal()
				icLocalCm = comp.GetInertiaTensorLocal()
			default:
				rLogger.Warn("Unknown or unhandled component type in inertia calculation", "type", fmt.Sprintf("%T", compGeneric))
				continue
			}

			if massC > 1e-9 {
				posCCmGlobal := posCRefGlobal.Add(posCCmLocal)
				d := posCCmGlobal.Subtract(overallRocketCG) // Vector from overallRocketCG to component's CM.

				// Parallel Axis Theorem: I_rocket_cg = I_comp_cm + m * ( (d.d)I - d (outer_product) d )
				dotD := d.Dot(d)
				term1 := types.IdentityMatrix3x3().MultiplyScalar(dotD * massC)
				dOuterProduct := types.Matrix3x3{
					M11: d.X * d.X, M12: d.X * d.Y, M13: d.X * d.Z,
					M21: d.Y * d.X, M22: d.Y * d.Y, M23: d.Y * d.Z,
					M31: d.Z * d.X, M32: d.Z * d.Y, M33: d.Z * d.Z,
				}
				term2 := dOuterProduct.MultiplyScalar(massC)
				parallelAxisCorrection := term1.Subtract(term2)
				IcShifted := icLocalCm.Add(parallelAxisCorrection)
				totalInertiaTensorBody = totalInertiaTensorBody.Add(IcShifted)
			}
		}

		inversePtr := totalInertiaTensorBody.Inverse()
		if inversePtr != nil {
			inverseTotalInertiaTensorBody = *inversePtr
		} else {
			rLogger.Error("Calculated total rocket body inertia tensor is singular! Using identity matrix as fallback for its inverse.")
			inverseTotalInertiaTensorBody = types.IdentityMatrix3x3() // Fallback to identity matrix
		}
	} else {
		rLogger.Warn("Rocket initial mass is zero or negative, skipping CG and inertia tensor calculation. CG and Inertia will be zero.")
		// overallRocketCG, totalInertiaTensorBody, and inverseTotalInertiaTensorBody remain zero
	}
	return overallRocketCG, totalInertiaTensorBody, inverseTotalInertiaTensorBody
}

// createPhysicsState creates and initializes the PhysicsState for the rocket.
func createPhysicsState(basic ecs.BasicEntity, overallRocketCG types.Vector3, totalInertiaTensorBody, inverseTotalInertiaTensorBody types.Matrix3x3, initialMass float64, motor *components.Motor, bodytube *components.Bodytube, nosecone *components.Nosecone, finset *components.TrapezoidFinset, parachute *components.Parachute) *states.PhysicsState {
	ps := &states.PhysicsState{
		Position: &types.Position{
			BasicEntity: basic,
			Vec:         overallRocketCG, // Initial position of the rocket CG
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
		InertiaTensorBody:        totalInertiaTensorBody,
		InverseInertiaTensorBody: inverseTotalInertiaTensorBody,
		// Assign components directly for physics system access
		Motor:    motor,
		Bodytube: bodytube,
		Nosecone: nosecone,
		Finset:   finset,
		Parachute: parachute, // Parachute is now used here
	}
	return ps
}

// NewRocketEntity creates a new rocket entity from OpenRocket data
func NewRocketEntity(world *ecs.World, orkData *openrocket.RocketDocument, motor *components.Motor, log *logf.Logger) *RocketEntity {
	if orkData == nil || motor == nil {
		log.Error("Cannot create RocketEntity: orkData or motor is nil")
		return nil
	}

	// --- 1. Create Components First ---
	createdComponents, bodytube, nosecone, finset, parachute, err := initComponentsFromORK(orkData, motor, log)
	if err != nil {
		log.Error("Failed to initialize components for RocketEntity", "error", err)
		return nil
	}

	// --- 2. Calculate Total Mass from Created Components ---
	initialMass := calculateInitialRocketMass(createdComponents, log)

	// Validate mass before creating entity
	if initialMass <= 0 {
		log.Error(fmt.Sprintf("Cannot create RocketEntity, calculated initial mass from components is invalid (%.4f).", initialMass))
		return nil
	}

	// --- 2a. Calculate Overall Rocket Center of Mass (CG) & Aggregate Inertia Tensor ---
	overallRocketCG, totalInertiaTensorBody, inverseTotalInertiaTensorBody := calculateCGAndInertia(initialMass, createdComponents, log)

	// --- 3. Create Rocket Entity ---
	basic := ecs.NewBasic()
	rocket := &RocketEntity{
		BasicEntity: &basic,
		PhysicsState: createPhysicsState(basic, overallRocketCG, totalInertiaTensorBody, inverseTotalInertiaTensorBody, initialMass, motor, bodytube, nosecone, finset, parachute),
		Mass: &types.Mass{
			BasicEntity: basic,
			Value:       initialMass, // Set mass using calculated value
		},
		components: createdComponents, // Assign the map of created components
	}

	return rocket
}

// --- Mass Calculation Helpers ---

// Interface to get mass from a component
type massProvider interface {
	GetMass() float64
}

// calculateInitialRocketMass sums masses from a map of created components.
func calculateInitialRocketMass(components map[string]interface{}, log *logf.Logger) float64 {
	var totalMass float64
	log.Info("Calculating total mass from components...") // Added
	for name, comp := range components {
		if provider, ok := comp.(massProvider); ok {
			mass := provider.GetMass()
			// Added detailed log for each component's mass
			log.Info("Component Mass Contribution", "name", name, "type", fmt.Sprintf("%T", comp), "mass_kg", mass)
			if math.IsNaN(mass) || mass < 0 {
				log.Warn("Invalid mass (%.4f) from component, skipping.", "component_name", name, "component_type", fmt.Sprintf("%T", comp), "mass", mass)
				continue // Skip negative mass components
			}
			totalMass += mass
		} else { // Added
			log.Warn("Component does not implement massProvider", "name", name, "type", fmt.Sprintf("%T", comp)) // Added
		}
	}

	log.Info("Final calculated total mass from components", "total_mass_kg", totalMass) // Added
	if totalMass <= 0 {
		log.Warn("Final calculated total mass from components is invalid or zero (%.4f). Returning 0.", "total_mass", totalMass)
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
