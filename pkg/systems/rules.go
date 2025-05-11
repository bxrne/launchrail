package systems

import (
	"math"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

// RulesSystem enforces rules of flight
type RulesSystem struct {
	world      *ecs.World
	config     *config.Engine
	entities   []*states.PhysicsState
	hasLiftoff bool
	hasApogee  bool
	hasLanded  bool
	logger     logf.Logger
}

// GetLastEvent returns the last event detected by the rules system
func (s *RulesSystem) GetLastEvent() types.Event {
	if s.hasLanded {
		return types.Land
	}
	if s.hasApogee {
		return types.Apogee
	}
	if s.hasLiftoff {
		return types.Liftoff
	}
	return types.None
}

// NewRulesSystem creates a new RulesSystem
func NewRulesSystem(world *ecs.World, config *config.Engine, logger logf.Logger) *RulesSystem {
	return &RulesSystem{
		world:      world,
		config:     config,
		entities:   make([]*states.PhysicsState, 0),
		hasLiftoff: false,
		hasApogee:  false, // Initialize hasApogee
		hasLanded:  false, // Initialize hasLanded
		logger:     logger,
	}
}

// Add adds a physics entity to the rules system
func (s *RulesSystem) Add(entity *states.PhysicsState) {
	s.entities = append(s.entities, entity)
}

// Update applies rules of flight to entities
func (s *RulesSystem) Update(dt float64) error {
	for _, entity := range s.entities {
		s.ProcessRules(entity)
	}
	return nil
}

func (s *RulesSystem) ProcessRules(entity *states.PhysicsState) types.Event {
	if entity == nil || entity.Entity == nil || entity.Position == nil || entity.Velocity == nil || entity.Motor == nil {
		s.logger.Debug("ProcessRules: entity or critical component is nil")
		return types.None
	}

	// Check for Liftoff (Motor burning and off the ground/rail?)
	if !s.hasLiftoff && entity.Motor.FSM.Current() == components.StateBurning && entity.Position.Vec.Y > 0.1 /* Small tolerance */ {
		s.logger.Info("Liftoff detected", "entityID", entity.Entity.ID(), "altitude", entity.Position.Vec.Y)
		s.hasLiftoff = true
		return types.Liftoff
	}

	// Check for apogee
	// Only check for apogee *after* liftoff
	if s.hasLiftoff && !s.hasApogee && s.DetectApogee(entity) {
		s.hasApogee = true
		// The Info log for apogee is now within DetectApogee itself
		return types.Apogee
	}

	// Check for landing after apogee using ground tolerance
	groundTolerance := 0.1 // Default, consider making this configurable if not already via s.config
	if s.config != nil && s.config.Simulation.GroundTolerance > 0 { // Check if config and value exist
		groundTolerance = s.config.Simulation.GroundTolerance
	}

	// Only check for landing *after* apogee
	if s.hasApogee && !s.hasLanded && entity.Position.Vec.Y <= groundTolerance {
		s.logger.Info("Landing detected", "entityID", entity.Entity.ID(), "altitude", entity.Position.Vec.Y)
		entity.Position.Vec.Y = 0 // Normalize to ground
		entity.Velocity.Vec.Y = 0 // Stop vertical movement
		// Consider zeroing other velocity components if it's a full stop
		entity.Velocity.Vec.X = 0
		entity.Velocity.Vec.Z = 0
		entity.Acceleration.Vec.Y = 0 // Stop vertical acceleration
		s.hasLanded = true
		return types.Land
	}

	return types.None
}

func (s *RulesSystem) DetectApogee(entity *states.PhysicsState) bool {
	const velocityWindow = 0.5 // m/s window to detect velocity near zero

	// Ensure entity and its relevant fields are not nil before accessing them
	if entity == nil || entity.Position == nil || entity.Velocity == nil || entity.Motor == nil || entity.Parachute == nil {
		s.logger.Error("DetectApogee: entity or critical component is nil")
		return false
	}

	s.logger.Debug("DetectApogee called", "entityID", entity.Entity.ID(), "posY", entity.Position.Vec.Y, "velY", entity.Velocity.Vec.Y)

	// Must be near zero vertical velocity
	if math.Abs(entity.Velocity.Vec.Y) > velocityWindow {
		s.logger.Debug("DetectApogee: vertical velocity outside window", "velY", entity.Velocity.Vec.Y, "window", velocityWindow)
		return false
	}
	s.logger.Debug("DetectApogee: vertical velocity WITHIN window", "velY", entity.Velocity.Vec.Y)

	// Motor must be idle
	if entity.Motor.FSM.Current() != components.StateIdle {
		s.logger.Debug("DetectApogee: motor not idle", "motorState", string(entity.Motor.FSM.Current()))
		return false
	}
	s.logger.Debug("DetectApogee: motor is IDLE")

	// Check parachute status - ensure it's not already deployed
	if entity.Parachute.Deployed {
		s.logger.Debug("DetectApogee: parachute already deployed", "parachuteDeployed", entity.Parachute.Deployed)
		return false
	}
	s.logger.Debug("DetectApogee: parachute OK (exists and not deployed)")

	// Must be above ground
	if entity.Position.Vec.Y <= 0 {
		s.logger.Debug("DetectApogee: not above ground", "posY", entity.Position.Vec.Y)
		return false
	}
	s.logger.Debug("DetectApogee: IS ABOVE GROUND")

	// Deploy parachute if conditions met
	s.logger.Info("APOGEE DETECTED: Deploying parachute!", "entityID", entity.Entity.ID(), "altitude", entity.Position.Vec.Y)
	entity.Parachute.Deploy()

	return true
}
