package entities_test

import (
	"encoding/xml"
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/bxrne/launchrail/pkg/entities"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zerodha/logf"
)

func createMockOpenRocketData() *openrocket.OpenrocketDocument {
	return &openrocket.OpenrocketDocument{
		XMLName: xml.Name{Local: "openrocket"},
		Rocket: openrocket.RocketDocument{
			XMLName: xml.Name{Local: "rocket"},
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
									TrapezoidFinsets: []openrocket.TrapezoidFinset{
										{
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
			},
		},
	}
}

func createMockMotor(log logf.Logger) *components.Motor {
	mockMotorData := &thrustcurves.MotorData{
		Designation: designation.Designation("MockMotor-A1-P"),
		TotalMass:   1.0,
		WetMass:     0.5,
		BurnTime:    2.0,
		Thrust:      [][]float64{{0, 10}, {2, 10}},
		MaxThrust:   10.0,
	}

	motor, err := components.NewMotor(ecs.NewBasic(), mockMotorData, log)
	if err != nil {
		log.Error("Failed to create mock motor in test setup", "error", err)
		return nil
	}
	return motor
}

func TestNewRocketEntity(t *testing.T) {
	logPtr := logger.GetLogger("debug")
	world := &ecs.World{}
	orkData := createMockOpenRocketData()
	motor := createMockMotor(*logPtr)
	rocket := entities.NewRocketEntity(world, orkData, motor, logPtr)

	assert.NotNil(t, rocket)
	assert.NotNil(t, rocket.Position)
	assert.NotNil(t, rocket.Velocity)
	assert.NotNil(t, rocket.Acceleration)
	assert.NotNil(t, rocket.Mass)
}

func TestGetComponent(t *testing.T) {
	logPtr := logger.GetLogger("debug")
	world := &ecs.World{}
	orkData := createMockOpenRocketData()
	motor := createMockMotor(*logPtr)
	rocket := entities.NewRocketEntity(world, orkData, motor, logPtr)

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
			component := rocket.GetComponent(tt.component)

			if tt.wantNil {
				assert.Nil(t, component)
			} else {
				assert.NotNil(t, component)
			}
		})
	}
}

func TestNewRocketEntityWithInvalidData(t *testing.T) {
	world := &ecs.World{}
	invalidOrkData := &openrocket.OpenrocketDocument{
		Rocket: openrocket.RocketDocument{
			Subcomponents: openrocket.Subcomponents{
				Stages: []openrocket.RocketStage{
					{
						SustainerSubcomponents: openrocket.SustainerSubcomponents{
							BodyTube: openrocket.BodyTube{
								Radius: "invalid",
							},
						},
					},
				},
			},
		},
	}
	motor := createMockMotor(*logger.GetLogger("debug"))
	rocket := entities.NewRocketEntity(world, invalidOrkData, motor, logger.GetLogger("debug"))

	assert.Nil(t, rocket)
}

func TestGetComponentConcurrency(t *testing.T) {
	logPtr := logger.GetLogger("debug")
	world := &ecs.World{}
	orkData := createMockOpenRocketData()
	motor := createMockMotor(*logPtr)
	rocket := entities.NewRocketEntity(world, orkData, motor, logPtr)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = rocket.GetComponent("motor")
			_ = rocket.GetComponent("bodytube")
			_ = rocket.GetComponent("nosecone")
			_ = rocket.GetComponent("finset")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestNewRocketEntity_NilData(t *testing.T) {
	motor := &components.Motor{}
	logPtr := logger.GetLogger("debug")
	rocket := entities.NewRocketEntity(nil, nil, motor, logPtr)
	assert.Nil(t, rocket, "Should return nil if ORK data is nil")

	rocket = entities.NewRocketEntity(nil, &openrocket.OpenrocketDocument{Rocket: openrocket.RocketDocument{}}, nil, logPtr)
	assert.Nil(t, rocket, "Should return nil if motor is nil")
}

func TestNewRocketEntity_ComponentErrors(t *testing.T) {
	world := &ecs.World{}
	invalidMotor := &components.Motor{Props: &thrustcurves.MotorData{TotalMass: 0.0}}
	orkDataValid, err := openrocket.Load("../../testdata/openrocket/l1.ork", "23.09")
	require.NoError(t, err)
	require.NotNil(t, orkDataValid)
	logPtr := logger.GetLogger("debug")
	rocketInvalidMotor := entities.NewRocketEntity(world, orkDataValid, invalidMotor, logPtr)
	assert.Nil(t, rocketInvalidMotor, "Should return nil if initial motor mass is invalid")
}
