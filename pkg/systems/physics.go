package systems

import (
	"fmt"
	"math"
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/barrowman"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

// Use object pools for vectors and matrices
var (
	vectorPool = sync.Pool{
		New: func() interface{} {
			return &types.Vector3{}
		},
	}
)

// PhysicsSystem calculates forces on entities
type PhysicsSystem struct {
	world           *ecs.World
	entities        []*states.PhysicsState // Changed to store pointers
	cpCalculator    *barrowman.CPCalculator
	workers         int
	gravity         float64
	groundTolerance float64
	logger          logf.Logger // Use non-pointer interface type
}

// Remove removes an entity from the system
func (s *PhysicsSystem) Remove(basic ecs.BasicEntity) {
	var deleteIndex int
	for i, entity := range s.entities {
		if entity.Entity.ID() == basic.ID() {
			deleteIndex = i
			break
		}
	}
	s.entities = append(s.entities[:deleteIndex], s.entities[deleteIndex+1:]...)
}

// NewPhysicsSystem creates a new PhysicsSystem
func NewPhysicsSystem(world *ecs.World, cfg *config.Engine, logger logf.Logger, workers int) *PhysicsSystem {
	if workers <= 0 {
		workers = 4 // Default number of workers
	}
	return &PhysicsSystem{
		world:           world,
		entities:        make([]*states.PhysicsState, 0),
		cpCalculator:    barrowman.NewCPCalculator(), // Initialize calculator
		workers:         workers,
		gravity:         cfg.Options.Launchsite.Atmosphere.ISAConfiguration.GravitationalAccel,
		groundTolerance: cfg.Simulation.GroundTolerance,
		logger:          logger, // Use logger directly
	}
}

// Update applies forces to entities
func (s *PhysicsSystem) Update(dt float64) error {
	if dt <= 0 || math.IsNaN(dt) {
		return fmt.Errorf("invalid timestep: %v", dt)
	}

	type result struct {
		err error
	}

	workChan := make(chan *states.PhysicsState)
	results := make(chan result, len(s.entities))
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entity := range workChan {
				if err := s.validateEntity(entity); err != nil {
					results <- result{err: err}
					continue
				}

				// --- Force Accumulation and State Update ---
				// Reset forces for this physics update step
				entity.AccumulatedForce = types.Vector3{}

				// Calculate Gravity Force (acts downwards in global frame)
				if entity.Mass != nil && entity.Mass.Value > 0 { // Only apply gravity if mass is valid
					gravityForce := types.Vector3{Y: -s.gravity * entity.Mass.Value}
					entity.AccumulatedForce = entity.AccumulatedForce.Add(gravityForce)
				} else {
					s.logger.Warn("Skipping gravity calculation due to invalid mass", "entity_id", entity.Entity.ID())
				}

				// Calculate Thrust Force (acts along rocket body axis, rotated to global frame)
				var thrustForce types.Vector3
				if entity.Motor != nil && !entity.Motor.IsCoasting() && entity.Orientation != nil && entity.Orientation.Quat != (types.Quaternion{}) {
					thrustMagnitude := entity.Motor.GetThrust()
					// Assume thrust acts along the rocket's local +Y axis
					localThrust := types.Vector3{Y: thrustMagnitude}
					thrustForce = *entity.Orientation.Quat.RotateVector(&localThrust)
					entity.AccumulatedForce = entity.AccumulatedForce.Add(thrustForce)
				} else if entity.Motor != nil && !entity.Motor.IsCoasting() {
					// Fallback if orientation is missing/invalid? Assume vertical thrust? Log warning?
					s.logger.Warn("Calculating thrust without valid orientation, assuming vertical", "entity_id", entity.Entity.ID())
					thrustMagnitude := entity.Motor.GetThrust()
					thrustForce = types.Vector3{Y: thrustMagnitude} // Vertical thrust
					entity.AccumulatedForce = entity.AccumulatedForce.Add(thrustForce)
				}

				// Calculate Drag Force (acts opposite to velocity vector)
				velocityMag := entity.Velocity.Vec.Magnitude()
				if velocityMag > 1e-6 { // Only calculate drag if moving significantly
					rho := getAtmosphericDensity(entity.Position.Vec.Y)
					if !math.IsNaN(rho) && rho > 0 {
						// Check for required geometry BEFORE calculating area
						if entity.Nosecone == nil || entity.Bodytube == nil {
							results <- result{err: fmt.Errorf("entity %d missing geometry components for drag calculation", entity.Entity.ID())}
							continue // Skip to the next entity
						}
						area := calculateReferenceArea(entity.Nosecone, entity.Bodytube)
						cd := 0.3 // Simplified constant drag coefficient for now
						dragMagnitude := 0.5 * rho * velocityMag * velocityMag * cd * area

						// Drag force vector opposes velocity
						dragDirection := entity.Velocity.Vec.Normalize().MultiplyScalar(-1)
						dragForce := dragDirection.MultiplyScalar(dragMagnitude)
						entity.AccumulatedForce = entity.AccumulatedForce.Add(dragForce)
					}
				}

				// *** Apply accumulated forces and update state ***
				s.updateEntityState(entity, entity.AccumulatedForce, dt)

				results <- result{err: nil} // Report success for this entity
			}
		}()
	}

	// Send all entities to workChan
	for _, entity := range s.entities {
		workChan <- entity
	}
	close(workChan)

	// Wait for all workers to finish in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect the first error encountered
	for res := range results {
		if res.err != nil {
			return res.err
		}
	}

	return nil
}

func (s *PhysicsSystem) validateEntity(entity *states.PhysicsState) error {
	entityID := uint64(0) // Default ID if entity or entity.Entity is nil
	if entity != nil && entity.Entity != nil {
		entityID = entity.Entity.ID()
	}
	if entity == nil {
		return fmt.Errorf("nil entity")
	}
	if entity.Position == nil || entity.Velocity == nil || entity.Acceleration == nil {
		return fmt.Errorf("entity missing required vectors")
	}
	if entity.Mass == nil {
		return fmt.Errorf("entity missing mass")
	}
	if entity.Mass.Value <= 0 {
		return fmt.Errorf("invalid entity mass value: %v", entity.Mass.Value)
	}
	// Check for essential geometry needed for drag/aerodynamics
	if entity.Nosecone == nil {
		return fmt.Errorf("entity %d missing Nosecone component", entityID)
	}
	if entity.Bodytube == nil {
		return fmt.Errorf("entity %d missing Bodytube component", entityID)
	}
	return nil
}

func (s *PhysicsSystem) handleGroundCollision(entity *states.PhysicsState) bool {
	// Check if entity is at or below ground level and moving downwards or stationary vertically.
	if entity.Position.Vec.Y <= s.groundTolerance && entity.Velocity.Vec.Y <= 0 {
		// Set vertical position exactly to ground level
		entity.Position.Vec.Y = 0
		// Zero out all velocity components
		entity.Velocity.Vec = types.Vector3{}
		// Zero out all acceleration components
		entity.Acceleration.Vec = types.Vector3{}
		// Optionally, zero out angular velocity as well?
		if entity.AngularVelocity != nil {
			*entity.AngularVelocity = types.Vector3{}
		}
		if entity.AngularAcceleration != nil {
			*entity.AngularAcceleration = types.Vector3{}
		}

		// Set landing event if not already landed
		if entity.CurrentEvent != types.Land {
			entity.CurrentEvent = types.Land
		}
		s.logger.Debug("Ground collision detected", "entity_id", entity.Entity.ID())

		return true // Collision handled
	}
	return false // No collision detected or handled
}

func (s *PhysicsSystem) updateEntityState(entity *states.PhysicsState, netForce types.Vector3, dt float64) {
	// Existing nil checks for entity and motor can remain...
	if entity.Motor == nil {
		// Reset or handle appropriately if motor is nil
		entity.Acceleration.Vec = types.Vector3{}
		entity.Velocity.Vec = types.Vector3{}
		entity.Position.Vec = types.Vector3{}
		return
	}

	// Check for invalid mass
	if entity.Mass == nil || entity.Mass.Value <= 1e-9 { // Use a small epsilon for mass check
		s.logger.Warn("Invalid or near-zero mass, skipping acceleration update", "entity_id", entity.Entity.ID(), "mass", entity.Mass)
		// Decide how to handle: zero acceleration? return?
		entity.Acceleration.Vec = types.Vector3{}
		// Maybe should still integrate velocity/position based on existing velocity?
		// For now, let's just zero accel and continue to integrate vel/pos below.
	} else {
		// Calculate acceleration vector: a = F/m
		newAccelerationVec := netForce.DivideScalar(entity.Mass.Value) // Assumes Vector3.DivideScalar exists

		// Check for NaN/Inf in calculated acceleration
		if math.IsNaN(newAccelerationVec.X) || math.IsInf(newAccelerationVec.X, 0) ||
			math.IsNaN(newAccelerationVec.Y) || math.IsInf(newAccelerationVec.Y, 0) ||
			math.IsNaN(newAccelerationVec.Z) || math.IsInf(newAccelerationVec.Z, 0) {
			s.logger.Error("Calculated invalid acceleration vector (NaN/Inf), skipping update",
				"entity_id", entity.Entity.ID(),
				"net_force", netForce,
				"mass", entity.Mass.Value,
				"calculated_accel", newAccelerationVec)
			entity.Acceleration.Vec = types.Vector3{} // Reset acceleration
			// Continue to integrate velocity/position
		} else {
			entity.Acceleration.Vec = newAccelerationVec
		}
	}

	// --- Integrate Linear Velocity & Position (Using Full Vectors) ---
	// v_new = v_old + a * dt
	newVelocityVec := entity.Velocity.Vec.Add(entity.Acceleration.Vec.MultiplyScalar(dt))

	// p_new = p_old + v_new * dt (Using updated velocity)
	newPositionVec := entity.Position.Vec.Add(newVelocityVec.MultiplyScalar(dt))

	// Check for NaN/Inf in new velocity/position
	if math.IsNaN(newVelocityVec.X) || math.IsInf(newVelocityVec.X, 0) || /* ... check Y, Z */
		math.IsNaN(newPositionVec.X) || math.IsInf(newPositionVec.X, 0) /* ... check Y, Z */ {
		s.logger.Error("Calculated invalid velocity or position vector (NaN/Inf), skipping state application",
			"entity_id", entity.Entity.ID(),
			"new_vel", newVelocityVec,
			"new_pos", newPositionVec)
		return // Don't apply the invalid state
	}

	// Apply ground constraint (based on Y position)
	if newPositionVec.Y <= 0 {
		s.handleGroundCollision(entity) // Assuming handleGroundCollision sets pos/vel correctly
		// Note: handleGroundCollision might need adjustment if it only modified Y components before.
	} else {
		// Update state only if not colliding
		entity.Velocity.Vec = newVelocityVec
		entity.Position.Vec = newPositionVec
	}

	// --- Angular Update (Remains the same) ---
	if entity.Orientation != nil && entity.AngularVelocity != nil && entity.AngularAcceleration != nil {
		// Simple angular integration
		entity.AngularVelocity.X += entity.AngularAcceleration.X * dt
		entity.AngularVelocity.Y += entity.AngularAcceleration.Y * dt
		entity.AngularVelocity.Z += entity.AngularAcceleration.Z * dt

		// Integrate orientation quaternion
		entity.Orientation.Quat.Integrate(*entity.AngularVelocity, dt)
	}
}

// Add adds an entity to the system
func (s *PhysicsSystem) Add(pe *states.PhysicsState) {
	s.entities = append(s.entities, pe) // Store pointer directly
}

// String returns the system name
func (s *PhysicsSystem) String() string {
	return "PhysicsSystem"
}
