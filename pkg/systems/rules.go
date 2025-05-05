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
	if entity == nil || entity.Position == nil || entity.Velocity == nil || entity.Motor == nil {
		return types.None
	}

	// Check for Liftoff (Motor burning and off the ground/rail?)
	if !s.hasLiftoff && entity.Motor.FSM.Current() == components.StateBurning && entity.Position.Vec.Y > 0.1 /* Small tolerance */ {
		s.hasLiftoff = true
		return types.Liftoff
	}

	// Check for apogee
	// Only check for apogee *after* liftoff
	if s.hasLiftoff && !s.hasApogee && s.DetectApogee(entity) {
		s.hasApogee = true
		return types.Apogee
	}

	// Check for landing after apogee using ground tolerance
	groundTolerance := s.config.Simulation.GroundTolerance
	// Only check for landing *after* apogee
	if s.hasApogee && !s.hasLanded && entity.Position.Vec.Y <= groundTolerance {
		entity.Position.Vec.Y = 0
		entity.Velocity.Vec.Y = 0
		entity.Acceleration.Vec.Y = 0
		s.hasLanded = true
		return types.Land
	}

	return types.None
}

func (s *RulesSystem) DetectApogee(entity *states.PhysicsState) bool {
	const velocityWindow = 0.5 // m/s window to detect velocity near zero

	// Must be near zero vertical velocity
	if math.Abs(entity.Velocity.Vec.Y) > velocityWindow { // Return false ONLY if velocity is OUTSIDE the window
		return false
	}

	// Motor must be idle
	if entity.Motor == nil || entity.Motor.FSM.Current() != components.StateIdle {
		return false
	}

	// Check parachute status
	if entity.Parachute == nil || entity.Parachute.Deployed {
		return false
	}

	// Must be above ground
	if entity.Position == nil || entity.Position.Vec.Y <= 0 {
		return false
	}

	// Deploy parachute if conditions met
	entity.Parachute.Deploy()

	return true
}
