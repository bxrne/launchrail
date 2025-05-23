package systems

import (
	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/zerodha/logf"
)

// LogParasiteSystem logs rocket state data
type LogParasiteSystem struct {
	world    *ecs.World
	logger   logf.Logger
	entities []*states.PhysicsState // Change to pointer slice
	dataChan chan *states.PhysicsState
	done     chan struct{}
}

// NewLogParasiteSystem creates a new LogParasiteSystem
func NewLogParasiteSystem(world *ecs.World, logger logf.Logger) *LogParasiteSystem {
	return &LogParasiteSystem{
		world:    world,
		logger:   logger,
		entities: make([]*states.PhysicsState, 0),
		done:     make(chan struct{}),
	}
}

// Start the LogParasiteSystem
func (s *LogParasiteSystem) Start(dataChan chan *states.PhysicsState) {
	s.dataChan = dataChan
	go s.processData()
}

// Stop the LogParasiteSystem
func (s *LogParasiteSystem) Stop() {
	close(s.done)
}

// processData logs rocket state data
func (s *LogParasiteSystem) processData() {
	for {
		select {
		case state := <-s.dataChan:
			// Skip logging if essential components are nil
			if state == nil || state.Position == nil || state.Velocity == nil ||
				state.Acceleration == nil || state.Orientation == nil ||
				state.Motor == nil || state.Parachute == nil {
				s.logger.Error("invalid_state", "error", "missing required components")
				continue
			}

			s.logger.Debug("rocket_state",
				"time", state.Time,
				"altitude", state.Position.Vec.Y,
				"velocity", state.Velocity.Vec.Y,
				"acceleration", state.Acceleration.Vec.Y,
				"orientation", state.Orientation.Quat,
				"thrust", state.Motor.GetThrust(),
				"motor_state", state.Motor.GetState(),
				"parachute_deployed", state.Parachute.Deployed,
			)
		case <-s.done:
			return
		}
	}
}

// Update implements ecs.System interface
func (s *LogParasiteSystem) Update(dt float32) {
	_ = s.update(float64(dt))
}

// UpdateWithError implements System interface
func (s *LogParasiteSystem) UpdateWithError(dt float64) error {
	return s.update(dt)
}

// update is the internal update method
func (s *LogParasiteSystem) update(dt float64) error {
	// No need to track time here - data comes from simulation state
	return nil
}

// Add adds entities to the system
func (s *LogParasiteSystem) Add(pe *states.PhysicsState) {
	s.entities = append(s.entities, pe) // Store pointer directly
}
