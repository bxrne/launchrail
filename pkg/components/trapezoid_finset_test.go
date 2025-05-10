package components_test

import (
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestTrapezoidFinset_GetMass(t *testing.T) {
	// Arrange
	finset := &components.TrapezoidFinset{
		BasicEntity: ecs.NewBasic(),
		Mass:        0.5,
	}

	// Act
	mass := finset.GetMass()

	// Assert
	assert.Equal(t, 0.5, mass)
}

func TestGetTrapezoidPlanformArea(t *testing.T) {
	testCases := []struct {
		name     string
		fin      *components.TrapezoidFinset
		expected float64
	}{
		{
			name: "standard dimensions",
			fin: &components.TrapezoidFinset{
				BasicEntity: ecs.NewBasic(),
				RootChord:   0.1,
				TipChord:    0.05,
				Span:        0.15,
			},
			expected: 0.01125, // (0.1 + 0.05) * 0.15 / 2
		},
		{
			name: "zero dimensions",
			fin: &components.TrapezoidFinset{
				BasicEntity: ecs.NewBasic(),
				RootChord:   0,
				TipChord:    0,
				Span:        0,
			},
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := tc.fin.GetPlanformArea()

			// Assert
			assert.InDelta(t, tc.expected, result, 0.0001)
		})
	}
}

func TestNewTrapezoidFinsetFromORK(t *testing.T) {
	// Arrange
	basic := ecs.NewBasic()
	ork := &openrocket.RocketDocument{
		Subcomponents: openrocket.Subcomponents{
			Stages: []openrocket.RocketStage{
				{
					SustainerSubcomponents: openrocket.SustainerSubcomponents{
						BodyTube: openrocket.BodyTube{
							Subcomponents: openrocket.BodyTubeSubcomponents{
								TrapezoidFinsets: []openrocket.TrapezoidFinset{
									{
										Name:     "TestFinset",
										FinCount: 4,
										RootChord: 0.1,
										TipChord:  0.05,
										Height:    0.07,
										SweepLength: 0.02, // Assuming this is sweep distance
										Thickness: 0.003,
										Material: openrocket.Material{
											Type:    "bulk",
											Density: 1250,
											Name:    "PLA",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Act
	finset := components.NewTrapezoidFinsetFromORK(basic, ork)

	// Assert
	assert.NotNil(t, finset)
	assert.Equal(t, 0.1, finset.RootChord)
	assert.Equal(t, 0.05, finset.TipChord)
	assert.Equal(t, 0.15, finset.Span)
	assert.Equal(t, 0.05, finset.SweepAngle)
	assert.Equal(t, types.Vector3{X: 1.0, Y: 0, Z: 0}, finset.Position.Vec)
}
