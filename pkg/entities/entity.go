package entities

import (
	"strconv"
	"sync"
)

// Entity is a unique identifier for an entity.
type Entity uint64

// String implements the Stringer interface and returns the string representation of the Entity
func (e Entity) String() string {
	return strconv.FormatUint(uint64(e), 10)
}

// ECS is the entity component system.
type ECS struct {
	entities   map[Entity]bool
	components map[Entity]map[string]interface{}
	nextEntity Entity
	mu         sync.Mutex
}

// NewECS creates a new ECS.
func NewECS() *ECS {
	return &ECS{
		entities:   make(map[Entity]bool),
		components: make(map[Entity]map[string]interface{}),
		nextEntity: 1,
	}
}

// CreateEntity creates a new entity.
func (ecs *ECS) CreateEntity() Entity {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	entity := ecs.nextEntity
	ecs.entities[entity] = true
	ecs.components[entity] = make(map[string]interface{})
	ecs.nextEntity++
	return entity
}

// AddComponent adds a component to an entity.
func (ecs *ECS) AddComponent(entity Entity, component interface{}) {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	ecs.components[entity]["component"] = component
}

// GetComponent retrieves a component by entity.
func (ecs *ECS) GetComponent(entity Entity) interface{} {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	return ecs.components[entity]["component"]
}

// GetNextEntity retrieves the next entity
func (ecs *ECS) GetNextEntity() Entity {
	return ecs.nextEntity
}
