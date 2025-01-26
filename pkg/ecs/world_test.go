package ecs_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/stretchr/testify/assert"
)

type MockComponent struct {
	name string
}

func (m MockComponent) Type() string {
	return m.name
}

func (m MockComponent) Update(dt float64) error {
	return nil
}

type MockSystem struct{}

func (ms MockSystem) Update(w *ecs.World, dt float64) error {
	return nil
}

func (ms MockSystem) Priority() int {
	return 0
}

func TestWorld_CreateEntity(t *testing.T) {
	world := ecs.NewWorld()

	entity := world.CreateEntity()
	assert.NotZero(t, entity, "Entity ID should not be zero")

	entity2 := world.CreateEntity()
	assert.NotEqual(t, entity, entity2, "Each entity should have a unique ID")
}

func TestWorld_AddComponent(t *testing.T) {
	world := ecs.NewWorld()
	entity := world.CreateEntity()

	component := MockComponent{name: "testComponent"}
	err := world.AddComponent(entity, component)
	assert.NoError(t, err, "Adding a component to an existing entity should not produce an error")

	retrieved, exists := world.GetComponent(entity, "testComponent")
	assert.True(t, exists, "Component should exist")
	assert.Equal(t, component, retrieved, "Retrieved component should match the added component")
}

func TestWorld_AddComponent_InvalidEntity(t *testing.T) {
	world := ecs.NewWorld()
	component := MockComponent{name: "testComponent"}

	err := world.AddComponent(999, component) // Non-existent entity
	assert.Error(t, err, "Adding a component to a non-existent entity should produce an error")
}

func TestWorld_Query(t *testing.T) {
	world := ecs.NewWorld()
	entity1 := world.CreateEntity()
	entity2 := world.CreateEntity()

	componentA := MockComponent{name: "ComponentA"}
	componentB := MockComponent{name: "ComponentB"}

	_ = world.AddComponent(entity1, componentA)
	_ = world.AddComponent(entity2, componentA)
	_ = world.AddComponent(entity2, componentB)

	matchedEntities := world.Query("ComponentA")
	assert.Len(t, matchedEntities, 2, "Querying by ComponentA should return two entities")

	matchedEntities = world.Query("ComponentA", "ComponentB")
	assert.Len(t, matchedEntities, 1, "Querying by ComponentA and ComponentB should return one entity")
	assert.Equal(t, entity2, matchedEntities[0], "Entity2 should match the query")
}

func TestWorld_Update(t *testing.T) {
	world := ecs.NewWorld()
	system := MockSystem{}

	world.AddSystem(system)

	err := world.Update(0.016) // Assume a frame time of 16ms
	assert.NoError(t, err, "Updating the world should not produce an error")
}

func TestWorld_AddSystem(t *testing.T) {
	world := ecs.NewWorld()
	system := MockSystem{}

	world.AddSystem(system)
}
