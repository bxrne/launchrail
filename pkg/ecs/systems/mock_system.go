package systems

import "github.com/bxrne/launchrail/pkg/ecs/entities"

// MockSystem is a mock system for testing
type MockSystem struct {
	target entities.Entity
}

// Apply applies the system to an entity
func (s *MockSystem) Apply() {
}

// NewMockSystem creates a new MockSystem instance
func NewMockSystem(target entities.Entity) *MockSystem {
	return &MockSystem{target: target}
}
