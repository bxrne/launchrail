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

	// --- 2a. Calculate Overall Rocket Center of Mass (CG) & Aggregate Inertia Tensor ---
	var overallRocketCG types.Vector3
	totalInertiaTensorBody := types.Matrix3x3{} // Initialize as zero matrix
	var inverseTotalInertiaTensorBody types.Matrix3x3

	if initialMass > 1e-9 { // Proceed only if there's mass
		var sumWeightedPosition types.Vector3

		// First pass: Calculate overall CG of the rocket
		// This requires each component to provide its mass and the global position of its CG.
		// Global position of component's CG = component.Position (its ref point) + component.GetCenterOfMassLocal()
		for _, compGeneric := range createdComponents {
			var mass_c float64
			var pos_c_ref_global types.Vector3   // Component's reference point in rocket global coordinates
			var pos_c_cm_local types.Vector3     // Component's CM relative to its reference point

			switch comp := compGeneric.(type) {
			case *components.Motor:
				mass_c = comp.GetMass()
				pos_c_ref_global = comp.GetPosition() // Needs GetPosition() method
				pos_c_cm_local = comp.GetCenterOfMassLocal() // Needs GetCenterOfMassLocal()
			case *components.Bodytube:
				mass_c = comp.GetMass()
				pos_c_ref_global = comp.Position // Directly access if public, or add GetPosition()
				pos_c_cm_local = comp.GetCenterOfMass() // Existing method
			case *components.Nosecone:
				mass_c = comp.GetMass()
				pos_c_ref_global = comp.GetPosition() // Needs GetPosition()
				pos_c_cm_local = comp.GetCenterOfMassLocal() // Needs GetCenterOfMassLocal()
			case *components.TrapezoidFinset:
				mass_c = comp.GetMass()
				pos_c_ref_global = comp.Position // Directly access if public, or add GetPosition()
				// TrapezoidFinset.CenterOfMass is already global for the finset itself.
				// For this calculation, we need its CM relative to its own attachment point `comp.Position`.
				// This requires re-evaluation of how TrapezoidFinset.CenterOfMass is defined or a new method.
				// For now, assume a GetCenterOfMassLocal() similar to others.
				pos_c_cm_local = comp.GetCenterOfMassLocal() // Needs this method returning local CM
			// Skipping Parachute for inertia tensor calculation for now
			default:
				continue // Skip non-physical components or unhandled types
			}

			if mass_c > 1e-9 {
				pos_c_cm_global := pos_c_ref_global.Add(pos_c_cm_local) // Vector3.Add() needed
				sumWeightedPosition = sumWeightedPosition.Add(pos_c_cm_global.MultiplyScalar(mass_c)) // Vector3.MultiplyScalar() needed
			}
		}
		overallRocketCG = sumWeightedPosition.DivideScalar(initialMass) // Vector3.DivideScalar() needed
		log.Info("Calculated Overall Rocket CG", "cgX", overallRocketCG.X, "cgY", overallRocketCG.Y, "cgZ", overallRocketCG.Z)

		// Second pass: Aggregate inertia tensors
		for _, compGeneric := range createdComponents {
			var mass_c float64
			var pos_c_ref_global types.Vector3
			var pos_c_cm_local types.Vector3
			var I_c_local_cm types.Matrix3x3 // Inertia tensor of component about its own CM, in rocket body axes

			switch comp := compGeneric.(type) {
			case *components.Motor:
				mass_c = comp.GetMass()
				pos_c_ref_global = comp.GetPosition()
				pos_c_cm_local = comp.GetCenterOfMassLocal()
				I_c_local_cm = comp.GetInertiaTensorLocal() // Needs GetInertiaTensorLocal()
			case *components.Bodytube:
				mass_c = comp.GetMass()
				pos_c_ref_global = comp.Position
				pos_c_cm_local = comp.GetCenterOfMass()
				I_c_local_cm = comp.GetInertiaTensor()
			case *components.Nosecone:
				mass_c = comp.GetMass()
				pos_c_ref_global = comp.GetPosition()
				pos_c_cm_local = comp.GetCenterOfMassLocal()
				I_c_local_cm = comp.GetInertiaTensorLocal() // Needs GetInertiaTensorLocal()
			case *components.TrapezoidFinset:
				mass_c = comp.GetMass()
				pos_c_ref_global = comp.Position
				pos_c_cm_local = comp.GetCenterOfMassLocal() // Needs local CM
				I_c_local_cm = comp.GetInertiaTensorLocal()      // Changed to GetInertiaTensorLocal for consistency
			default:
				continue
			}

			if mass_c > 1e-9 {
				pos_c_cm_global := pos_c_ref_global.Add(pos_c_cm_local)
				d := pos_c_cm_global.Subtract(overallRocketCG) // Vector from overallRocketCG to component's CM. Vector3.Subtract() needed

				// Parallel Axis Theorem: I_rocket_cg = I_comp_cm + m * ( (d.d)I - d (outer_product) d )
				dot_d := d.Dot(d) // Vector3.Dot() needed
				// Term m * (d.d)I
				term1 := types.IdentityMatrix3x3().MultiplyScalar(dot_d * mass_c) // types.IdentityMatrix3x3() and Matrix3x3.MultiplyScalar() needed

				// Term m * (d (outer_product) d)
				// (d outer_product d) = [dx*dx, dx*dy, dx*dz]
				//                       [dy*dx, dy*dy, dy*dz]
				//                       [dz*dx, dz*dy, dz*dz]
				dOuterProduct := types.Matrix3x3{
					M11: d.X * d.X, M12: d.X * d.Y, M13: d.X * d.Z,
					M21: d.Y * d.X, M22: d.Y * d.Y, M23: d.Y * d.Z,
					M31: d.Z * d.X, M32: d.Z * d.Y, M33: d.Z * d.Z,
				}
				term2 := dOuterProduct.MultiplyScalar(mass_c)

				parallelAxisCorrection := term1.Subtract(term2) // Matrix3x3.Subtract() needed
				I_c_shifted := I_c_local_cm.Add(parallelAxisCorrection) // Matrix3x3.Add() needed

				totalInertiaTensorBody = totalInertiaTensorBody.Add(I_c_shifted)
			}
		}

		inversePtr := totalInertiaTensorBody.Inverse() // Matrix3x3.Inverse() needed, assuming it returns *Matrix3x3 or nil
		if inversePtr != nil {
			inverseTotalInertiaTensorBody = *inversePtr
		} else {
			log.Error("Calculated total rocket body inertia tensor is singular!")
			// Keep inverseTotalInertiaTensorBody as zero matrix if singular
		}
	} else {
		log.Warn("Rocket initial mass is zero, skipping CG and inertia tensor calculation.")
		// totalInertiaTensorBody and inverseTotalInertiaTensorBody remain zero matrices
	}


	// --- 3. Create Rocket Entity ---
	basic := ecs.NewBasic()
	rocket := &RocketEntity{
		BasicEntity: &basic,
		PhysicsState: &states.PhysicsState{
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
			// Store calculated inertia tensors
			InertiaTensorBody:        totalInertiaTensorBody,
			InverseInertiaTensorBody: inverseTotalInertiaTensorBody,
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
