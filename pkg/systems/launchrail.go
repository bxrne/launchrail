package systems

import (
	"fmt"
	"math"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/zerodha/logf"
)

// LaunchRail represents a launch rail
type LaunchRail struct {
	Length      float64
	Angle       float64 // Angle from vertical in degrees
	Orientation float64 // Compass orientation in degrees
}

// LaunchRailSystem constrains entities to a launch rail
type LaunchRailSystem struct {
	world     *ecs.World
	entities  []*states.PhysicsState
	rail      LaunchRail
	onRail    bool
	railExitY float64 // Y position at rail exit
	logger    *logf.Logger
}

// NewLaunchRailSystem creates a new LaunchRailSystem
func NewLaunchRailSystem(world *ecs.World, length, angleDegrees, orientationDegrees float64, logger *logf.Logger) *LaunchRailSystem {
	// Convert angle to radians
	angleRad := angleDegrees * math.Pi / 180.0

	if logger == nil {
		fmt.Println("Warning: LaunchRailSystem received a nil logger, using default logf logger.")
		defaultLogger := logf.New(logf.Opts{}) // Create the logger value
		logger = &defaultLogger                // Assign its address
	}

	return &LaunchRailSystem{
		world:    world,
		entities: make([]*states.PhysicsState, 0),
		rail: LaunchRail{
			Length:      length,
			Angle:       angleRad,                             // Stored in radians
			Orientation: orientationDegrees * math.Pi / 180.0, // Store in radians if used for calculations
		},
		onRail:    true,
		railExitY: length * math.Cos(angleRad), // Calculate Y position at rail exit
		logger:    logger,                      // Assign the passed (or default) logger
	}
}

// Add adds a physics entity to the launch rail system
func (s *LaunchRailSystem) Add(pe *states.PhysicsState) {
	s.entities = append(s.entities, pe)
}

// Update implements ecs.System interface
func (s *LaunchRailSystem) Update(dt float32) {
	_ = s.update(float64(dt))
}

// UpdateWithError implements System interface
func (s *LaunchRailSystem) UpdateWithError(dt float64) error {
	return s.update(dt)
}

// update is the internal update method
func (s *LaunchRailSystem) update(dt float64) error {
	if !s.onRail {
		return nil
	}

	for _, entity := range s.entities {
		if !s.onRail {
			continue
		}

		// Update motor first
		if err := s.updateEntityMotor(entity, dt); err != nil {
			return err
		}

		// Update entity physics on rail
		s.updateEntityOnRail(entity, dt)
	}
	return nil
}

// updateEntityMotor updates the motor for a given entity
func (s *LaunchRailSystem) updateEntityMotor(entity *states.PhysicsState, dt float64) error {
	if entity.Motor != nil {
		return entity.Motor.Update(dt)
	}
	return nil
}

// updateEntityOnRail handles the physics update for an entity while constrained to the rail
func (s *LaunchRailSystem) updateEntityOnRail(entity *states.PhysicsState, dt float64) {
	const gravity = 9.81
	angleRad := s.rail.Angle

	// Get thrust from motor (if any)
	thrust := 0.0
	if entity.Motor != nil {
		thrust = entity.Motor.GetThrust()
	}

	// Calculate forces along rail direction
	netForceAlongRail := calculateNetForceAlongRail(thrust, angleRad, entity.Mass.Value, gravity)

	if shouldResetMotion(netForceAlongRail, entity.Velocity.Vec.Y) {
		resetEntityMotion(entity)
	} else {
		// Apply forces and update motion
		applyForcesAndUpdateMotion(entity, netForceAlongRail, angleRad, dt)

		// Check if entity has reached the end of the rail
		if hasReachedEndOfRail(entity, s.rail.Length) {
			s.onRail = false
		}
	}
}

// calculateNetForceAlongRail calculates the net force along the rail direction
func calculateNetForceAlongRail(thrust, angleRad, mass, gravity float64) float64 {
	return thrust*math.Cos(angleRad) - mass*gravity*math.Sin(angleRad)
}

// shouldResetMotion determines if the entity's motion should be reset to zero
func shouldResetMotion(netForceAlongRail, velocityY float64) bool {
	return netForceAlongRail <= 0 && velocityY <= 0
}

// resetEntityMotion resets an entity's velocity and acceleration to zero
func resetEntityMotion(entity *states.PhysicsState) {
	entity.Velocity.Vec.X = 0
	entity.Velocity.Vec.Y = 0
	entity.Acceleration.Vec.X = 0
	entity.Acceleration.Vec.Y = 0
}

// applyForcesAndUpdateMotion updates acceleration, velocity, and position
func applyForcesAndUpdateMotion(entity *states.PhysicsState, netForceAlongRail, angleRad, dt float64) {
	// Apply forces along rail direction
	entity.Acceleration.Vec.X = netForceAlongRail / entity.Mass.Value * math.Sin(angleRad)
	entity.Acceleration.Vec.Y = netForceAlongRail / entity.Mass.Value * math.Cos(angleRad)

	// Update velocity and position
	entity.Velocity.Vec.X += entity.Acceleration.Vec.X * dt
	entity.Velocity.Vec.Y += entity.Acceleration.Vec.Y * dt
	entity.Position.Vec.X += entity.Velocity.Vec.X * dt
	entity.Position.Vec.Y += entity.Velocity.Vec.Y * dt
}

// hasReachedEndOfRail determines if the entity has reached the end of the rail
func hasReachedEndOfRail(entity *states.PhysicsState, railLength float64) bool {
	distanceAlongRail := math.Sqrt(
		entity.Position.Vec.X*entity.Position.Vec.X +
			entity.Position.Vec.Y*entity.Position.Vec.Y)

	return distanceAlongRail >= railLength
}

// InitializeRocketPosition sets the initial X and Y position of the rocket based on the launch rail.
func (lrs *LaunchRailSystem) InitializeRocketPosition(rocketState *states.PhysicsState) {
	lrs.logger.Debug("LaunchRailSystem.InitializeRocketPosition called")

	// Calculate initial position based on rail length and angle (angle is stored in radians)
	// Assuming Z is 0 for a 2D plane launch (X,Y)
	initialPosX := lrs.rail.Length * math.Sin(lrs.rail.Angle)
	initialPosY := lrs.rail.Length * math.Cos(lrs.rail.Angle)

	lrs.logger.Info("Setting initial rocket position from launch rail",
		"railLength", lrs.rail.Length,
		"railAngleDeg", lrs.rail.Angle*(180/math.Pi),
		"calculatedInitialPosX", initialPosX,
		"calculatedInitialPosY", initialPosY,
		"rocketPosY_before_set", rocketState.Position.Vec.Y)

	rocketState.Position.Vec.X = initialPosX
	rocketState.Position.Vec.Y = initialPosY
	// rocketState.Position.Vec.Z should remain as it is (likely 0 if not set elsewhere)

	lrs.logger.Info("Rocket initial position set", "newPosX", rocketState.Position.Vec.X, "newPosY", rocketState.Position.Vec.Y)
}

// GetRail returns the launch rail configuration
func (s *LaunchRailSystem) GetRail() LaunchRail {
	return s.rail
}

// GetEntities returns the tracked entities
func (s *LaunchRailSystem) GetEntities() []*states.PhysicsState {
	return s.entities
}

// IsOnRail returns whether the system is still constraining to the rail
func (s *LaunchRailSystem) IsOnRail() bool {
	return s.onRail
}
