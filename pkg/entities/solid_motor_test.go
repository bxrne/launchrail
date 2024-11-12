package entities_test

import (
	"testing"
	"time"

	"github.com/bxrne/launchrail/pkg/entities"
)

func TestNewSolidMotor(t *testing.T) {
	motor, err := entities.NewSolidMotor("../../testdata/cesaroni-l645.eng")
	if err != nil {
		t.Fatalf("Failed to create SolidMotor: %v", err)
	}

	if motor.Manufacturer != "CTI" {
		t.Errorf("Expected Manufacturer 'CTI', got '%s'", motor.Manufacturer)
	}

	if motor.Designation != "3419-L645-GR-P" {
		t.Errorf("Expected Designation '3419-L645-GR-P', got '%s'", motor.Designation)
	}

	if motor.Diameter != 75 {
		t.Errorf("Expected Diameter 75, got %f", motor.Diameter)
	}

	if motor.Length != 486 {
		t.Errorf("Expected Length 486, got %f", motor.Length)
	}

	if motor.TotalImpulse != 2.1441 {
		t.Errorf("Expected TotalImpulse 2.1441, got %f", motor.TotalImpulse)
	}

	if motor.Propellant != "P" {
		t.Errorf("Expected Propellant 'P', got '%s'", motor.Propellant)
	}

	if motor.AverageThrust != 3.7518 {
		t.Errorf("Expected AverageThrust 3.7518, got %f", motor.AverageThrust)
	}

	if motor.BurnTime != 5*time.Second+390*time.Millisecond {
		t.Errorf("Expected BurnTime 5.39s, got %v", motor.BurnTime)
	}
}
