package entities_test

import (
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/bxrne/launchrail/pkg/entities"
	openrocket "github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockOpenRocketData() *openrocket.RocketDocument {
	return &openrocket.RocketDocument{
		Subcomponents: openrocket.Subcomponents{
			Stages: []openrocket.RocketStage{
				{
					SustainerSubcomponents: openrocket.SustainerSubcomponents{
						Nosecone: openrocket.Nosecone{
							Material: openrocket.Material{
								Density: 1.0,
								Name:    "Test Material",
								Type:    "BULK",
							},
							Length:         1.0,
							Thickness:      0.1,
							AftRadius:      0.5,
							ShapeParameter: 0.5,
							Shape:          "OGIVE",
							Subcomponents:  openrocket.NoseSubcomponents{},
						},
						BodyTube: openrocket.BodyTube{
							Material: openrocket.Material{
								Density: 1.0,
								Name:    "Test Material",
								Type:    "BULK",
							},
							Length:    2.0,
							Thickness: 0.1,
							Radius:    "0.5",
							Subcomponents: openrocket.BodyTubeSubcomponents{
								Parachute: openrocket.Parachute{
									CD: "auto 1.0",
								},
								TrapezoidFinset: openrocket.TrapezoidFinset{
									Material: openrocket.Material{
										Density: 1.0,
										Name:    "Test Material",
										Type:    "BULK",
									},
									RootChord: 0.2,
									TipChord:  0.1,
									Height:    0.15,
									Thickness: 0.003,
									FinCount:  4,
								},
							},
						},
					},
				},
			},
		},
	}
}

func createMockMotor() *components.Motor {
	return &components.Motor{
		ID:   ecs.NewBasic(),
		Mass: 1.0, // Set a valid initial mass
		Props: &thrustcurves.MotorData{
			Designation: designation.Designation("MockMotor-A1-P"),
		},
	}
}

// TEST: GIVEN valid OpenRocket data WHEN NewRocketEntity is called THEN a new rocket entity is created
func TestNewRocketEntity(t *testing.T) {
	// Arrange
	world := &ecs.World{}
	orkData := createMockOpenRocketData()
	motor := createMockMotor()

	// Act
	rocket := entities.NewRocketEntity(world, orkData, motor)

	// Assert
	assert.NotNil(t, rocket)
	assert.NotNil(t, rocket.Position)
	assert.NotNil(t, rocket.Velocity)
	assert.NotNil(t, rocket.Acceleration)
	assert.NotNil(t, rocket.Mass)
}

// TEST: GIVEN a rocket entity WHEN GetComponent is called with valid component name THEN the component is returned
func TestGetComponent(t *testing.T) {
	// Arrange
	world := &ecs.World{}
	orkData := createMockOpenRocketData()
	motor := createMockMotor()
	rocket := entities.NewRocketEntity(world, orkData, motor)

	tests := []struct {
		name      string
		component string
		wantNil   bool
	}{
		{"motor component", "motor", false},
		{"bodytube component", "bodytube", false},
		{"nosecone component", "nosecone", false},
		{"finset component", "finset", false},
		{"non-existent component", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			component := rocket.GetComponent(tt.component)

			// Assert
			if tt.wantNil {
				assert.Nil(t, component)
			} else {
				assert.NotNil(t, component)
			}
		})
	}
}

// TEST: GIVEN invalid OpenRocket data WHEN NewRocketEntity is called THEN nil is returned
func TestNewRocketEntityWithInvalidData(t *testing.T) {
	// Arrange
	world := &ecs.World{}
	invalidOrkData := &openrocket.RocketDocument{
		Subcomponents: openrocket.Subcomponents{
			Stages: []openrocket.RocketStage{
				{
					SustainerSubcomponents: openrocket.SustainerSubcomponents{
						BodyTube: openrocket.BodyTube{
							Radius: "invalid", // This will cause an error
						},
					},
				},
			},
		},
	}
	motor := createMockMotor()

	// Act
	rocket := entities.NewRocketEntity(world, invalidOrkData, motor)

	// Assert
	assert.Nil(t, rocket)
}

// TEST: GIVEN a rocket entity with multiple components WHEN GetComponent is called concurrently THEN no race conditions occur
func TestGetComponentConcurrency(t *testing.T) {
	// Arrange
	world := &ecs.World{}
	orkData := createMockOpenRocketData()
	motor := createMockMotor()
	rocket := entities.NewRocketEntity(world, orkData, motor)

	// Act & Assert
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			// Access components concurrently
			_ = rocket.GetComponent("motor")
			_ = rocket.GetComponent("bodytube")
			_ = rocket.GetComponent("nosecone")
			_ = rocket.GetComponent("finset")
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestNewRocketEntity_NilData(t *testing.T) {
	motor := &components.Motor{} // Dummy motor
	rocket := entities.NewRocketEntity(nil, nil, motor)
	assert.Nil(t, rocket, "Should return nil if ORK data is nil")

	rocket = entities.NewRocketEntity(nil, &openrocket.RocketDocument{}, nil)
	assert.Nil(t, rocket, "Should return nil if motor is nil")
}

func TestNewRocketEntity_ComponentErrors(t *testing.T) {
	world := &ecs.World{}
	// motor := &components.Motor{Props: &thrustcurves.MotorData{TotalMass: 0.1}} // Unused Basic valid motor

	// Test case 1: Bodytube creation error (e.g., invalid ORK data)
	// This requires ORK data that causes NewBodytubeFromORK to error.
	// We might need a specific test ORK file or mock the function.
	// For now, let's assume we can create such data or focus on other errors.

	// Test case 2: Invalid initial motor mass
	invalidMotor := &components.Motor{Props: &thrustcurves.MotorData{TotalMass: 0.0}}
	orkDataValid, err := openrocket.Load("../../testdata/openrocket/l1.ork", "23.09")
	require.NoError(t, err)                                                                   // Check error from Load
	require.NotNil(t, orkDataValid)                                                           // Ensure data loaded
	rocketInvalidMotor := entities.NewRocketEntity(world, &orkDataValid.Rocket, invalidMotor) // Pass address of Rocket field
	assert.Nil(t, rocketInvalidMotor, "Should return nil if initial motor mass is invalid")

	// Test case 3: Invalid total mass calculation
	// This would require mocking component GetMass() or having components with zero/negative mass.
	// Difficult to test directly without more control over component creation/mocking.
}
