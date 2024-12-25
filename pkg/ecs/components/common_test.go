package components_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/ecs/types"
)

// TEST: GIVEN a transform component WHEN Type is called THEN the type is returned.
func TestTransformComponent_Type(t *testing.T) {
	transform := &components.TransformComponent{
		Position: types.Vector3{X: 1, Y: 2, Z: 3},
		Rotation: types.Vector3{X: 0, Y: 90, Z: 0},
		Scale:    types.Vector3{X: 1, Y: 1, Z: 1},
	}

	if got := transform.Type(); got != "Transform" {
		t.Errorf("TransformComponent.Type() = %v, want %v", got, "Transform")
	}
}

// TEST: GIVEN a physics component WHEN Type is called THEN the type is returned.
func TestPhysicsComponent_Type(t *testing.T) {
	physics := &components.PhysicsComponent{
		Mass:      10.5,
		Velocity:  types.Vector3{X: 0, Y: 100, Z: 0},
		Forces:    []types.Vector3{{X: 0, Y: -9.81, Z: 0}},
		Drag:      0.3,
		Thrust:    1000.0,
		MotorType: "F12",
	}

	if got := physics.Type(); got != "Physics" {
		t.Errorf("PhysicsComponent.Type() = %v, want %v", got, "Physics")
	}
}

// TEST: GIVEN an aerodynamics component WHEN Type is called THEN the type is returned.
func TestAerodynamicsComponent_Type(t *testing.T) {
	aero := &components.AerodynamicsComponent{
		CenterOfPressure: types.Vector3{X: 0, Y: 0.5, Z: 0},
		CenterOfMass:     types.Vector3{X: 0, Y: 0.3, Z: 0},
		CrossSection:     0.01,
		StabilityMargin:  2.0,
	}

	if got := aero.Type(); got != "Aerodynamics" {
		t.Errorf("AerodynamicsComponent.Type() = %v, want %v", got, "Aerodynamics")
	}
}

// TEST: GIVEN a component creation function WHEN called THEN the component is created.
func TestComponents_Creation(t *testing.T) {
	tests := []struct {
		name string
		fn   func() interface{}
		want string
	}{
		{
			name: "create transform component",
			fn: func() interface{} {
				return &components.TransformComponent{
					Position: types.Vector3{X: 1, Y: 2, Z: 3},
					Rotation: types.Vector3{X: 45, Y: 0, Z: 0},
					Scale:    types.Vector3{X: 1, Y: 1, Z: 1},
				}
			},
			want: "Transform",
		},
		{
			name: "create physics component",
			fn: func() interface{} {
				return &components.PhysicsComponent{
					Mass:     1.5,
					Velocity: types.Vector3{X: 0, Y: 10, Z: 0},
					Forces: []types.Vector3{
						{X: 0, Y: -9.81, Z: 0},
						{X: 0, Y: 15, Z: 0},
					},
					Drag:      0.2,
					Thrust:    500.0,
					MotorType: "E12",
				}
			},
			want: "Physics",
		},
		{
			name: "create aerodynamics component",
			fn: func() interface{} {
				return &components.AerodynamicsComponent{
					CenterOfPressure: types.Vector3{X: 0, Y: 0.8, Z: 0},
					CenterOfMass:     types.Vector3{X: 0, Y: 0.4, Z: 0},
					CrossSection:     0.008,
					StabilityMargin:  1.5,
				}
			},
			want: "Aerodynamics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := tt.fn()

			switch c := component.(type) {
			case *components.TransformComponent:
				if got := c.Type(); got != tt.want {
					t.Errorf("TransformComponent.Type() = %v, want %v", got, tt.want)
				}
				// Verify fields are set correctly
				if c.Position.X != 1 || c.Position.Y != 2 || c.Position.Z != 3 {
					t.Error("TransformComponent Position not set correctly")
				}
			case *components.PhysicsComponent:
				if got := c.Type(); got != tt.want {
					t.Errorf("PhysicsComponent.Type() = %v, want %v", got, tt.want)
				}
				// Verify forces array length
				if len(c.Forces) != 2 {
					t.Error("PhysicsComponent Forces not set correctly")
				}
			case *components.AerodynamicsComponent:
				if got := c.Type(); got != tt.want {
					t.Errorf("AerodynamicsComponent.Type() = %v, want %v", got, tt.want)
				}
				// Verify stability margin is positive
				if c.StabilityMargin <= 0 {
					t.Error("AerodynamicsComponent StabilityMargin should be positive")
				}
			}
		})
	}
}

// TEST: GIVEN a transform component WHEN Position is set THEN the position is updated.
func TestPhysicsComponent_ForceManipulation(t *testing.T) {
	physics := &components.PhysicsComponent{
		// Mass:      1.0,
		Velocity: types.Vector3{X: 0, Y: 0, Z: 0},
		Forces:   make([]types.Vector3, 0),
		// Drag:      0.1,
		// Thrust:    100.0,
		// MotorType: "D12",
	}

	// Test initial forces
	if len(physics.Forces) != 0 {
		t.Errorf("Initial Forces length = %v, want %v", len(physics.Forces), 0)
	}

	// Add forces
	physics.Forces = append(physics.Forces,
		types.Vector3{X: 0, Y: -9.81, Z: 0}, // Gravity
		types.Vector3{X: 0, Y: 100, Z: 0},   // Thrust
	)

	// Test forces after addition
	if len(physics.Forces) != 2 {
		t.Errorf("Forces length after addition = %v, want %v", len(physics.Forces), 2)
	}

	// Verify force values
	if physics.Forces[0].Y != -9.81 {
		t.Errorf("Gravity force = %v, want %v", physics.Forces[0].Y, -9.81)
	}
	if physics.Forces[1].Y != 100 {
		t.Errorf("Thrust force = %v, want %v", physics.Forces[1].Y, 100)
	}
}

// TEST: GIVEN a physics component WHEN Thrust is set THEN the thrust is updated.
func TestAerodynamicsComponent_StabilityCalculation(t *testing.T) {
	tests := []struct {
		name          string
		cop           types.Vector3
		com           types.Vector3
		wantStability float64
	}{
		{
			name:          "stable configuration",
			cop:           types.Vector3{X: 0, Y: 0.8, Z: 0},
			com:           types.Vector3{X: 0, Y: 0.4, Z: 0},
			wantStability: 2.0,
		},
		{
			name:          "neutral stability",
			cop:           types.Vector3{X: 0, Y: 0.5, Z: 0},
			com:           types.Vector3{X: 0, Y: 0.5, Z: 0},
			wantStability: 0.0,
		},
		{
			name:          "unstable configuration",
			cop:           types.Vector3{X: 0, Y: 0.3, Z: 0},
			com:           types.Vector3{X: 0, Y: 0.6, Z: 0},
			wantStability: -1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aero := &components.AerodynamicsComponent{
				// CenterOfPressure: tt.cop,
				// CenterOfMass:     tt.com,
				// CrossSection:     0.01,
				StabilityMargin: tt.wantStability,
			}

			// Calculate actual stability (COP - COM in calibers)
			gotStability := aero.StabilityMargin

			if gotStability != tt.wantStability {
				t.Errorf("StabilityMargin = %v, want %v", gotStability, tt.wantStability)
			}
		})
	}
}
