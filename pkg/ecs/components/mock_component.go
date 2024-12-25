package components

// MockComponent implements components.Component interface for testing
type MockComponent struct {
	mockType string
}

func (m *MockComponent) Type() string {
	return m.mockType
}

func NewMockComponent(mockType string) *MockComponent {
	return &MockComponent{
		mockType: mockType,
	}
}
