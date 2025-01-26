package entities_test

import (
	"math"
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bxrne/launchrail/pkg/ecs/entities"
)

var md = &thrustcurves.MotorData{
	Thrust:    [][]float64{{0, 10}, {1, 20}, {2, 15}},
	TotalMass: 50,
	BurnTime:  2.0,
	AvgThrust: 15,
}

func TestNewRocket(t *testing.T) {
	motor := components.NewMotor(100.0, md)
	nosecone := &entities.Nosecone{Radius: 0.1}

	rocket := entities.NewRocket(1, 10.0, motor, nosecone, 0.5)

	assert.Equal(t, 1, rocket.ID)
	assert.Equal(t, 10.0, rocket.Physics.Mass)
	assert.Equal(t, 0.5, rocket.Aerodynamics.DragCoefficient)
	assert.Equal(t, math.Pi*(0.1*0.1), rocket.Aerodynamics.Area)
}

func TestRocketUpdate(t *testing.T) {
	t.Run("Invalid Timestep", func(t *testing.T) {
		motor := components.NewMotor(100.0, md)
		nosecone := &entities.Nosecone{Radius: 0.1}
		rocket := entities.NewRocket(1, 10.0, motor, nosecone, 0.5)

		err := rocket.Update(-0.1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid timestep")
	})

	t.Run("Successful Update", func(t *testing.T) {
		motor := components.NewMotor(100.0, md)
		nosecone := &entities.Nosecone{Radius: 0.1}
		rocket := entities.NewRocket(1, 10.0, motor, nosecone, 0.5)

		initialVelocity := rocket.Physics.Velocity

		err := rocket.Update(0.1)
		require.NoError(t, err)

		// Check that velocity has changed
		assert.NotEqual(t, initialVelocity, rocket.Physics.Velocity)
	})
}

func TestRocketString(t *testing.T) {
	motor := components.NewMotor(100.0, md)
	nosecone := &entities.Nosecone{Radius: 0.1}
	rocket := entities.NewRocket(1, 10.0, motor, nosecone, 0.5)

	stringRep := rocket.String()
	assert.Contains(t, stringRep, "Rocket{ID: 1")
	assert.Contains(t, stringRep, "Motor:")
	assert.Contains(t, stringRep, "Nosecone:")
	assert.Contains(t, stringRep, "Physics:")
	assert.Contains(t, stringRep, "Aerodynamics:")
}
