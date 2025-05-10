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
										// Mock the Position field correctly
										Position: openrocket.Position{Value: 1.0},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Extract the specific OpenRocket finset data needed for the constructor
	assert.True(t, len(ork.Subcomponents.Stages) > 0, "No stages in ORK data")
	assert.True(t, len(ork.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets) > 0, "No trapezoid finsets in ORK data")
	orkSpecificFinset := ork.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.TrapezoidFinsets[0]
	
	finsetPosition := types.Vector3{X: orkSpecificFinset.Position.Value, Y: 0, Z: 0}
	finsetMaterial := orkSpecificFinset.Material

	// Act
	finset, err := components.NewTrapezoidFinsetFromORK(&orkSpecificFinset, finsetPosition, finsetMaterial)

	// Assert
	assert.NoError(t, err) // Check for errors during creation
	assert.NotNil(t, finset)
	assert.Equal(t, 0.1, finset.RootChord)
	assert.Equal(t, 0.05, finset.TipChord)
	assert.Equal(t, 0.07, finset.Span)
	assert.Equal(t, 0.02, finset.SweepDistance)
	assert.Equal(t, types.Vector3{X: 1.0, Y: 0, Z: 0}, finset.Position) // Assert component's attachment position

	// Assert Mass
	// Expected PlanformArea = (0.1 + 0.05) * 0.07 / 2 = 0.00525
	// Expected SingleFinVolume = 0.00525 * 0.003 = 0.00001575
	// Expected SingleFinMass = 0.00001575 * 1250 = 0.0196875
	// Expected TotalMass = 0.0196875 * 4 = 0.07875
	expectedMass := 0.07875
	assert.InDelta(t, expectedMass, finset.GetMass(), 1e-6, "Calculated mass does not match expected")

	// Calculate expected CenterOfMass.X
	// xCgLocalNum := (SweepDistance * (RootChord + 2*TipChord)) + (RootChord^2 + RootChord*TipChord + TipChord^2)
	// xCgLocalDen := 3 * (RootChord + TipChord)
	// xCgLocal := xCgLocalNum / xCgLocalDen
	// Expected x_cg_local = (0.02 * (0.1 + 2*0.05) + (0.1*0.1 + 0.1*0.05 + 0.05*0.05)) / (3 * (0.1 + 0.05))
	// Expected x_cg_local = (0.02 * 0.2 + (0.01 + 0.005 + 0.0025)) / (3 * 0.15)
	// Expected x_cg_local = (0.004 + 0.0175) / 0.45
	// Expected x_cg_local = 0.0215 / 0.45 = 0.04777777777
	expectedXcgLocal := 0.0215 / 0.45
	expectedCM := types.Vector3{
		X: finset.Position.X + expectedXcgLocal,
		Y: 0,
		Z: 0,
	}
	assert.InDelta(t, expectedCM.X, finset.CenterOfMass.X, 1e-6, "CenterOfMass.X does not match expected")
	assert.InDelta(t, expectedCM.Y, finset.CenterOfMass.Y, 1e-6, "CenterOfMass.Y does not match expected")
	assert.InDelta(t, expectedCM.Z, finset.CenterOfMass.Z, 1e-6, "CenterOfMass.Z does not match expected")
}
