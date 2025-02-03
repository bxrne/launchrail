package systems

import (
	"github.com/EngoEngine/ecs"
)

// RocketState represents the current state of the rocket for parasites
type RocketState struct {
	Time         float64
	Altitude     float64
	Velocity     float64
	Acceleration float64
	Thrust       float64
	MotorState   string
}

// ParasiteSystem defines the interface for parasite systems
type ParasiteSystem interface {
	Update(dt float32) error
	Add(entity *ecs.BasicEntity, components ...interface{})
	Priority() int
	Start(dataChan chan RocketState)
	Stop()
}
