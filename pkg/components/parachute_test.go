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
