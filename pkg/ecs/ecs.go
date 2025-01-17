package ecs

import (
	"sync"
)

// Entity represents a unique identifier for an entity.
type Entity int

// Component is an interface that all components must implement.
type Component interface {
	Update(dt float64) error
}

// ECS structure
type ECS struct {
	mu         sync.RWMutex
	entities   []Entity
	components map[Entity][]Component
}

// NewECS initializes a new ECS.
func NewECS() *ECS {
	return &ECS{
		entities:   []Entity{},
		components: make(map[Entity][]Component),
	}
}

// AddEntity adds a new entity with components.
func (ecs *ECS) AddEntity(entity Entity, components ...Component) {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()
	ecs.entities = append(ecs.entities, entity)
	ecs.components[entity] = components
}

// Update updates all entities in the ECS.
func (ecs *ECS) Update(dt float64) {
	ecs.mu.RLock()
	defer ecs.mu.RUnlock()
	for _, entity := range ecs.entities {
		for _, component := range ecs.components[entity] {
			if err := component.Update(dt); err != nil {
				// Handle error (e.g., log it)
			}
		}
	}
}

// String returns a string representation of the ECS.
func (ecs *ECS) String() string {
	ecs.mu.RLock()
	defer ecs.mu.RUnlock()
	var str string
	for _, entity := range ecs.entities {
		str += "Entity: " + string(entity) + "\n"

	}
	return str
}
