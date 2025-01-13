package components

// MockComponent implements components.Component interface for testing
type MockComponent struct {
	mockType string
}

func (m *MockComponent) Type() string {
	return m.mockType
}

func (m *MockComponent) Update(dt float64) {}

func NewMockComponent(mockType string) *MockComponent {
	return &MockComponent{
		mockType: mockType,
	}
}
