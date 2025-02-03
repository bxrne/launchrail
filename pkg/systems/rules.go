package systems

import (
	"github.com/EngoEngine/ecs"
)

// Event represents a significant event in flight
type Event int

const (
	None Event = iota - 1
	Apogee
	Land
)

// RulesSystem enforces rules of flight
type RulesSystem struct {
	world     *ecs.World
	entities  []PhysicsEntity
	hadApogee bool    // Track if apogee has been reached
	maxAlt    float64 // Track max altitude for apogee detection
}

// NewRulesSystem creates a new RulesSystem
func NewRulesSystem(world *ecs.World) *RulesSystem {
	return &RulesSystem{
		world:     world,
		entities:  make([]PhysicsEntity, 0),
		hadApogee: false,
		maxAlt:    0,
	}
}

// Add adds a physics entity to the rules system
func (s *RulesSystem) Add(pe *PhysicsEntity) {
	s.entities = append(s.entities, PhysicsEntity{pe.Entity, pe.Position, pe.Velocity, pe.Acceleration, pe.Mass, pe.Motor, pe.Bodytube, pe.Nosecone, pe.Finset})
}

// Update applies rules of flight to entities
func (s *RulesSystem) Update(dt float32) error {
	event := s.processRules(dt)
	// Process the event if needed
	switch event {
	case Apogee:
		// Do something
	case Land:
		// Do something
	}

	return nil
}

func (s *RulesSystem) processRules(dt float32) Event {
	// Move existing Update logic here
	for _, entity := range s.entities {
		if event := s.checkApogee(entity); event != None {
			return event
		}
		if event := s.checkLanding(entity); event != None {
			return event
		}
	}
	return None
}

func (s *RulesSystem) checkApogee(entity PhysicsEntity) Event {
	currentAlt := entity.Position.Vec.Y
	currentVel := entity.Velocity.Vec.Y

	if currentAlt > s.maxAlt {
		s.maxAlt = currentAlt
	}

	if !s.hadApogee && currentVel < 0 {
		motorState := entity.Motor.GetState()
		if motorState == "BURNOUT" || motorState == "COASTING" {
			s.hadApogee = true
			return Apogee
		}
	}
	return None
}

func (s *RulesSystem) checkLanding(entity PhysicsEntity) Event {
	// Only check for landing if we've passed apogee
	if !s.hadApogee {
		return None
	}

	// Check if we've hit the ground with downward velocity
	if entity.Position.Vec.Y <= 0 && entity.Velocity.Vec.Y < 0 {
		// Reset state on landing
		entity.Position.Vec.Y = 0
		entity.Velocity.Vec.Y = 0
		entity.Acceleration.Vec.Y = 0
		entity.Motor.SetState("LANDED")
		return Land
	}
	return None
}

// Remove removes an entity from the rules system
func (s *RulesSystem) Remove(basic ecs.BasicEntity) {
	var deleteIndex = -1
	for i, e := range s.entities {
		if e.Entity.ID() == basic.ID() {
			deleteIndex = i
			break
		}
	}
	if deleteIndex >= 0 {
		s.entities = append(s.entities[:deleteIndex], s.entities[deleteIndex+1:]...)
	}
}

// Priority returns the system priority
func (s *RulesSystem) Priority() int {
	return 100 // Run after all other systems
}
