package ecs

import (
	"fmt"
	"sync"

	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/ecs/entities"
	"github.com/bxrne/launchrail/pkg/ecs/systems"
)

// World manages all entities, components, and systems
type World struct {
	entities   map[entities.EntityID]bool
	components map[string]map[entities.EntityID]components.Component
	systems    []systems.System
	nextID     entities.EntityID
	mu         sync.RWMutex
}

func NewWorld() *World {
	return &World{
		entities:   make(map[entities.EntityID]bool),
		components: make(map[string]map[entities.EntityID]components.Component),
		systems:    make([]systems.System, 0),
	}
}

// CreateEntity creates a new entity and returns its ID
func (w *World) CreateEntity() entities.EntityID {
	w.mu.Lock()
	defer w.mu.Unlock()

	id := w.nextID
	w.nextID++
	w.entities[id] = true
	return id
}

// AddComponent adds a component to an entity
func (w *World) AddComponent(entity entities.EntityID, component components.Component) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.entities[entity] {
		return fmt.Errorf("entity %d does not exist", entity)
	}

	componentType := component.Type()
	if w.components[componentType] == nil {
		w.components[componentType] = make(map[entities.EntityID]components.Component)
	}
	w.components[componentType][entity] = component
	return nil
}

// GetComponent returns a component for an entity
func (w *World) GetComponent(entity entities.EntityID, componentType string) (components.Component, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if components, ok := w.components[componentType]; ok {
		if component, ok := components[entity]; ok {
			return component, true
		}
	}
	return nil, false
}

// AddSystem adds a system to the world
func (w *World) AddSystem(system systems.System) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.systems = append(w.systems, system)
	// Sort systems by priority
	// You might want to implement a proper sort here
}

// Update updates all systems
func (w *World) Update(dt float64) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, system := range w.systems {
		system.Update(dt)
	}
}
