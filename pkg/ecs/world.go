package ecs

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/ecs/entities"
	"github.com/bxrne/launchrail/pkg/ecs/systems"
)

// World manages all entities, components, and systems
type World struct {
	entities []entities.Entity
	systems  []systems.System
}

func NewWorld(rocket *entities.Rocket) *World {
	w := &World{
		entities: make([]entities.Entity, 0),
		systems:  make([]systems.System, 0),
	}

	w.AddEntity(rocket)
	return w
}

// AddEntity adds an entity to the World
func (w *World) AddEntity(e entities.Entity) {
	w.entities = append(w.entities, e)
}

// AddSystem adds a system to the NewWorld
func (w *World) AddSystem(s systems.System) {
	w.systems = append(w.systems, s)
}

// Update calls the Update method on all systems
func (w *World) Update(dt float64) error {
	for _, s := range w.systems {
		s.Update(dt)
	}

	for _, e := range w.entities {
		err := e.Update(dt)
		if err != nil {
			return err
		}
	}

	return nil
}

// String returns a string representation of the NewWorld
func (w *World) String() string {
	var entStr string
	for i, e := range w.entities {
		entStr += fmt.Sprintf("Entity %d: %s\n", i, e)
	}

	var sysStr string
	for i, s := range w.systems {
		sysStr += fmt.Sprintf("System %d: %s\n", i, s)
	}

	return fmt.Sprintf("World{Entities: %s, Systems: %s}", entStr, sysStr)
}

// Describe returns a string representation of the NewWorld
func (w *World) Describe() string {
	return fmt.Sprintf("%d entities and %d systems", len(w.entities), len(w.systems))
}
