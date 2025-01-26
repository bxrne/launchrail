package systems_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/systems"
	"github.com/bxrne/launchrail/pkg/ecs/types"
	"github.com/stretchr/testify/assert"
)

type MockMotorComponent struct {
	thrust float64
}

func (m *MockMotorComponent) Type() string {
	return ecs.ComponentMotor
}

func (m *MockMotorComponent) GetThrust() float64 {
	return m.thrust
}

func (m *MockMotorComponent) Update(dt float64) error {
	return nil
}

func (m *MockMotorComponent) GetMass() float64 {
	return 10.0
}

type MockAerodynamicsComponent struct{}

func (a *MockAerodynamicsComponent) Type() string {
	return ecs.ComponentAerodynamics
}

func (a *MockAerodynamicsComponent) CalculateDrag(velocity types.Vector3) types.Vector3 {
	return types.Vector3{X: -velocity.X, Y: -velocity.Y, Z: -velocity.Z} // Simple drag calculation
}

func (a *MockAerodynamicsComponent) Update(dt float64) error {
	return nil
}

func TestRocketSystem_Update(t *testing.T) {
	world := ecs.NewWorld()
	rocketSystem := systems.NewRocketSystem()

	// Create a new entity with necessary components
	entity := world.CreateEntity()

	motorComp := &MockMotorComponent{thrust: 10.0}
	physicsComp := &MockPhysicsComponent{}
	aeroComp := &MockAerodynamicsComponent{}

	assert.NoError(t, world.AddComponent(entity, motorComp))
	assert.NoError(t, world.AddComponent(entity, physicsComp))
	assert.NoError(t, world.AddComponent(entity, aeroComp))

	// Run the RocketSystem Update
	err := rocketSystem.Update(world, 0.016) // dt = 16ms
	assert.NoError(t, err, "RocketSystem update should not return an error")

	// Verify that forces have been added to the physics component
	assert.Len(t, physicsComp.forces, 2, "There should be two forces applied: thrust and drag")

	// Check if the thrust force was added correctly
	thrustForce := types.Vector3{Y: motorComp.GetThrust()}
	assert.Contains(t, physicsComp.forces, thrustForce, "Thrust force should be applied to the physics component")

	// Check if drag force is applied
	dragForce := aeroComp.CalculateDrag(physicsComp.GetVelocity())
	assert.Contains(t, physicsComp.forces, dragForce, "Drag force should be applied to the physics component")
}

func TestRocketSystem_Update_EntityWithoutRequiredComponents(t *testing.T) {
	world := ecs.NewWorld()
	rocketSystem := systems.NewRocketSystem()

	// Create a new entity with only one component
	entity := world.CreateEntity()

	motorComp := &MockMotorComponent{thrust: 10.0}
	assert.NoError(t, world.AddComponent(entity, motorComp))

	// Run the RocketSystem Update, which should not process the entity because it's missing components
	err := rocketSystem.Update(world, 0.016)
	assert.NoError(t, err, "RocketSystem update should not return an error when entity is missing components")

	// Ensure no forces are applied due to missing components
	_, exists := world.GetComponent(entity, ecs.ComponentPhysics)
	assert.False(t, exists, "Physics component should not exist for this entity")

	_, exists = world.GetComponent(entity, ecs.ComponentAerodynamics)
	assert.False(t, exists, "Aerodynamics component should not exist for this entity")
}
