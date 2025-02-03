package systems

import (
	"math"

	"github.com/EngoEngine/ecs"
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
	entities  []PhysicsEntity
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
		entities: make([]PhysicsEntity, 0),
		rail: LaunchRail{
			Length:      length,
			Angle:       angleRad,
			Orientation: orientation,
		},
		onRail:    true,
		railExitY: length * math.Cos(angleRad), // Calculate Y position at rail exit
	}
}

// Add adds a physics entity to the launch rail system
func (s *LaunchRailSystem) Add(pe *PhysicsEntity) {
	s.entities = append(s.entities, PhysicsEntity{pe.Entity, pe.Position, pe.Velocity, pe.Acceleration, pe.Mass, pe.Motor, pe.Bodytube, pe.Nosecone, pe.Finset})
}

// Update applies launch rail constraints to entities
func (s *LaunchRailSystem) Update(dt float32) error {
	if !s.onRail {
		return nil
	}

	for _, entity := range s.entities {
		// While on rail, constrain motion to rail direction
		if s.onRail {
			// Get total acceleration magnitude including thrust
			totalAccel := entity.Acceleration.Y
			if entity.Motor != nil {
				thrust := entity.Motor.GetThrust()
				totalAccel += thrust / entity.Mass.Value
			}

			// Apply acceleration along rail direction
			angleRad := s.rail.Angle
			entity.Acceleration.X = float64(totalAccel) * math.Sin(angleRad)
			entity.Acceleration.Y = float64(totalAccel) * math.Cos(angleRad)
			entity.Acceleration.Z = 0

			// Update velocity along rail
			entity.Velocity.X = entity.Acceleration.X * float64(dt)
			entity.Velocity.Y = entity.Acceleration.Y * float64(dt)
			entity.Velocity.Z = 0

			// Update position along rail
			distanceAlongRail := math.Sqrt(
				entity.Position.X*entity.Position.X +
					entity.Position.Y*entity.Position.Y)

			// Check if we've reached end of rail
			if distanceAlongRail >= s.rail.Length {
				s.onRail = false
				return nil
			}
		}
	}
	return nil
}

// Priority returns the system priority
func (s *LaunchRailSystem) Priority() int {
	return 1 // Run before physics system
}
