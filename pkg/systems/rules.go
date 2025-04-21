package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
)

// Event represents a significant event in flight
type Event int

const (
	None Event = iota
	Apogee
	Land
)

// RulesSystem enforces rules of flight
type RulesSystem struct {
	world     *ecs.World
	config    *config.Engine
	entities  []*states.PhysicsState
	hasApogee bool
	hasLanded bool // Add this field
}

// GetLastEvent returns the last event detected by the rules system
func (s *RulesSystem) GetLastEvent() Event {
	if s.hasLanded {
		return Land
	}
	if s.hasApogee {
		return Apogee
	}
	return None
}

// NewRulesSystem creates a new RulesSystem
func NewRulesSystem(world *ecs.World, config *config.Engine) *RulesSystem {
	return &RulesSystem{
		world:     world,
		config:    config,
		entities:  make([]*states.PhysicsState, 0),
		hasApogee: false,
	}
}

// Add adds a physics entity to the rules system
func (s *RulesSystem) Add(entity *states.PhysicsState) {
	s.entities = append(s.entities, entity)
}

// Update applies rules of flight to entities
func (s *RulesSystem) Update(dt float64) error {
	for _, entity := range s.entities {
		s.processRules(entity)
	}
	return nil
}

func (s *RulesSystem) processRules(entity *states.PhysicsState) Event {
	if entity == nil || entity.Position == nil || entity.Velocity == nil || entity.Motor == nil {
		return None
	}

	// Check for apogee
	if !s.hasApogee && s.detectApogee(entity) {
		s.hasApogee = true
		return Apogee
	}

	// Check for landing after apogee using ground tolerance
	groundTolerance := s.config.Simulation.GroundTolerance
	if s.hasApogee && !s.hasLanded && entity.Position.Vec.Y <= groundTolerance {
		entity.Position.Vec.Y = 0
		entity.Velocity.Vec.Y = 0
		entity.Acceleration.Vec.Y = 0
		s.hasLanded = true
		return Land
	}

	return None
}

func (s *RulesSystem) detectApogee(entity *states.PhysicsState) bool {
	// Must be moving downward to be at apogee
	if entity.Velocity.Vec.Y >= 0 {
		return false
	}

	// Must be coasting (motor burned out)
	if entity.Motor.FSM.Current() != components.StateIdle {
		return false
	}

	// Must be above ground
	if entity.Position.Vec.Y <= s.config.Simulation.GroundTolerance {
		return false
	}

	// If we get here, we're at apogee
	if entity.Parachute != nil && !entity.Parachute.IsDeployed() {
		entity.Parachute.Deploy()
		return true
	}

	return false
}
