package entities_test

import (
	"os"
	"testing"
	"time"

	"github.com/bxrne/launchrail/pkg/entities"

	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN a valid motor file WHEN NewSolidMotor is called THEN it should parse the motor data correctly
func TestNewSolidMotor_ValidFile(t *testing.T) {
	motorFileContent := `
; This is a comment
M1234 54 100 APCP 500 250 ManufacturerX
0.0 0
0.5 100
1.0 200
1.5 150
2.0 0
`
	tmpFile, err := os.CreateTemp("", "motor_test_*.eng")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(motorFileContent))
	assert.NoError(t, err)
	tmpFile.Close()

	motor, err := entities.NewSolidMotor(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, motor)

	assert.Equal(t, "M1234", motor.Designation)
	assert.Equal(t, 54.0, motor.Diameter)
	assert.Equal(t, 100.0, motor.Length)
	assert.Equal(t, "APCP", motor.Propellant)
	assert.Equal(t, 500.0, motor.TotalImpulse)
	assert.Equal(t, 250.0, motor.AverageThrust)
	assert.Equal(t, "ManufacturerX", motor.Manufacturer)
	assert.Equal(t, 2*time.Second, motor.BurnTime)
	assert.Len(t, motor.ThrustCurve, 5)
}

// TEST: GIVEN a non-existent motor file WHEN NewSolidMotor is called THEN it should return an error
func TestNewSolidMotor_FileNotFound(t *testing.T) {
	_, err := entities.NewSolidMotor("non_existent_file.eng")
	assert.Error(t, err)
}

// TEST: GIVEN a motor file with missing data WHEN NewSolidMotor is called THEN it should handle the missing data gracefully
func TestNewSolidMotor_MissingData(t *testing.T) {
	motorFileContent := `
; This is a comment
M1234 54 100 APCP 500 250 ManufacturerX
`
	tmpFile, err := os.CreateTemp("", "motor_test_missing_data_*.eng")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(motorFileContent))
	assert.NoError(t, err)
	tmpFile.Close()

	motor, err := entities.NewSolidMotor(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, motor)

	assert.Equal(t, "M1234", motor.Designation)
	assert.Equal(t, 54.0, motor.Diameter)
	assert.Equal(t, 100.0, motor.Length)
	assert.Equal(t, "APCP", motor.Propellant)
	assert.Equal(t, 500.0, motor.TotalImpulse)
	assert.Equal(t, 250.0, motor.AverageThrust)
	assert.Equal(t, "ManufacturerX", motor.Manufacturer)
	assert.Equal(t, 0*time.Second, motor.BurnTime)
	assert.Len(t, motor.ThrustCurve, 0)
}
