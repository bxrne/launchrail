package components_test

// func TestNewTrapezoidFinsetFromORK(t *testing.T) {
// 	basicEntity := ecs.BasicEntity{}

// 	// Mock OpenRocket data
// 	ork := &openrocket.RocketDocument{
// 		Subcomponents: openrocket.RocketSubcomponents{
// 			Stages: []openrocket.Stage{
// 				{
// 					SustainerSubcomponents: openrocket.SustainerSubcomponents{
// 						BodyTube: openrocket.BodyTube{
// 							Subcomponents: openrocket.BodyTubeSubcomponents{
// 								TrapezoidFinset: openrocket.TrapezoidFinset{
// 									RootChord:   12.0,
// 									TipChord:    6.0,
// 									Height:      5.0,
// 									SweepLength: 3.0,
// 									Thickness:   0.15,
// 									Material:    openrocket.Material{Density: 2.0},
// 									AxialOffset: openrocket.Value{Value: 1.0},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	finset := NewTrapezoidFinsetFromORK(basicEntity, ork)

// 	assert.NotNil(t, finset)
// 	assert.Equal(t, 12.0, finset.RootChord, "RootChord should match OpenRocket data")
// 	assert.Equal(t, 6.0, finset.TipChord, "TipChord should match OpenRocket data")
// 	assert.Equal(t, 5.0, finset.Span, "Span should match OpenRocket data")
// 	assert.Equal(t, 3.0, finset.SweepAngle, "SweepAngle should match OpenRocket data")
// 	assert.Equal(t, 1.0, finset.Position.X, "Position.X should match OpenRocket data")
// }

// func TestGetPlanformArea(t *testing.T) {
// 	finset := &TrapezoidFinset{
// 		RootChord: 10.0,
// 		TipChord:  5.0,
// 		Span:      4.0,
// 	}

// 	expectedArea := (10.0 + 5.0) * 4.0 / 2
// 	assert.Equal(t, expectedArea, finset.GetPlanformArea(), "Planform area calculation should be correct")
// }
