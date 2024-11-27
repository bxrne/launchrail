package entities

import "sync"

// Entity is a unique identifier for an entity.
type Entity uint64

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
func (ecs *ECS) AddComponent(entity Entity, component interface{}, componentName string) {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	ecs.components[entity][componentName] = component
}

// GetComponent retrieves a component by entity and component name.
func (ecs *ECS) GetComponent(entity Entity, componentName string) interface{} {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	return ecs.components[entity][componentName]
}

// GetNextEntity returns the next entity
func (ecs *ECS) GetNextEntity() Entity {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	return ecs.nextEntity

}
