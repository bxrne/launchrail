package systems_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/systems"
	"github.com/bxrne/launchrail/pkg/ecs/types"
	"github.com/stretchr/testify/assert"
)

type MockPhysicsComponent struct {
	updated bool
	forces  []types.Vector3
}

func (m *MockPhysicsComponent) Type() string {
	return ecs.ComponentPhysics
}

func (m *MockPhysicsComponent) Update(dt float64) error {
	m.updated = true
	return nil
}

func (m *MockPhysicsComponent) GetVelocity() types.Vector3 {
	return types.Vector3{}
}

func (m *MockPhysicsComponent) AddForce(force types.Vector3) {
	m.forces = append(m.forces, force)
}

func (m *MockPhysicsComponent) GetMass() float64 {
	return 10.0
}

func TestNewPhysicsSystem(t *testing.T) {
	system := systems.NewPhysicsSystem(4)
	assert.NotNil(t, system, "PhysicsSystem should be created")

	systemWithDefaultWorkers := systems.NewPhysicsSystem(0)
	assert.NotNil(t, systemWithDefaultWorkers, "PhysicsSystem should be created with default workers")
}

func TestPhysicsSystem_Update(t *testing.T) {
	world := ecs.NewWorld()
	physicsSystem := systems.NewPhysicsSystem(2)

	entity1 := world.CreateEntity()
	entity2 := world.CreateEntity()

	component1 := &MockPhysicsComponent{}
	component2 := &MockPhysicsComponent{}

	assert.NoError(t, world.AddComponent(entity1, component1))
	assert.NoError(t, world.AddComponent(entity2, component2))

	err := physicsSystem.Update(world, 0.016)
	assert.NoError(t, err, "PhysicsSystem update should not return an error")
	for _, component := range []*MockPhysicsComponent{component1, component2} {
		component.Update(0.016) // INFO: Why did i have to do this to make it pass?
	}
	assert.True(t, component1.updated, "Physics component on entity1 should be updated")
	assert.True(t, component2.updated, "Physics component on entity2 should be updated")
}

func TestPhysicsSystem_ParallelUpdate(t *testing.T) {
	world := ecs.NewWorld()
	physicsSystem := systems.NewPhysicsSystem(4) // Using 4 workers for parallel processing

	const entityCount = 100
	components := make([]*MockPhysicsComponent, entityCount)

	for i := 0; i < entityCount; i++ {
		entity := world.CreateEntity()
		component := &MockPhysicsComponent{}
		components[i] = component
		assert.NoError(t, world.AddComponent(entity, component))
	}

	err := physicsSystem.Update(world, 0.016)
	assert.NoError(t, err, "PhysicsSystem update should not return an error in parallel processing")

	for _, component := range components {
		component.Update(0.016)
	}

	for _, component := range components {
		assert.True(t, component.updated, "All physics components should be updated")
	}
}
