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
				if entity.Motor != nil {
					if err := entity.Motor.Update(dt); err != nil {
						results <- result{err: err}
						continue
					}
				}
				netForce, err := s.calculateNetForce(entity, types.Vector3{})
				if err != nil {
					results <- result{err: err}
					continue
				}
				s.updateEntityState(entity, netForce, dt)
				results <- result{err: nil}
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
	if entity == nil {
		return fmt.Errorf("nil entity")
	}
	if entity.Position == nil || entity.Velocity == nil || entity.Acceleration == nil {
		return fmt.Errorf("entity missing required vectors")
	}
	if entity.Mass == nil {
		return fmt.Errorf("entity missing mass")
	}
	return nil
}

func (s *PhysicsSystem) handleGroundCollision(entity *states.PhysicsState) bool {
	if entity.Position.Vec.Y <= s.groundTolerance && entity.Velocity.Vec.Y <= 0 {
		entity.Position.Vec.Y = 0
		entity.Velocity.Vec.Y = 0
		entity.Acceleration.Vec.Y = 0
		return true
	}
	return false
}

func (s *PhysicsSystem) calculateNetForce(entity *states.PhysicsState, force types.Vector3) (float64, error) {
	if entity == nil || entity.Mass == nil || entity.Mass.Value <= 0 {
		return 0, fmt.Errorf("invalid entity or mass")
	}

	var netForce float64 = -s.gravity * entity.Mass.Value // Start with gravity

	// Add thrust if motor is active
	if entity.Motor != nil && !entity.Motor.IsCoasting() {
		thrust := entity.Motor.GetThrust()
		if !math.IsNaN(thrust) && !math.IsInf(thrust, 0) {
			netForce += thrust
		}
	}

	// Add drag force
	velocity := math.Sqrt(
		entity.Velocity.Vec.X*entity.Velocity.Vec.X +
			entity.Velocity.Vec.Y*entity.Velocity.Vec.Y)

	dragForceY := 0.0
	if velocity > 0 && !math.IsNaN(velocity) {
		rho := getAtmosphericDensity(entity.Position.Vec.Y)
		if !math.IsNaN(rho) && rho > 0 {
			if entity.Nosecone == nil || entity.Bodytube == nil {
				return 0, fmt.Errorf("missing geometry components: Nosecone or Bodytube is nil")
			}
			area := calculateReferenceArea(entity.Nosecone, entity.Bodytube)
			cd := 0.3 // Base drag coefficient
			if velocity > 100 {
				cd = 0.5
			}
			// Calculate drag force components
			dragForceY = -0.5 * rho * cd * area * velocity * entity.Velocity.Vec.Y
			netForce += dragForceY
		}
	}

	// Add external force
	if !math.IsNaN(force.Y) && !math.IsInf(force.Y, 0) {
		netForce += force.Y
	}

	return netForce, nil
}

func (s *PhysicsSystem) updateEntityState(entity *states.PhysicsState, netForce float64, dt float64) {
	if entity.Motor == nil {
		entity.Acceleration.Vec.X = 0
		entity.Acceleration.Vec.Y = 0
		entity.Acceleration.Vec.Z = 0
		entity.Velocity.Vec.X = 0
		entity.Velocity.Vec.Y = 0
		entity.Velocity.Vec.Z = 0
		entity.Position.Vec.X = 0
		entity.Position.Vec.Y = 0
		entity.Position.Vec.Z = 0
		return
	}
	if math.IsNaN(netForce) || math.IsInf(netForce, 0) {
		entity.Acceleration.Vec.Y = 0
		return // Skip update if force is invalid
	}

	// Calculate acceleration
	if entity.Mass == nil || entity.Mass.Value <= 0 {
		s.logger.Error("Invalid mass, skipping update", "entity_id", entity.Entity.ID(), "mass", entity.Mass)
		entity.Acceleration.Vec.Y = 0
		return
	}

	s.logger.Debug("Physics Update @ t=%.4f: TotalForce=%.4f, Mass=%.4f",
		0.0,      // Simulation time is not available in this context
		netForce, // Log the force vector (consider logging components X,Y,Z separately if needed)
		entity.Mass.Value)

	if entity.Mass.Value <= 1e-6 { // Avoid division by zero or near-zero mass
		s.logger.Debug("Skipping acceleration update: Mass near zero",
			"time", 0.0, // Simulation time is not available in this context
			"mass", entity.Mass.Value,
		)
		return // Skip update for this entity if mass is too small
	}

	s.logger.Debug("Calculating acceleration inputs",
		"netForceY", netForce,
		"mass", entity.Mass.Value,
	)
	newAcceleration := netForce / entity.Mass.Value

	s.logger.Debug("Post-acceleration calculation",
		"entity_id", entity.Entity.ID(),
		"new_acceleration", newAcceleration)

	if math.IsNaN(newAcceleration) || math.IsInf(newAcceleration, 0) {
		s.logger.Error("Calculated invalid acceleration (NaN/Inf), skipping update",
			"entity_id", entity.Entity.ID(),
			"net_force", netForce,
			"mass", entity.Mass.Value,
			"calculated_accel", newAcceleration)
		entity.Acceleration.Vec.Y = 0
		return
	}

	// Update translational state
	entity.Acceleration.Vec.Y = newAcceleration
	newVelocity := entity.Velocity.Vec.Y + entity.Acceleration.Vec.Y*dt
	newPosition := entity.Position.Vec.Y + newVelocity*dt
	if math.IsNaN(newVelocity) || math.IsInf(newVelocity, 0) {
		return
	}

	if math.IsNaN(newPosition) || math.IsInf(newPosition, 0) {
		return
	}

	// Apply ground constraint
	if newPosition <= 0 {
		s.handleGroundCollision(entity)
		return
	}

	entity.Velocity.Vec.Y = newVelocity
	entity.Position.Vec.Y = newPosition

	// Update rotation (6DOF)
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
