package plugin_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/plugin"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zerodha/logf"
)

// MockSimulationPlugin implements the SimulationPlugin interface for testing
type MockSimulationPlugin struct {
	mock.Mock
}

func (m *MockSimulationPlugin) Initialize(log logf.Logger, cfg *config.Config) error {
	args := m.Called(log, cfg)
	return args.Error(0)
}

func (m *MockSimulationPlugin) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSimulationPlugin) Version() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSimulationPlugin) BeforeSimStep(state *states.PhysicsState) error {
	args := m.Called(state)
	return args.Error(0)
}

func (m *MockSimulationPlugin) AfterSimStep(state *states.PhysicsState) error {
	args := m.Called(state)
	return args.Error(0)
}

func (m *MockSimulationPlugin) Cleanup() error {
	args := m.Called()
	return args.Error(0)
}

// TestSimulationPluginInterface tests that the interface is properly defined
func TestSimulationPluginInterface(t *testing.T) {
	// Test that a mock implementation satisfies the interface
	var mockPlugin plugin.SimulationPlugin = &MockSimulationPlugin{}
	assert.NotNil(t, mockPlugin, "Mock plugin should implement SimulationPlugin interface")
}

// TestMockPluginInitialize tests the Initialize method
func TestMockPluginInitialize(t *testing.T) {
	mockPlugin := &MockSimulationPlugin{}
	logger := logf.New(logf.Opts{})
	cfg := &config.Config{}

	// Set up mock expectation
	mockPlugin.On("Initialize", logger, cfg).Return(nil)

	// Test the method
	err := mockPlugin.Initialize(logger, cfg)
	assert.NoError(t, err)

	// Verify expectations
	mockPlugin.AssertExpectations(t)
}

// TestMockPluginName tests the Name method
func TestMockPluginName(t *testing.T) {
	mockPlugin := &MockSimulationPlugin{}
	expectedName := "test-plugin"

	// Set up mock expectation
	mockPlugin.On("Name").Return(expectedName)

	// Test the method
	name := mockPlugin.Name()
	assert.Equal(t, expectedName, name)

	// Verify expectations
	mockPlugin.AssertExpectations(t)
}

// TestMockPluginVersion tests the Version method
func TestMockPluginVersion(t *testing.T) {
	mockPlugin := &MockSimulationPlugin{}
	expectedVersion := "1.0.0"

	// Set up mock expectation
	mockPlugin.On("Version").Return(expectedVersion)

	// Test the method
	version := mockPlugin.Version()
	assert.Equal(t, expectedVersion, version)

	// Verify expectations
	mockPlugin.AssertExpectations(t)
}

// TestMockPluginBeforeSimStep tests the BeforeSimStep method
func TestMockPluginBeforeSimStep(t *testing.T) {
	mockPlugin := &MockSimulationPlugin{}
	state := &states.PhysicsState{
		Time:     0.0,
		Position: &types.Position{Vec: types.Vector3{X: 0, Y: 0, Z: 0}},
		Velocity: &types.Velocity{Vec: types.Vector3{X: 0, Y: 0, Z: 0}},
	}

	// Set up mock expectation
	mockPlugin.On("BeforeSimStep", state).Return(nil)

	// Test the method
	err := mockPlugin.BeforeSimStep(state)
	assert.NoError(t, err)

	// Verify expectations
	mockPlugin.AssertExpectations(t)
}

// TestMockPluginAfterSimStep tests the AfterSimStep method
func TestMockPluginAfterSimStep(t *testing.T) {
	mockPlugin := &MockSimulationPlugin{}
	state := &states.PhysicsState{
		Time:     1.0,
		Position: &types.Position{Vec: types.Vector3{X: 0, Y: 10, Z: 0}},
		Velocity: &types.Velocity{Vec: types.Vector3{X: 0, Y: 5, Z: 0}},
	}

	// Set up mock expectation
	mockPlugin.On("AfterSimStep", state).Return(nil)

	// Test the method
	err := mockPlugin.AfterSimStep(state)
	assert.NoError(t, err)

	// Verify expectations
	mockPlugin.AssertExpectations(t)
}

// TestMockPluginCleanup tests the Cleanup method
func TestMockPluginCleanup(t *testing.T) {
	mockPlugin := &MockSimulationPlugin{}

	// Set up mock expectation
	mockPlugin.On("Cleanup").Return(nil)

	// Test the method
	err := mockPlugin.Cleanup()
	assert.NoError(t, err)

	// Verify expectations
	mockPlugin.AssertExpectations(t)
}

// TestPluginLifecycle tests a complete plugin lifecycle
func TestPluginLifecycle(t *testing.T) {
	mockPlugin := &MockSimulationPlugin{}
	logger := logf.New(logf.Opts{})
	cfg := &config.Config{}

	// Set up lifecycle expectations
	mockPlugin.On("Name").Return("lifecycle-test-plugin")
	mockPlugin.On("Version").Return("1.0.0")
	mockPlugin.On("Initialize", logger, cfg).Return(nil)

	// Create test states for simulation steps
	initialState := &states.PhysicsState{
		Time:     0.0,
		Position: &types.Position{Vec: types.Vector3{X: 0, Y: 0, Z: 0}},
		Velocity: &types.Velocity{Vec: types.Vector3{X: 0, Y: 0, Z: 0}},
	}
	updatedState := &states.PhysicsState{
		Time:     0.01,
		Position: &types.Position{Vec: types.Vector3{X: 0, Y: 0.1, Z: 0}},
		Velocity: &types.Velocity{Vec: types.Vector3{X: 0, Y: 1, Z: 0}},
	}

	mockPlugin.On("BeforeSimStep", initialState).Return(nil)
	mockPlugin.On("AfterSimStep", updatedState).Return(nil)
	mockPlugin.On("Cleanup").Return(nil)

	// Test lifecycle
	name := mockPlugin.Name()
	assert.Equal(t, "lifecycle-test-plugin", name)

	version := mockPlugin.Version()
	assert.Equal(t, "1.0.0", version)

	err := mockPlugin.Initialize(logger, cfg)
	assert.NoError(t, err)

	err = mockPlugin.BeforeSimStep(initialState)
	assert.NoError(t, err)

	err = mockPlugin.AfterSimStep(updatedState)
	assert.NoError(t, err)

	err = mockPlugin.Cleanup()
	assert.NoError(t, err)

	// Verify all expectations were met
	mockPlugin.AssertExpectations(t)
}

// TestPluginInterfaceContract tests that the interface contract is properly defined
func TestPluginInterfaceContract(t *testing.T) {
	// This test ensures that any type implementing SimulationPlugin
	// must have all the required methods

	// Create a function that requires a SimulationPlugin
	testPluginContract := func(p plugin.SimulationPlugin) {
		assert.NotNil(t, p, "Plugin should not be nil")

		// Test that all methods exist (will panic if they don't)
		name := p.Name()
		assert.IsType(t, "", name, "Name should return string")

		version := p.Version()
		assert.IsType(t, "", version, "Version should return string")

		// Methods that return errors should be callable
		// (We don't call them here as they might have side effects)
		assert.NotNil(t, p.Initialize, "Initialize method should exist")
		assert.NotNil(t, p.BeforeSimStep, "BeforeSimStep method should exist")
		assert.NotNil(t, p.AfterSimStep, "AfterSimStep method should exist")
		assert.NotNil(t, p.Cleanup, "Cleanup method should exist")
	}

	// Test with mock plugin
	mockPlugin := &MockSimulationPlugin{}
	mockPlugin.On("Name").Return("contract-test")
	mockPlugin.On("Version").Return("1.0.0")

	testPluginContract(mockPlugin)
	mockPlugin.AssertExpectations(t)
}
