package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
)

// enum for events
type Event int

const (
	Apogee Event = iota
	Land
)

type RulesSystem struct {
	world     *ecs.World
	entities  []physicsEntity
	hadApogee bool      // Track if apogee has been reached
	maxAlt    float64   // Track max altitude for apogee detection
	altitudes []float64 // Track all altitudes for landing detection (calculus)
}

func NewRulesSystem(world *ecs.World) *RulesSystem {
	return &RulesSystem{
		world:     world,
		entities:  make([]physicsEntity, 0),
		hadApogee: false,
		maxAlt:    0,
		altitudes: make([]float64, 0),
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
		// Log data
		s.altitudes = append(s.altitudes, entity.Position.Z)

		if entity.Position.Z > s.maxAlt {
			s.maxAlt = entity.Position.Z
		}

		// check for apogee (rocket has ascended, delta vy is now 0 and its starts to fall and its motor is burnout)
		if entity.Velocity.Z < 0 && entity.Motor.FSM.GetState() == "idle" && !s.hadApogee {
			s.hadApogee = true
			return Apogee
		}

		// check for landing (rocket has landed, delta vy is now 0 and its starts to fall) and acceleration < 0
		if entity.Velocity.Z == 0 && s.hadApogee && entity.Acceleration.Z < 0 {
			return Land
		}

	}
	return -1
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
