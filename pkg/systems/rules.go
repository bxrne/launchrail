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
		hasApogee:  false, // Initialize hasApogee
		hasLanded:  false, // Initialize hasLanded
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

// checkValidEntity verifies if entity and its components are valid
func (s *RulesSystem) checkValidEntity(entity *states.PhysicsState) bool {
	if entity == nil || entity.Entity == nil || entity.Position == nil || entity.Velocity == nil || entity.Motor == nil {
		s.logger.Debug("ProcessRules: entity or critical component is nil")
		return false
	}
	return true
}

// checkLiftoff detects if liftoff has occurred
func (s *RulesSystem) checkLiftoff(entity *states.PhysicsState) bool {
	if !s.hasLiftoff && entity.Motor.FSM.Current() == components.StateBurning && entity.Position.Vec.Y > 0.1 /* Small tolerance */ {
		s.logger.Info("Liftoff detected", "entityID", entity.Entity.ID(), "altitude", entity.Position.Vec.Y)
		return true
	}
	return false
}

// handleParachute manages parachute deployment logic
func (s *RulesSystem) handleParachute(entity *states.PhysicsState, apogeeNewlyDetected bool) types.Event {
	if entity.Parachute == nil || entity.Parachute.Deployed {
		return types.None
	}

	switch entity.Parachute.Trigger {
	case components.ParachuteTriggerApogee:
		return s.handleApogeeParachute(entity, apogeeNewlyDetected)
	default:
		s.logger.Warn("Unknown parachute trigger type or no trigger logic implemented", "trigger", entity.Parachute.Trigger)
		return types.None
	}
}

// handleApogeeParachute handles parachute deployment at apogee
func (s *RulesSystem) handleApogeeParachute(entity *states.PhysicsState, apogeeNewlyDetected bool) types.Event {
	deployAtApogee := apogeeNewlyDetected && (entity.Position.Vec.Y >= entity.Parachute.DeployAltitude || entity.Parachute.DeployAltitude <= 0)

	if deployAtApogee {
		s.logger.Info("Deploying parachute at apogee", "entityID", entity.Entity.ID(), "altitude", entity.Position.Vec.Y, "deployAltitudeSetting", entity.Parachute.DeployAltitude)
		entity.Parachute.Deploy()
		return types.ParachuteDeploy
	}

	deployPostApogeeAltitude := s.hasApogee && entity.Velocity.Vec.Y < 0 &&
		entity.Position.Vec.Y <= entity.Parachute.DeployAltitude &&
		entity.Parachute.DeployAltitude > 0

	if deployPostApogeeAltitude {
		s.logger.Info("Deploying parachute post-apogee at specified altitude", "entityID", entity.Entity.ID(), "altitude", entity.Position.Vec.Y, "deployAltitudeSetting", entity.Parachute.DeployAltitude)
		entity.Parachute.Deploy()
		return types.ParachuteDeploy
	}

	return types.None
}

// checkLanding detects if landing has occurred
func (s *RulesSystem) checkLanding(entity *states.PhysicsState) (bool, float64) {
	groundTolerance := 0.1 // Default tolerance
	if s.config != nil && s.config.Simulation.GroundTolerance > 0 {
		groundTolerance = s.config.Simulation.GroundTolerance
	}

	return s.hasApogee && !s.hasLanded && entity.Position.Vec.Y <= groundTolerance, groundTolerance
}

// handleLanding processes landing event
func (s *RulesSystem) handleLanding(entity *states.PhysicsState) {
	s.logger.Info("Landing detected", "entityID", entity.Entity.ID(), "altitude", entity.Position.Vec.Y)
	entity.Position.Vec.Y = 0 // Normalize to ground
	entity.Velocity.Vec.Y = 0 // Stop vertical movement
	entity.Velocity.Vec.X = 0
	entity.Velocity.Vec.Z = 0
	entity.Acceleration.Vec.Y = 0 // Stop vertical acceleration
	s.hasLanded = true
}

func (s *RulesSystem) ProcessRules(entity *states.PhysicsState) types.Event {
	if !s.checkValidEntity(entity) {
		return types.None
	}

	// Check for Liftoff
	if s.checkLiftoff(entity) {
		s.hasLiftoff = true
		return types.Liftoff
	}

	// Check for Apogee
	apogeeNewlyDetected := s.DetectApogee(entity)

	// Parachute Deployment Logic
	parachuteEvent := s.handleParachute(entity, apogeeNewlyDetected)
	if parachuteEvent != types.None {
		return parachuteEvent
	}

	// Check for landing after apogee
	landed, _ := s.checkLanding(entity)
	if landed {
		s.handleLanding(entity)
		return types.Land
	}

	return types.None
}

func (s *RulesSystem) DetectApogee(entity *states.PhysicsState) bool {
	// Increased velocity window to be more lenient in detecting apogee
	const velocityWindow = 1.0 // m/s window to detect velocity near zero or negative

	// Ensure entity and its relevant fields are not nil before accessing them
	if entity == nil || entity.Entity == nil || entity.Position == nil || entity.Velocity == nil || entity.Motor == nil || entity.Parachute == nil {
		s.logger.Error("DetectApogee: entity or critical component is nil")
		return false
	}

	// Log initial state when function is called
	s.logger.Debug("DetectApogee called",
		"entityID", entity.Entity.ID(),
		"altitude", entity.Position.Vec.Y,
		"velocityY", entity.Velocity.Vec.Y,
		"motorState", string(entity.Motor.FSM.Current()),
		"hasLiftoff", s.hasLiftoff,
		"hasApogee", s.hasApogee,
	)

	// Conditions for apogee detection:
	// 1. Vertical velocity is near zero or negative.
	// 2. Liftoff has occurred.
	// 3. Apogee has not already been detected.
	if (math.Abs(entity.Velocity.Vec.Y) < velocityWindow || entity.Velocity.Vec.Y < 0) &&
		s.hasLiftoff && !s.hasApogee {
		// Additional check: Motor should not be burning
		if entity.Motor.FSM.Current() == components.StateBurning {
			s.logger.Info("DetectApogee REJECT: motor still burning", "motorState", string(entity.Motor.FSM.Current()))
			return false
		}

		s.logger.Info("APOGEE DETECTED", "entityID", entity.Entity.ID(), "altitude", entity.Position.Vec.Y, "velocityY", entity.Velocity.Vec.Y)
		s.hasApogee = true // Set that apogee has been detected
		return true        // Apogee newly detected this step
	}

	s.logger.Info("DetectApogee REJECT: vertical velocity outside window and not negative", "velocityY", entity.Velocity.Vec.Y, "targetWindow", velocityWindow)
	return false // Apogee not newly detected this step
}
