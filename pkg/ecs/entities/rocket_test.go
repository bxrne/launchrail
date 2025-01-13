package entities_test

import (
	"fmt"
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/ecs/entities"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN a mass and a component set WHEN NewRocket is called THEN a new Rocket instance is returned with the given mass and components
func TestNewRocket(t *testing.T) {
	mass := 1.0
	component := components.NewMockComponent("mock")
	rocket := entities.NewRocket(mass, component)
	assert.Equal(t, rocket.Mass, mass)
	assert.Equal(t, rocket.Components[0], component)
}

// TEST: GIVEN a rocket instance WHEN String is called THEN a string representation of the rocket is returned
func TestRocket_String(t *testing.T) {
	rocket := entities.NewRocket(1.0)
	expected := fmt.Sprintf("Rocket{ID: %d, Position: %v, Velocity: %v, Acceleration: %v, Mass: %.2f, Forces: %v, Components: %v}", rocket.ID, rocket.Position, rocket.Velocity, rocket.Acceleration, rocket.Mass, rocket.Forces, rocket.Components)
	assert.Equal(t, rocket.String(), expected)
}

// TEST: GIVEN a rocket instance WHEN Describe is called THEN a string representation of the rocket is returned
func TestRocket_Describe(t *testing.T) {
	rocket := entities.NewRocket(1.0)
	expected := fmt.Sprintf("Rocket{ID: %d, Position: %v, Velocity: %v, Acceleration: %v, Mass: %.2f}", rocket.ID, rocket.Position, rocket.Velocity, rocket.Acceleration, rocket.Mass)
	assert.Equal(t, rocket.Describe(), expected)
}

// TEST: GIVEN a rocket instance and a component WHEN AddComponent is called THEN the component is added to the Rocket
func TestRocket_AddComponent(t *testing.T) {
	rocket := entities.NewRocket(1.0)
	component := components.NewMockComponent("mock")
	rocket.AddComponent(component)
	assert.Equal(t, rocket.Components[0], component)
}

// TEST: GIVEN a rocket instance and a component WHEN RemoveComponent is called THEN the component is removed from the rocket
func TestRocket_RemoveComponent(t *testing.T) {
	rocket := entities.NewRocket(1.0)
	component := components.NewMockComponent("mock")
	rocket.AddComponent(component)
	rocket.RemoveComponent(component)
	assert.Equal(t, len(rocket.Components), 0)
}

// TEST: GIVEN a rocket instance and a delta time WHEN Update is called THEN the rocket is updated
func TestRocket_Update(t *testing.T) {
	rocket := entities.NewRocket(1.0)
	component := components.NewMockComponent("mock")
	rocket.AddComponent(component)
	rocket.Update(1.0)

}
