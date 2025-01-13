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
	entities   []entities.Entity
	components []components.Component
	systems    []systems.System
	mu         sync.RWMutex
}

func NewWorld(Rocket *entities.Rocket) *World {
	w := &World{
		entities:   make([]entities.Entity, 0),
		components: make([]components.Component, 0),
		systems:    make([]systems.System, 0),
	}

	w.AddEntity(Rocket)
	return w
}

// AddEntity adds an entity to the World
func (w *World) AddEntity(e entities.Entity) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.entities = append(w.entities, e)
}

// AddComponent adds a component to the NewWorld
func (w *World) AddComponent(c components.Component) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.components = append(w.components, c)
}

// AddSystem adds a system to the NewWorld
func (w *World) AddSystem(s systems.System) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.systems = append(w.systems, s)
}

// Update calls the Update method on all systems
func (w *World) Update(dt float64) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, s := range w.systems {
		s.Update(dt)
	}
}

// String returns a string representation of the NewWorld
func (w *World) String() string {
	return fmt.Sprintf("%d entities, %d components, and %d systems", len(w.entities), len(w.components), len(w.systems))
}

// Describe returns a string representation of the NewWorld
func (w *World) Describe() string {
	return fmt.Sprintf("%d entities, %d components, and %d systems", len(w.entities), len(w.components), len(w.systems))
}
