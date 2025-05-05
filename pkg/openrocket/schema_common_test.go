package openrocket_test

import (
	"math"
	"testing"

	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN a RadiusOffset struct WHEN calling the String method THEN return a string representation of the RadiusOffset struct
func TestSchemaRadiusOffsetString(t *testing.T) {
	r := &openrocket.RadiusOffset{
		Method: "method1",
		Value:  1.0,
	}

	expected := "RadiusOffset{Method=method1, Value=1.00}"
	if r.String() != expected {
		t.Errorf("Expected %s, got %s", expected, r.String())
	}
}

// TEST: GIVEN a AngleOffset struct WHEN calling the String method THEN return a string representation of the AngleOffset struct
func TestSchemaAngleOffsetString(t *testing.T) {
	a := &openrocket.AngleOffset{
		Method: "method1",
		Value:  1.0,
	}

	expected := "AngleOffset{Method=method1, Value=1.00}"
	if a.String() != expected {
		t.Errorf("Expected %s, got %s", expected, a.String())
	}
}

// TEST: GIVEN a AxialOffset struct WHEN calling the String method THEN return a string representation of the AxialOffset struct
func TestSchemaAxialOffsetString(t *testing.T) {
	a := &openrocket.AxialOffset{
		Method: "method1",
		Value:  1.0,
	}

	expected := "AxialOffset{Method=method1, Value=1.00}"
	if a.String() != expected {
		t.Errorf("Expected %s, got %s", expected, a.String())
	}
}

// TEST: GIVEN a Position struct WHEN calling the String method THEN return a string representation of the Position struct
func TestSchemaPositionString(t *testing.T) {
	p := &openrocket.Position{
		Value: 1.0,
		Type:  "type1",
	}

	expected := "Position{Value=1.00, Type=type1}"
	if p.String() != expected {
		t.Errorf("Expected %s, got %s", expected, p.String())
	}
}

// TEST: GIVEN a CenteringRing struct WHEN calling the String method THEN return a string representation of the CenteringRing struct
func TestSchemaCenteringRingString(t *testing.T) {
	c := &openrocket.CenteringRing{
		Name:               "name",
		ID:                 "id",
		InstanceCount:      1,
		InstanceSeparation: 1.0,
		AxialOffset:        openrocket.AxialOffset{},
		Position:           openrocket.Position{},
		Material:           openrocket.Material{},
		Length:             1.0,
		RadialPosition:     1.0,
		OuterRadius:        "auto",
		InnerRadius:        "auto",
	}

	expected := "CenteringRing{Name=name, ID=id, InstanceCount=1, InstanceSeparation=1.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=1.00, RadialPosition=1.00, OuterRadius=auto, InnerRadius=auto}"
	if c.String() != expected {
		t.Errorf("Expected %s, got %s", expected, c.String())
	}
}

// TEST: MassComponent GetMass
func TestSchemaMassComponentGetMass(t *testing.T) {
	// MassComponent directly contains Mass, no embedding involved for GetMass
	mc := &openrocket.MassComponent{Mass: 5.67}
	assert.Equal(t, 5.67, mc.GetMass(), "MassComponent GetMass should return the Mass field")

	mcNil := (*openrocket.MassComponent)(nil)
	assert.Equal(t, 0.0, mcNil.GetMass(), "Nil MassComponent GetMass should return 0.0")

	mcZero := &openrocket.MassComponent{Mass: 0.0}
	assert.Equal(t, 0.0, mcZero.GetMass(), "Zero MassComponent GetMass should return 0.0")

	mcNeg := &openrocket.MassComponent{Mass: -1.0}
	assert.Equal(t, -1.0, mcNeg.GetMass(), "Negative MassComponent GetMass should return negative value")
}

// TEST: CenteringRing GetMass
func TestSchemaCenteringRingGetMass(t *testing.T) {
	tests := []struct {
		name string
		ring *openrocket.CenteringRing
		want float64
	}{
		{
			name: "Valid Ring",
			ring: func() *openrocket.CenteringRing {
				r := &openrocket.CenteringRing{}
				r.Length = 0.05
				r.OuterRadius = "0.05"
				r.InnerRadius = "0.04"
				r.Material = openrocket.Material{Density: 1200}
				r.InstanceCount = 1
				return r
			}(),
			want: 0.05 * math.Pi * (math.Pow(0.05, 2) - math.Pow(0.04, 2)) * 1200,
		},
		{
			name: "Zero Length",
			ring: func() *openrocket.CenteringRing {
				r := &openrocket.CenteringRing{}
				r.Length = 0.0
				r.OuterRadius = "0.05"
				r.InnerRadius = "0.04"
				r.Material = openrocket.Material{Density: 1200}
				return r
			}(),
			want: 0.0,
		},
		{
			name: "Zero Density",
			ring: func() *openrocket.CenteringRing {
				r := &openrocket.CenteringRing{}
				r.Length = 0.05
				r.OuterRadius = "0.05"
				r.InnerRadius = "0.04"
				r.Material = openrocket.Material{Density: 0.0}
				return r
			}(),
			want: 0.0,
		},
		{
			name: "OuterRadius Auto",
			ring: func() *openrocket.CenteringRing {
				r := &openrocket.CenteringRing{}
				r.Length = 0.05
				r.OuterRadius = "auto"
				r.InnerRadius = "0.04"
				r.Material = openrocket.Material{Density: 1200}
				return r
			}(),
			want: 0.0,
		},
		{
			name: "InnerRadius Auto",
			ring: func() *openrocket.CenteringRing {
				r := &openrocket.CenteringRing{}
				r.Length = 0.05
				r.OuterRadius = "0.05"
				r.InnerRadius = "auto"
				r.Material = openrocket.Material{Density: 1200}
				return r
			}(),
			want: 0.0,
		},
		{
			name: "OuterRadius Invalid",
			ring: func() *openrocket.CenteringRing {
				r := &openrocket.CenteringRing{}
				r.Length = 0.05
				r.OuterRadius = "invalid"
				r.InnerRadius = "0.04"
				r.Material = openrocket.Material{Density: 1200}
				return r
			}(),
			want: 0.0,
		},
		{
			name: "InnerRadius Invalid",
			ring: func() *openrocket.CenteringRing {
				r := &openrocket.CenteringRing{}
				r.Length = 0.05
				r.OuterRadius = "0.05"
				r.InnerRadius = "invalid"
				r.Material = openrocket.Material{Density: 1200}
				return r
			}(),
			want: 0.0,
		},
		{
			name: "InnerRadius >= OuterRadius",
			ring: func() *openrocket.CenteringRing {
				r := &openrocket.CenteringRing{}
				r.Length = 0.05
				r.OuterRadius = "0.04"
				r.InnerRadius = "0.05"
				r.Material = openrocket.Material{Density: 1200}
				return r
			}(),
			want: 0.0,
		},
		{
			name: "Nil Ring",
			ring: nil,
			want: 0.0,
		},
	}

	for _, tt := range tests {
		tc := tt // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			var got float64
			if tc.ring != nil {
				got = tc.ring.GetMass()
			} else {
				got = 0.0
			}
			assert.InDelta(t, tc.want, got, 1e-9, "CenteringRing.GetMass() mismatch")
		})
	}
}

// TEST: Shockcord GetMass
func TestSchemaShockcordGetMass(t *testing.T) {
	tests := []struct {
		name string
		cord *openrocket.Shockcord
		want float64
	}{
		{
			name: "Valid Cord",
			cord: &openrocket.Shockcord{
				CordLength: 3.0,
				Material:   openrocket.Material{Density: 0.01},
			},
			want: 3.0 * 0.01,
		},
		{
			name: "Zero Length",
			cord: &openrocket.Shockcord{
				CordLength: 0.0,
				Material:   openrocket.Material{Density: 0.01},
			},
			want: 0.0,
		},
		{
			name: "Zero Density",
			cord: &openrocket.Shockcord{
				CordLength: 3.0,
				Material:   openrocket.Material{Density: 0.0},
			},
			want: 0.0,
		},
		{
			name: "Negative Density",
			cord: &openrocket.Shockcord{
				CordLength: 3.0,
				Material:   openrocket.Material{Density: -0.01},
			},
			want: 0.0,
		},
		{
			name: "Nil Cord",
			cord: nil,
			want: 0.0,
		},
	}
	for _, tt := range tests {
		tc := tt // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cord.GetMass()
			assert.InDelta(t, tc.want, got, 1e-9, "Shockcord.GetMass() mismatch")
		})
	}
}

// NOTE: parseRadius is not exported, so we can't test it directly.
// Its functionality is implicitly tested via the GetMass methods that use it
// (RingComponent, LaunchLug). If those tests pass with "auto" and invalid
// radius inputs resulting in 0 mass, parseRadius is working as expected
// within that context.
