package systems

import (
	"math"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/states"
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
	entities  []*states.PhysicsState // Change to pointer slice
	rail      LaunchRail
	onRail    bool
	railExitY float64 // Y position at rail exit
}

// Add adds a physics entity to the launch rail system
func NewLaunchRailSystem(world *ecs.World, length, angle, orientation float64) *LaunchRailSystem {
	// Convert angle to radians
	angleRad := angle * math.Pi / 180.0

	return &LaunchRailSystem{
		world:    world,
		entities: make([]*states.PhysicsState, 0),
		rail: LaunchRail{
			Length:      length,
			Angle:       angleRad,
			Orientation: orientation,
		},
		onRail:    true,
		railExitY: length * math.Cos(angleRad), // Calculate Y position at rail exit
	}
}

// Add adds entities to the system
func (s *LaunchRailSystem) Add(pe *states.PhysicsState) {
	s.entities = append(s.entities, pe) // Store pointer directly
}

// Update applies launch rail constraints to entities
func (s *LaunchRailSystem) Update(dt float64) error {
	if !s.onRail {
		return nil
	}

	const gravity = 9.81

	for _, entity := range s.entities {
		if !s.onRail {
			continue
		}

		// Update motor first
		if entity.Motor != nil {
			if err := entity.Motor.Update(dt); err != nil {
				return err
			}
		}

		// Get total acceleration magnitude including thrust
		angleRad := s.rail.Angle
		thrust := 0.0
		if entity.Motor != nil {
			thrust = entity.Motor.GetThrust()
		}

		// Calculate forces along rail direction
		netForceAlongRail := thrust*math.Cos(angleRad) - entity.Mass.Value*gravity*math.Sin(angleRad)

		if netForceAlongRail <= 0 && entity.Velocity.Vec.Y <= 0 {
			entity.Velocity.Vec.X = 0
			entity.Velocity.Vec.Y = 0
			entity.Acceleration.Vec.X = 0
			entity.Acceleration.Vec.Y = 0
		} else {
			// Apply forces along rail direction
			entity.Acceleration.Vec.X = netForceAlongRail / entity.Mass.Value * math.Sin(angleRad)
			entity.Acceleration.Vec.Y = netForceAlongRail / entity.Mass.Value * math.Cos(angleRad)

			// Update velocity and position
			entity.Velocity.Vec.X += entity.Acceleration.Vec.X * dt
			entity.Velocity.Vec.Y += entity.Acceleration.Vec.Y * dt
			entity.Position.Vec.X += entity.Velocity.Vec.X * dt
			entity.Position.Vec.Y += entity.Velocity.Vec.Y * dt

			// Check if we've reached end of rail
			distanceAlongRail := math.Sqrt(
				entity.Position.Vec.X*entity.Position.Vec.X +
					entity.Position.Vec.Y*entity.Position.Vec.Y)

			if distanceAlongRail >= s.rail.Length {
				s.onRail = false
			}
		}
	}
	return nil
}

// Priority returns the system priority
func (s *LaunchRailSystem) Priority() int {
	return 1 // Run before physics system
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
