package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
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
	config    *config.Config
	entities  []*PhysicsEntity
	hasApogee bool
}

// GetLastEvent returns the last event detected by the rules system
func (s *RulesSystem) GetLastEvent() Event {
	if s.hasApogee {
		return Apogee
	}
	return None
}

// NewRulesSystem creates a new RulesSystem
func NewRulesSystem(world *ecs.World, config *config.Config) *RulesSystem {
	return &RulesSystem{
		world:     world,
		config:    config,
		entities:  make([]*PhysicsEntity, 0),
		hasApogee: false,
	}
}

// Add adds a physics entity to the rules system
func (s *RulesSystem) Add(entity *PhysicsEntity) {
	s.entities = append(s.entities, entity)
}

// Update applies rules of flight to entities
func (s *RulesSystem) Update(dt float64) error {
	for _, entity := range s.entities {
		s.processRules(entity)
	}
	return nil
}

func (s *RulesSystem) processRules(entity *PhysicsEntity) Event {
	if entity == nil || entity.Position == nil || entity.Velocity == nil || entity.Motor == nil {
		return None
	}

	// Check for apogee
	if !s.hasApogee &&
		entity.Motor.GetState() == "BURNOUT" &&
		entity.Velocity.Vec.Y < 0 {
		s.hasApogee = true
		return Apogee
	}

	// Check for landing after apogee
	if s.hasApogee && entity.Position.Vec.Y <= 0 {
		entity.Position.Vec.Y = 0
		entity.Velocity.Vec.Y = 0
		entity.Acceleration.Vec.Y = 0
		return Land
	}

	return None
}

// Remove removes an entity from the rules system
func (s *RulesSystem) Remove(basic ecs.BasicEntity) {
	var del = -1
	for i, e := range s.entities {
		if e.Entity.ID() == basic.ID() {
			del = i
			break
		}
	}
	if del >= 0 {
		s.entities = append(s.entities[:del], s.entities[del+1:]...)
	}
}

// Priority returns the system priority
func (s *RulesSystem) Priority() int {
	return 100 // Run after all other systems
}
