package entities

import (
	"fmt"
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	openrocket "github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

// InertialComponent defines an interface for components that contribute to mass and inertia.
type InertialComponent interface {
	GetMass() float64
	GetPosition() types.Vector3             // Position of the component's reference point in rocket global axes
	GetCenterOfMassLocal() types.Vector3    // Position of the component's CM relative to its reference point, in component local axes
	GetInertiaTensorLocal() types.Matrix3x3 // Inertia tensor of component about its own CM, in component local axes
}

// RocketEntity represents a complete rocket with all its components
type RocketEntity struct {
	*ecs.BasicEntity
	*states.PhysicsState
	*types.Mass // This will store the initial total mass for reference
	components map[string]interface{}
	mu         sync.RWMutex
}

// GetCurrentMassKg calculates the current total mass of the rocket by summing
// the masses of its components. It pays special attention to components
// that implement InertialComponent (like the motor) to get their current mass.
func (r *RocketEntity) GetCurrentMassKg() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	currentMass := 0.0
	for _, compGeneric := range r.components {
		if inertialComp, ok := compGeneric.(InertialComponent); ok {
			currentMass += inertialComp.GetMass() // Gets dynamic mass from motor, fixed from others if they implement it
		} else if compWithMass, ok := compGeneric.(interface{ GetMass() float64 }); ok {
			// Fallback for components that have GetMass but aren't (yet) full InertialComponents
			// This path might be hit by Bodytube, Nosecone etc., if not yet updated to InertialComponent fully.
			// It assumes their GetMass() returns their current (fixed) mass.
			currentMass += compWithMass.GetMass()
		} 
	}
	if currentMass <= 1e-9 {
		// This case should ideally not happen if initialized correctly
		// Fallback to initial mass if something went wrong with component sum
		if r.Mass != nil {
			return r.Mass.Value
		}
		return 0 // Or handle error
	}
	return currentMass
}

// Helper function to initialize and update motor from ORK data
func initializeMotorFromORK(orkData *openrocket.OpenrocketDocument, motor *components.Motor, log *logf.Logger) error {
	motorName := "unknown"
	if motor.Props != nil && motor.Props.Designation != "" {
		motorName = string(motor.Props.Designation)
	}
	if motor.GetMass() <= 0 {
		return fmt.Errorf("motor '%s' has invalid initial mass (%.4f)", motorName, motor.GetMass())
	}

	if len(orkData.Rocket.Subcomponents.Stages) > 0 {
		orkDefinitionMotor := orkData.Rocket.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.InnerTube.MotorMount.Motor
		if orkDefinitionMotor.Designation != "" {
			motor.Length = orkDefinitionMotor.Length
			motor.Diameter = orkDefinitionMotor.Diameter
			log.Info("Updated motor dimensions from ORK data", "name", motorName, "length", motor.Length, "diameter", motor.Diameter)
		} else {
			log.Warn("Motor definition in OpenRocket data seems incomplete (empty designation); dimensions not updated.")
		}
	} else {
		log.Warn("No stages found in OpenRocket data; cannot update motor dimensions.")
	}
	return nil
}

// Helper function to create finset from ORK data if present
func createFinsetFromORK(orkData *openrocket.OpenrocketDocument, log *logf.Logger) (*components.TrapezoidFinset, error) {
	if len(orkData.Rocket.Subcomponents.Stages) == 0 ||
		len(orkData.Rocket.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets) == 0 ||
		orkData.Rocket.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets[0].FinCount <= 0 {
		log.Info("No finset data found or fin count is zero, skipping finset creation.")
		return nil, nil // No finset data or fin count is zero, not an error
	}

	orkSpecificFinset := orkData.Rocket.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets[0]
	finsetPosition := types.Vector3{X: orkSpecificFinset.Position.Value, Y: 0, Z: 0}
	finsetMaterial := orkSpecificFinset.Material

	createdFinset, finsetErr := components.NewTrapezoidFinsetFromORK(&orkSpecificFinset, finsetPosition, finsetMaterial)
	if finsetErr != nil {
		log.Error("Failed to create Finset component from ORK data", "finset_name", orkSpecificFinset.Name, "error", finsetErr)
		return nil, fmt.Errorf("failed to create Finset: %w", finsetErr) // Return error to caller
	} else if createdFinset == nil {
		log.Error("Failed to create Finset component (nil returned without error)", "finset_name", orkSpecificFinset.Name)
		return nil, fmt.Errorf("NewTrapezoidFinsetFromORK returned nil for %s without explicit error", orkSpecificFinset.Name)
	}

	log.Info("Created Finset component", "id", createdFinset.ID(), "mass", createdFinset.GetMass())
	return createdFinset, nil
}

// initComponentsFromORK creates and initializes rocket components from OpenRocket data.
func initComponentsFromORK(orkData *openrocket.OpenrocketDocument, motor *components.Motor, log *logf.Logger) (map[string]interface{}, *components.Bodytube, *components.Nosecone, *components.TrapezoidFinset, *components.Parachute, error) {
	createdComponents := make(map[string]interface{}) // Use interface{} for flexibility

	// Motor
	if err := initializeMotorFromORK(orkData, motor, log); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	createdComponents["motor"] = motor

	// Bodytube
	bodytube, err := components.NewBodytubeFromORK(ecs.NewBasic(), orkData)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("error creating Bodytube from ORK: %w", err)
	}
	createdComponents["bodytube"] = bodytube
	log.Info("Created BodyTube component", "id", bodytube.ID.ID(), "mass", bodytube.GetMass())

	// Finset (optional)
	finset, err := createFinsetFromORK(orkData, log)
	if err != nil {
		// Log the error but don't necessarily fail the whole rocket creation if finset is considered optional
		log.Warn("Error creating finset, proceeding without it", "error", err)
		// If finset is critical, this should return the error:
		// return nil, nil, nil, nil, nil, fmt.Errorf("error creating Finset from ORK: %w", err)
	} else if finset != nil {
		createdComponents["finset"] = finset
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

// Helper for calculateCGAndInertia to process a single component
func processInertialComponent(compGeneric interface{}, overallRocketCG types.Vector3, sumMass *float64, sumWeightedPosition *types.Vector3, totalInertiaTensorBody *types.Matrix3x3, isSecondPass bool, rLogger *logf.Logger) {
	inertialComp, ok := compGeneric.(InertialComponent)
	if !ok {
		// Parachute is not an InertialComponent, skip it silently or log if unexpected.
		if _, isParachute := compGeneric.(*components.Parachute); !isParachute {
			rLogger.Warn("Component does not implement InertialComponent and is not a Parachute", "type", fmt.Sprintf("%T", compGeneric))
		}
		return
	}

	massC := inertialComp.GetMass()
	if massC <= 1e-9 {
		return // Skip components with negligible or zero mass
	}

	posCRefGlobal := inertialComp.GetPosition()
	posCCmLocal := inertialComp.GetCenterOfMassLocal()
	posCCmGlobal := posCRefGlobal.Add(posCCmLocal) // CM of component_c in rocket global coordinates

	if !isSecondPass { // First pass: CG calculation
		*sumMass += massC
		*sumWeightedPosition = sumWeightedPosition.Add(posCCmGlobal.MultiplyScalar(massC))
	} else { // Second pass: Inertia calculation
		icLocalCm := inertialComp.GetInertiaTensorLocal()
		d := posCCmGlobal.Subtract(overallRocketCG) // Vector from overallRocketCG to component's CM.

		// Parallel Axis Theorem: I_rocket_cg = I_comp_cm + m * ( (d.d)I - d (outer_product) d )
		dotD := d.Dot(d)
		term1 := types.IdentityMatrix3x3().MultiplyScalar(dotD * massC)

		dOuterProductD := types.Matrix3x3{
			M11: d.X * d.X, M12: d.X * d.Y, M13: d.X * d.Z,
			M21: d.Y * d.X, M22: d.Y * d.Y, M23: d.Y * d.Z,
			M31: d.Z * d.X, M32: d.Z * d.Y, M33: d.Z * d.Z,
		}.MultiplyScalar(massC)

		// Contribution of this component to total inertia (Ic_rocket_cg = Ic_cm + PAT_term)
		// PAT_term = m * ( (d·d)I - d⊗d )
		patTerm := term1.Subtract(dOuterProductD) // This is the term from the parallel axis theorem
		inertiaContribution := icLocalCm.Add(patTerm)
		*totalInertiaTensorBody = totalInertiaTensorBody.Add(inertiaContribution)
	}
}

// PhysicsStateConfig holds parameters for creating a PhysicsState.
type PhysicsStateConfig struct {
	OverallRocketCG               types.Vector3
	TotalInertiaTensorBody        types.Matrix3x3
	InverseTotalInertiaTensorBody types.Matrix3x3
	InitialMass                   float64
	Motor                         *components.Motor
	Bodytube                      *components.Bodytube
	Nosecone                      *components.Nosecone
	Finset                        *components.TrapezoidFinset
	Parachute                     *components.Parachute
}

// createPhysicsState creates and initializes the PhysicsState for the rocket.
func createPhysicsState(basic ecs.BasicEntity, cfg *PhysicsStateConfig) *states.PhysicsState {
	ps := &states.PhysicsState{
		Position: &types.Position{
			BasicEntity: basic,
			Vec:         cfg.OverallRocketCG, // Initial position of the rocket CG
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
		AngularAcceleration:      &types.Vector3{},
		AngularVelocity:          &types.Vector3{},
		InertiaTensorBody:        cfg.TotalInertiaTensorBody,
		InverseInertiaTensorBody: cfg.InverseTotalInertiaTensorBody,
		// Assign components directly for physics system access
		Motor:     cfg.Motor,
		Bodytube:  cfg.Bodytube,
		Nosecone:  cfg.Nosecone,
		Finset:    cfg.Finset,
		Parachute: cfg.Parachute,
	}
	return ps
}

// calculateCGAndInertia calculates the overall rocket center of mass (CG) and aggregate inertia tensor.
func calculateCGAndInertia(initialMass float64, createdComponents map[string]interface{}, rLogger *logf.Logger) (types.Vector3, types.Matrix3x3, types.Matrix3x3) {
	var overallRocketCG types.Vector3
	var totalInertiaTensorBody types.Matrix3x3
	var inverseTotalInertiaTensorBody types.Matrix3x3

	if initialMass <= 1e-9 { // Avoid division by zero or tiny mass issues
		rLogger.Warn("Initial mass is zero or negative. CG and Inertia will be zero.")
		return overallRocketCG, totalInertiaTensorBody, inverseTotalInertiaTensorBody
	}

	// First pass: Calculate overall rocket Center of Mass (CG)
	var sumMass float64
	sumWeightedPosition := types.Vector3{X: 0, Y: 0, Z: 0}

	for _, compGeneric := range createdComponents {
		processInertialComponent(compGeneric, overallRocketCG, &sumMass, &sumWeightedPosition, &totalInertiaTensorBody, false, rLogger)
	}

	if sumMass > 1e-9 {
		overallRocketCG = sumWeightedPosition.DivideScalar(sumMass)
	} else {
		rLogger.Warn("Sum of component masses for CG calculation is zero or negative. CG will be at origin.")
		// overallRocketCG remains {0,0,0}
	}

	// Second pass: Calculate aggregate inertia tensor about overallRocketCG (Parallel Axis Theorem)
	for _, compGeneric := range createdComponents {
		processInertialComponent(compGeneric, overallRocketCG, &sumMass, &sumWeightedPosition, &totalInertiaTensorBody, true, rLogger)
	}

	inversePtr := totalInertiaTensorBody.Inverse()
	if inversePtr != nil {
		inverseTotalInertiaTensorBody = *inversePtr
	} else {
		rLogger.Error("Failed to invert total inertia tensor (matrix is singular). Inverse will be zero matrix.", "tensor", totalInertiaTensorBody)
		// inverseTotalInertiaTensorBody remains a zero matrix, which is a safe default.
	}

	return overallRocketCG, totalInertiaTensorBody, inverseTotalInertiaTensorBody
}

// NewRocketEntity creates a new rocket entity from OpenRocket data and a motor.
// It initializes physical properties, components, CG, and inertia.
func NewRocketEntity(world *ecs.World, orkData *openrocket.OpenrocketDocument, motor *components.Motor, log *logf.Logger) *RocketEntity {
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
	initialMass := 0.0
	for _, compGeneric := range createdComponents {
		if comp, ok := compGeneric.(interface{ GetMass() float64 }); ok {
			initialMass += comp.GetMass()
		} else {
			log.Warn("Component does not have GetMass method", "type", fmt.Sprintf("%T", compGeneric))
		}
	}
	if initialMass <= 1e-9 {
		log.Error("Total initial mass of the rocket is zero or negative after component initialization.")
		return nil
	}

	overallRocketCG, totalInertiaTensorBody, inverseTotalInertiaTensorBody := calculateCGAndInertia(initialMass, createdComponents, log)

	// --- 3. Create Basic Entity and Physics State ---
	basic := ecs.NewBasic()

	physicsStateCfg := &PhysicsStateConfig{
		OverallRocketCG:               overallRocketCG,
		TotalInertiaTensorBody:        totalInertiaTensorBody,
		InverseTotalInertiaTensorBody: inverseTotalInertiaTensorBody,
		InitialMass:                   initialMass,
		Motor:                         motor,
		Bodytube:                      bodytube,
		Nosecone:                      nosecone,
		Finset:                        finset,
		Parachute:                     parachute,
	}
	physicsState := createPhysicsState(basic, physicsStateCfg)

	// --- 4. Construct RocketEntity ---
	r := &RocketEntity{
		BasicEntity:  &basic,
		PhysicsState: physicsState,
		Mass: &types.Mass{
			BasicEntity: basic,
			Value:       initialMass, // Set mass using calculated value
		},
		components: createdComponents, // Assign the map of created components
	}

	return r
}

// AddComponent adds a component to the entity
func (r *RocketEntity) GetComponent(name string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.components[name]
}
