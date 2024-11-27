package systems

import (
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/entities"
)

// State represents the state of an entity at a given tick.
type State struct {
	Position     components.Position
	Velocity     components.Velocity
	Acceleration components.Acceleration
}

// StateManager manages the states of entities over time.
type StateManager struct {
	states []map[entities.Entity]State
}

// NewStateManager creates a new StateManager.
func NewStateManager() *StateManager {
	return &StateManager{states: []map[entities.Entity]State{}}
}

// StoreState stores the current state of all entities.
func (sm *StateManager) StoreState(ecs *entities.ECS) {
	currentState := make(map[entities.Entity]State)
	for entity := entities.Entity(1); entity < ecs.GetNextEntity(); entity++ {
		position, _ := ecs.GetComponent(entity, "Position").(*components.Position)
		velocity, _ := ecs.GetComponent(entity, "Velocity").(*components.Velocity)
		acceleration, _ := ecs.GetComponent(entity, "Acceleration").(*components.Acceleration)
		currentState[entity] = State{
			Position:     *position,
			Velocity:     *velocity,
			Acceleration: *acceleration,
		}
	}
	sm.states = append(sm.states, currentState)
}

// GetStates returns the stored states.
func (sm *StateManager) GetStates() []map[entities.Entity]State {
	return sm.states
}
