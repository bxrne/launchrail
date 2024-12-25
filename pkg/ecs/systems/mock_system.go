package systems

// MockSystem implements systems.System interface for testing
type MockSystem struct {
	updateCalled bool
	updateDt     float64
	priority     int
}

func (s *MockSystem) Update(dt float64) {
	s.updateCalled = true
	s.updateDt = dt
}

func (s *MockSystem) Priority() int {
	return s.priority
}

func (s *MockSystem) GetUpdateCalled() bool {
	return s.updateCalled
}

func (s *MockSystem) GetUpdateDt() float64 {
	return s.updateDt
}

func NewMockSystem(priority int) *MockSystem {
	return &MockSystem{
		priority: priority,
	}
}
