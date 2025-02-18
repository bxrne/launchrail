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
	s.entities = append(s.entities, PhysicsEntity{pe.Entity, pe.Position, pe.Velocity, pe.Acceleration, pe.Mass, pe.Motor, pe.Bodytube, pe.Nosecone, pe.Finset, pe.Parachute})
}

// Update applies launch rail constraints to entities
func (s *LaunchRailSystem) Update(dt float32) error {
	if !s.onRail {
		return nil
	}

	const gravity = 9.81

	for _, entity := range s.entities {
		if !s.onRail {
			continue
		}

		// Get total acceleration magnitude including thrust
		angleRad := s.rail.Angle
		thrust := 0.0
		if entity.Motor != nil {
			thrust = entity.Motor.GetThrust()
		}

		netForce := thrust - (entity.Mass.Value * gravity)
		if netForce <= 0 {
			// Hold rocket on rail, zero out motion
			entity.Acceleration.Vec.X = 0
			entity.Acceleration.Vec.Y = 0
			entity.Velocity.Vec.X = 0
			entity.Velocity.Vec.Y = 0
		} else {
			// Rocket can leave rail
			entity.Acceleration.Vec.X = netForce / entity.Mass.Value * math.Sin(angleRad)
			entity.Acceleration.Vec.Y = netForce / entity.Mass.Value * math.Cos(angleRad)
			entity.Velocity.Vec.X += entity.Acceleration.Vec.X * float64(dt)
			entity.Velocity.Vec.Y += entity.Acceleration.Vec.Y * float64(dt)
			entity.Position.Vec.X += entity.Velocity.Vec.X * float64(dt)
			entity.Position.Vec.Y += entity.Velocity.Vec.Y * float64(dt)
		}

		// Update position along rail
		distanceAlongRail := math.Sqrt(
			entity.Position.Vec.X*entity.Position.Vec.X +
				entity.Position.Vec.Y*entity.Position.Vec.Y)

		// Check if we've reached end of rail
		if distanceAlongRail >= s.rail.Length {
			s.onRail = false
			return nil
		}
	}
	return nil
}

// Priority returns the system priority
func (s *LaunchRailSystem) Priority() int {
	return 1 // Run before physics system
}
