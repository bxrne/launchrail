package systems

import (
	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/types"
)

type RocketSystem struct{}

func NewRocketSystem() *RocketSystem {
    return &RocketSystem{}
}

func (s *RocketSystem) Priority() int {
    return 0 // Runs first
}

func (s *RocketSystem) Update(world *ecs.World, dt float64) error {
    // Query entities that have all required components
    entities := world.Query(
        ecs.ComponentMotor,
        ecs.ComponentPhysics,
        ecs.ComponentAerodynamics,
    )

    for _, entity := range entities {
        // Get components
        motorComp, _ := world.GetComponent(entity, ecs.ComponentMotor)
        physComp, _ := world.GetComponent(entity, ecs.ComponentPhysics)
        aeroComp, _ := world.GetComponent(entity, ecs.ComponentAerodynamics)

        // Type assert to specific interfaces
        motor := motorComp.(ecs.MotorComponent)
        physics := physComp.(ecs.PhysicsComponent)
        aero := aeroComp.(ecs.AerodynamicsComponent)

        // Apply motor thrust
        thrustForce := types.Vector3{Y: motor.GetThrust()}
        physics.AddForce(thrustForce)

        // Apply aerodynamic forces
        dragForce := aero.CalculateDrag(physics.GetVelocity())
        physics.AddForce(dragForce)
    }

    return nil
}
