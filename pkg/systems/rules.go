package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
)

type Event int

const (
	None Event = iota - 1
	Apogee
	Land
)

type RulesSystem struct {
	world     *ecs.World
	entities  []physicsEntity
	hadApogee bool    // Track if apogee has been reached
	maxAlt    float64 // Track max altitude for apogee detection
}

func NewRulesSystem(world *ecs.World) *RulesSystem {
	return &RulesSystem{
		world:     world,
		entities:  make([]physicsEntity, 0),
		hadApogee: false,
		maxAlt:    0,
	}
}

func (s *RulesSystem) Add(entity *ecs.BasicEntity, pos *components.Position,
	vel *components.Velocity, acc *components.Acceleration, mass *components.Mass,
	motor *components.Motor, bodytube *components.Bodytube, nosecone *components.Nosecone,
	finset *components.TrapezoidFinset) {
	s.entities = append(s.entities, physicsEntity{entity, pos, vel, acc, mass, motor, bodytube, nosecone, finset})
}

func (s *RulesSystem) Update(dt float32) Event {
	for _, entity := range s.entities {
		// Update max altitude and track current values
		currentAlt := entity.Position.Y
		currentVel := entity.Velocity.Y

		// Check for landing (negative altitude means ground penetration)
		if currentAlt <= 0 {
			return Land
		}

		// Track maximum altitude
		if currentAlt > s.maxAlt {
			s.maxAlt = currentAlt
		}

		// Detect apogee when velocity changes from positive to negative
		// and motor has finished burning
		if !s.hadApogee && currentVel < 0 && entity.Motor.IsCoasting() {
			s.hadApogee = true
			return Apogee
		}
	}

	return None
}

func (s *RulesSystem) Remove(basic ecs.BasicEntity) {
	var deleteIndex int = -1
	for i, e := range s.entities {
		if e.BasicEntity.ID() == basic.ID() {
			deleteIndex = i
			break
		}
	}
	if deleteIndex >= 0 {
		s.entities = append(s.entities[:deleteIndex], s.entities[deleteIndex+1:]...)
	}
}

func (s *RulesSystem) Priority() int {
	return 100 // Run after all other systems
}
