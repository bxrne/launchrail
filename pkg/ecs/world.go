package ecs

import (
	"fmt"
	"sync"
)

type EntityID uint64

type World struct {
    mu sync.RWMutex

    entities    map[EntityID]struct{}
    components  map[EntityID]map[string]Component
    systems     []System
    nextEntity  EntityID
}

func NewWorld() *World {
    return &World{
        entities:    make(map[EntityID]struct{}),
        components:  make(map[EntityID]map[string]Component),
        systems:     make([]System, 0),
        nextEntity:  1,
    }
}

func (w *World) CreateEntity() EntityID {
    w.mu.Lock()
    defer w.mu.Unlock()

    id := w.nextEntity
    w.nextEntity++

    w.entities[id] = struct{}{}
    w.components[id] = make(map[string]Component)
    return id
}

func (w *World) AddComponent(entity EntityID, component Component) error {
    w.mu.Lock()
    defer w.mu.Unlock()

    if _, exists := w.entities[entity]; !exists {
        return fmt.Errorf("entity %d does not exist", entity)
    }

    w.components[entity][component.Type()] = component
    return nil
}

func (w *World) GetComponent(entity EntityID, componentType string) (Component, bool) {
    w.mu.RLock()
    defer w.mu.RUnlock()

    if components, exists := w.components[entity]; exists {
        if component, exists := components[componentType]; exists {
            return component, true
        }
    }
    return nil, false
}

func (w *World) Query(componentTypes ...string) []EntityID {
    w.mu.RLock()
    defer w.mu.RUnlock()

    var matches []EntityID
entityLoop:
    for entity := range w.entities {
        for _, compType := range componentTypes {
            if _, exists := w.components[entity][compType]; !exists {
                continue entityLoop
            }
        }
        matches = append(matches, entity)
    }
    return matches
}

func (w *World) Update(dt float64) error {
    // Sort systems by priority if needed
    for _, system := range w.systems {
        if err := system.Update(w, dt); err != nil {
            return fmt.Errorf("system update failed: %w", err)
        }
    }
    return nil
}

func (w *World) AddSystem(system System) {
    w.mu.Lock()
    defer w.mu.Unlock()
    w.systems = append(w.systems, system)
}
