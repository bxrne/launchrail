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

// ParasiteSystem extends the base System interface
type ParasiteSystem interface {
	ecs.System
	Start(dataChan chan RocketState)
	Stop()
}
