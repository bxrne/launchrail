package components_test

import (
	"testing"
	"time"

	"github.com/bxrne/launchrail/pkg/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestMotor(t *testing.T) *components.SolidMotor {
	motorFilePath := "../../testdata/cesaroni-l645.eng"
	motor, err := components.NewSolidMotor(motorFilePath, 0.5, 0.25, 3)
	require.NoError(t, err)
	return motor
}

// TEST: GIVEN a valid motor configuration WHEN creating a new solid motor THEN initialize with correct parameters
func TestNewSolidMotor_ValidMotor(t *testing.T) {
	motor := createTestMotor(t)

	assert.Equal(t, "3419-L645-GR-P", motor.Designation)
	assert.Equal(t, 75.0, motor.Diameter)
	assert.Equal(t, 3, len(motor.Grains))

	grainLength := motor.Length / float64(len(motor.Grains))
	for _, grain := range motor.Grains {
		assert.Equal(t, grainLength, grain.InitialLength)
		assert.Equal(t, grainLength, grain.CurrentLength)
		assert.Equal(t, motor.Diameter, grain.Diameter)
	}

	assert.Equal(t, 0*time.Second, motor.CurrentState.ElapsedTime)
	assert.Equal(t, 0.0, motor.CurrentState.CurrentThrust)
	assert.Equal(t, 0.25, motor.CurrentState.RemainingMass)
}

// TEST: GIVEN an initialized motor WHEN updating state with multiple time steps THEN track motor progression
func TestMotorUpdate(t *testing.T) {
	motor := createTestMotor(t)

	timeSteps := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		500 * time.Millisecond,
	}

	for _, step := range timeSteps {
		err := motor.Update(step)
		assert.NoError(t, err)
	}

	assert.True(t, motor.CurrentState.ElapsedTime > 0)
	assert.True(t, motor.CurrentState.RemainingMass < 0.25)
	assert.True(t, motor.CurrentState.BurnedPropellant > 0)
}

// TEST: GIVEN an attempt to update motor state WHEN using invalid time steps THEN return appropriate errors
func TestMotorInvalidTimeStep(t *testing.T) {
	motor := createTestMotor(t)

	err := motor.Update(0)
	assert.Error(t, err, "Zero time step should return an error")

	err = motor.Update(-1 * time.Second)
	assert.Error(t, err, "Negative time step should return an error")
}

// TEST: GIVEN an initialized motor WHEN burning occurs THEN verify grain length reduction
func TestMotorGrainBurning(t *testing.T) {
	motor := createTestMotor(t)

	initialLengths := make([]float64, len(motor.Grains))
	for i, grain := range motor.Grains {
		initialLengths[i] = grain.CurrentLength
	}

	err := motor.Update(1 * time.Second)
	assert.NoError(t, err)

	for i, grain := range motor.Grains {
		assert.Less(t, grain.CurrentLength, initialLengths[i],
			"Grain %d length should decrease after burning", i)
	}
}
