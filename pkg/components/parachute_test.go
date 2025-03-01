package components_test

import (
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/types"
)

// TEST: GIVEN a Parachute struct WHEN calling the String method THEN return a string representation of the Parachute struct
func TestParachuteString(t *testing.T) {
	p := &components.Parachute{
		ID:              ecs.NewBasic(),
		Position:        types.Vector3{X: 0, Y: 0, Z: 0},
		Diameter:        1.0,
		DragCoefficient: 1.0,
		Strands:         1,
		Area:            0.25 * 3.14159 * 1.0 * 1.0,
	}

	expected := "Parachute{ID={10 <nil> []}, Position=Vector3{X: 0.00, Y: 0.00, Z: 0.00}, Diameter=1.00, DragCoefficient=1.00, Strands=1, Area=0.79}"
	if p.String() != expected {
		t.Errorf("Expected %s, got %s", expected, p.String())
	}
}

// TEST: GIVEN a diameter, drag coefficient, strands, and trigger WHEN calling NewParachute THEN return a new Parachute instance
func TestNewParachute(t *testing.T) {
	p := components.NewParachute(ecs.NewBasic(), 1.0, 1.0, 1, components.ParachuteTriggerNone)

	if p.Diameter != 1.0 {
		t.Errorf("Expected 1.0, got %f", p.Diameter)
	}
	if p.DragCoefficient != 1.0 {
		t.Errorf("Expected 1.0, got %f", p.DragCoefficient)
	}
	if p.Strands != 1 {
		t.Errorf("Expected 1, got %d", p.Strands)
	}
	if p.Trigger != components.ParachuteTriggerNone {
		t.Errorf("Expected ParachuteTriggerNone, got %s", p.Trigger)
	}
}

// TEST: GIVEN a parachute WHEN Update is called nil is returned as the Error
func TestUpdate(t *testing.T) {
	p := &components.Parachute{}
	if p.Update(0) != nil {
		t.Error("Expected nil, got an error")
	}
}

// TEST: GIVEN a parachute WHEN Type is called the correct type is returned
func TestType(t *testing.T) {
	p := &components.Parachute{}
	if p.Type() != "Parachute" {
		t.Errorf("Expected Parachute, got %s", p.Type())
	}
}

// TEST: GIVEN a Parachute WHEN GetPlanformArea is called THEN return the planform Area
func TestGetPlanformArea(t *testing.T) {
	p := &components.Parachute{
		Diameter: 1.0,
	}

	if p.GetPlanformArea() != 0.00 {
		t.Errorf("Expected 0.79, got %f", p.GetPlanformArea())
	}
}

// TEST: GIVEN a parachute WHEN GetMass is called THEN return the mass of the Parachute
func TestGetMass(t *testing.T) {
	p := &components.Parachute{}
	if p.GetMass() != 0.0 {
		t.Errorf("Expected 0.0, got %f", p.GetMass())
	}
}

// TEST: GIVEN a parachute WHEN calling GetDensity THEN return the density of the parachute
func TestGetDensity(t *testing.T) {
	p := &components.Parachute{}
	if p.GetDensity() != 0.0 {
		t.Errorf("Expected 0.0, got %f", p.GetDensity())
	}
}

// TEST: GIVEN a parachute WHEN calling GetVolume THEN return the volume of the parachute
func TestGetVolume(t *testing.T) {
	p := &components.Parachute{}
	if p.GetVolume() != 0.0 {
		t.Errorf("Expected 0.0, got %f", p.GetVolume())
	}
}

// TEST: GIVEN a parachute WHEN calling GetSurfaceArea THEN return the surface area of the parachute
func TestGetSurfaceArea(t *testing.T) {
	p := &components.Parachute{}
	if p.GetSurfaceArea() != 0.0 {
		t.Errorf("Expected 0.0, got %f", p.GetSurfaceArea())
	}
}

// TEST: GIVEN a parachute WHEN calling IsDeployed THEN return whether the parachute is IsDeployed
func TestIsDeployed(t *testing.T) {
	p := &components.Parachute{}
	if p.IsDeployed() != false {
		t.Errorf("Expected false, got %t", p.IsDeployed())
	}
}

// TEST: GIVEN a parachute WHEN calling Deploy THEN set the parachute to deployed
func TestDeploy(t *testing.T) {
	p := &components.Parachute{}
	p.Deploy()
	if p.IsDeployed() != true {
		t.Errorf("Expected true, got %t", p.IsDeployed())
	}
}
