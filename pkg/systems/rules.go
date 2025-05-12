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

func (s *RulesSystem) ProcessRules(entity *states.PhysicsState) types.Event {
	if entity == nil || entity.Entity == nil || entity.Position == nil || entity.Velocity == nil || entity.Motor == nil {
		s.logger.Debug("ProcessRules: entity or critical component is nil")
		return types.None
	}

	// Check for Liftoff (Motor burning and off the ground/rail?)
	if !s.hasLiftoff && entity.Motor.FSM.Current() == components.StateBurning && entity.Position.Vec.Y > 0.1 /* Small tolerance */ {
		s.logger.Info("Liftoff detected", "entityID", entity.Entity.ID(), "altitude", entity.Position.Vec.Y)
		s.hasLiftoff = true
		return types.Liftoff
	}

	// Check for Apogee
	// Note: DetectApogee now sets s.hasApogee and returns if it was *newly* detected.
	apogeeNewlyDetected := s.DetectApogee(entity)
	if apogeeNewlyDetected {
		// Deployment logic is now handled below, considering parachute settings
	}

	// Parachute Deployment Logic
	if entity.Parachute != nil && !entity.Parachute.Deployed {
		switch entity.Parachute.Trigger {
		case components.ParachuteTriggerApogee:
			deployAtApogee := apogeeNewlyDetected && (entity.Position.Vec.Y >= entity.Parachute.DeployAltitude || entity.Parachute.DeployAltitude <= 0)
			deployPostApogeeAltitude := s.hasApogee && entity.Velocity.Vec.Y < 0 && entity.Position.Vec.Y <= entity.Parachute.DeployAltitude && entity.Parachute.DeployAltitude > 0

			if deployAtApogee {
				s.logger.Info("Deploying parachute at apogee", "entityID", entity.Entity.ID(), "altitude", entity.Position.Vec.Y, "deployAltitudeSetting", entity.Parachute.DeployAltitude)
				entity.Parachute.Deploy()
				return types.ParachuteDeploy // Can be Apogee and ParachuteDeploy
			} else if deployPostApogeeAltitude {
				s.logger.Info("Deploying parachute post-apogee at specified altitude", "entityID", entity.Entity.ID(), "altitude", entity.Position.Vec.Y, "deployAltitudeSetting", entity.Parachute.DeployAltitude)
				entity.Parachute.Deploy()
				return types.ParachuteDeploy
			}
		// TODO: Add cases for other trigger types like Altitude, Delay, etc.
		default:
			s.logger.Warn("Unknown parachute trigger type or no trigger logic implemented", "trigger", entity.Parachute.Trigger)
		}
	}

	// Check for landing after apogee using ground tolerance
	groundTolerance := 0.1                                          // Default, consider making this configurable if not already via s.config
	if s.config != nil && s.config.Simulation.GroundTolerance > 0 { // Check if config and value exist
		groundTolerance = s.config.Simulation.GroundTolerance
	}

	// Only check for landing *after* apogee
	if s.hasApogee && !s.hasLanded && entity.Position.Vec.Y <= groundTolerance {
		s.logger.Info("Landing detected", "entityID", entity.Entity.ID(), "altitude", entity.Position.Vec.Y)
		entity.Position.Vec.Y = 0 // Normalize to ground
		entity.Velocity.Vec.Y = 0 // Stop vertical movement
		// Consider zeroing other velocity components if it's a full stop
		entity.Velocity.Vec.X = 0
		entity.Velocity.Vec.Z = 0
		entity.Acceleration.Vec.Y = 0 // Stop vertical acceleration
		s.hasLanded = true
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
		return true       // Apogee newly detected this step
	}

	s.logger.Info("DetectApogee REJECT: vertical velocity outside window and not negative", "velocityY", entity.Velocity.Vec.Y, "targetWindow", velocityWindow)
	return false // Apogee not newly detected this step
}
