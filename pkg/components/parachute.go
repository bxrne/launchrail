package components

import (
	"fmt"
	"math"
	"strconv"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/openrocket"
	"github.com/bxrne/launchrail/pkg/types"
)

// Parachute is a component that allows the entity to descend under drag
// and control its descent rate.
type Parachute struct {
	ID              ecs.BasicEntity
	Name            string
	Position        types.Vector3
	Diameter        float64
	DragCoefficient float64
	Strands         int
	LineLength      float64
	Area            float64
	Trigger         ParachuteTrigger
	DeployAltitude  float64
	DeployDelay     float64
	Deployed        bool
}

// ParachuteTrigger represents the trigger configuration of the parachute
type ParachuteTrigger string

const (
	// ParachuteTriggerNone represents no trigger
	ParachuteTriggerNone ParachuteTrigger = "none"
	// ParachuteTriggerApogee represents an apogee trigger
	ParachuteTriggerApogee ParachuteTrigger = "apogee"
	// ParachuteTriggerEjection represents an ejection charge trigger
	ParachuteTriggerEjection ParachuteTrigger = "ejection"
)

// String returns a string representation of the Parachute struct
func (p *Parachute) String() string {
	return fmt.Sprintf("Parachute{ID={%d %v %v}, Name=%s, Position=%v, Diameter=%.2f, DragCoefficient=%.2f, Strands=%d, LineLength=%.2f, Area=%.2f, Trigger=%s, DeployAltitude=%.2f, DeployDelay=%.2f}", p.ID.ID()-1, p.ID.Parent(), p.ID.Children(), p.Name, p.Position, p.Diameter, p.DragCoefficient, p.Strands, p.LineLength, p.Area, p.Trigger, p.DeployAltitude, p.DeployDelay)
}

// NewParachute creates a new parachute instance
func NewParachute(id ecs.BasicEntity, diameter, dragCoefficient float64, strands int, trigger ParachuteTrigger) *Parachute {
	return &Parachute{
		ID:              id,
		Name:            "",
		Position:        types.Vector3{X: 0, Y: 0, Z: 0},
		Diameter:        diameter,
		DragCoefficient: dragCoefficient,
		Strands:         strands,
		LineLength:      0,
		Area:            0.25 * math.Pi * diameter * diameter,
		Trigger:         trigger,
		DeployAltitude:  0,
		DeployDelay:     0,
	}
}

// parseAuto takes a string in the format "auto 0.75" and returns the float value
func parseAuto(auto string) (float64, error) {
	if auto == "" {
		return 0, fmt.Errorf("empty string")
	}
	if auto == "auto" {
		return 0.8, nil
	}
	val, err := strconv.ParseFloat(auto, 64)
	if err == nil {
		return val, nil
	}
	if len(auto) > 5 && auto[:5] == "auto " {
		return strconv.ParseFloat(auto[5:], 64)
	}
	return 0, fmt.Errorf("cannot parse '%s' as float or 'auto <value>'", auto)
}

// NewParachuteFromORK creates a new parachute instance from an ORK Document
func NewParachuteFromORK(id ecs.BasicEntity, orkData *openrocket.OpenrocketDocument) (*Parachute, error) {
	if orkData == nil {
		return nil, fmt.Errorf("OpenRocket data is nil")
	}
	if len(orkData.Rocket.Subcomponents.Stages) == 0 ||
		orkData.Rocket.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.ID == "" {
		return nil, fmt.Errorf("parachute definition not found or invalid rocket structure in ORK data: no stages or bodytube missing or BodyTube ID is empty")
	}

	orkParachuteDefinition := &orkData.Rocket.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.Parachute

	if orkParachuteDefinition.ID == "" {
		return nil, fmt.Errorf("parachute definition not found or invalid rocket structure in ORK data: parachute missing or ID is empty")
	}

	drag, err := parseAuto(orkParachuteDefinition.CD)
	if err != nil {
		drag = 0.8
	}
	if drag <= 0 {
		drag = 0.8
	}

	deployEvent := orkParachuteDefinition.DeployEvent
	if orkParachuteDefinition.DeploymentConfig.DeployEvent != "" {
		deployEvent = orkParachuteDefinition.DeploymentConfig.DeployEvent
	}

	return &Parachute{
		ID:              id,
		Name:            orkParachuteDefinition.Name,
		Position:        types.Vector3{X: 0, Y: 0, Z: 0},
		Diameter:        orkParachuteDefinition.Diameter,
		DragCoefficient: drag,
		Strands:         orkParachuteDefinition.LineCount,
		LineLength:      orkParachuteDefinition.LineLength,
		Area:            0.25 * math.Pi * orkParachuteDefinition.Diameter * orkParachuteDefinition.Diameter,
		Trigger:         ParachuteTrigger(deployEvent),
		DeployAltitude:  orkParachuteDefinition.DeployAltitude,
		DeployDelay:     orkParachuteDefinition.DeployDelay,
		Deployed:        false,
	}, nil
}

// Update updates the parachute component
func (p *Parachute) Update(dt float64) error {
	return nil
}

// Type returns the type of the component
func (p *Parachute) Type() string {
	return "Parachute"
}

// GetPlanformArea returns the planform area of the parachute
func (p *Parachute) GetPlanformArea() float64 {
	return p.Area
}

// GetMass returns the mass of the parachute component in kg
func (p *Parachute) GetMass() float64 {
	return 0.0
}

// GetDensity returns the density of the Parachute
func (p *Parachute) GetDensity() float64 {
	return 0.0
}

// GetVolume returns the volume of the parachute
func (p *Parachute) GetVolume() float64 {
	return 0.0
}

// GetSurfaceArea returns the surface area of the Parachute
func (p *Parachute) GetSurfaceArea() float64 {
	return 0.0
}

// IsDeployed returns whether the parachute is currently deployed
func (p *Parachute) IsDeployed() bool {
	return p.Deployed
}

// Deploy deploys the parachute
func (p *Parachute) Deploy() {
	p.Deployed = true
}
