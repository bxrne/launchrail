package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/states"
)

// Event represents a significant event during simulation
type EventLog struct {
	Time    float64
	Type    string
	Details string
}

// ParasiteSystem extends the base System interface
type ParasiteSystem interface {
	ecs.System
	Start(dataChan chan *states.PhysicsState)
	Stop()
}
