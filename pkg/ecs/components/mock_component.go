package components

// MockComponent implements components.Component interface for testing
type MockComponent struct {
	mockType string
}

// String returns the type of the component
func (m *MockComponent) String() string {
	return m.mockType
}

// Update updates the component
func (m *MockComponent) Update(dt float64) error {
	// INFO: Empty, just meeting interface requirements
	return nil
}

// NewMockComponent creates a new MockComponent instance
func NewMockComponent(mockType string) *MockComponent {
	return &MockComponent{
		mockType: mockType,
	}
}
