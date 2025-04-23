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
	Position        types.Vector3
	Diameter        float64
	DragCoefficient float64
	Strands         int
	Area            float64
	Trigger         ParachuteTrigger
	Deployed        bool
}

// ParachuteTrigger represents the trigger configuration of the parachute
type ParachuteTrigger string

const (
	// ParachuteTriggerNone represents no trigger
	ParachuteTriggerNone ParachuteTrigger = "none"
	// ParachuteTriggerApogee represents an apogee trigger
	ParachuteTriggerApogee ParachuteTrigger = "apogee"
)

// String returns a string representation of the Parachute struct
func (p *Parachute) String() string {
	return fmt.Sprintf("Parachute{ID={%d %v %v}, Position=%v, Diameter=%.2f, DragCoefficient=%.2f, Strands=%d, Area=%.2f}", p.ID.ID()-1, p.ID.Parent(), p.ID.Children(), p.Position, p.Diameter, p.DragCoefficient, p.Strands, p.Area)
}

// NewParachute creates a new parachute instance
func NewParachute(id ecs.BasicEntity, diameter, dragCoefficient float64, strands int, trigger ParachuteTrigger) *Parachute {
	return &Parachute{
		ID:              id,
		Position:        types.Vector3{X: 0, Y: 0, Z: 0},
		Diameter:        diameter,
		DragCoefficient: dragCoefficient,
		Strands:         strands,
		Area:            0.25 * math.Pi * diameter * diameter,
		Trigger:         trigger,
	}
}

// parseAuto takes a string in the format "auto 0.75" and returns the float value
func parseAuto(auto string) (float64, error) {
	if auto == "" {
		return 0, fmt.Errorf("empty string")
	}
	if auto == "auto" {
		return 0, nil
	}
	return strconv.ParseFloat(auto[5:], 64)
}

// NewParachuteFromORK creates a new parachute instance from an ORK Document
func NewParachuteFromORK(id ecs.BasicEntity, orkData *openrocket.RocketDocument) (*Parachute, error) {
	if orkData == nil || len(orkData.Subcomponents.Stages) == 0 {
		return nil, fmt.Errorf("invalid OpenRocket data: missing stages")
	}

	orkParachute := orkData.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube.Subcomponents.Parachute

	// take auto out of the cd "auto 0.75" -> 0.75
	drag, err := parseAuto(orkParachute.CD)
	if err != nil {
		return nil, fmt.Errorf("invalid drag coefficient: %s", orkParachute.CD)
	}
	return &Parachute{
		ID:              id,
		Position:        types.Vector3{X: 0, Y: 0, Z: 0},
		Diameter:        orkParachute.Diameter,
		DragCoefficient: drag,
		Strands:         orkParachute.LineCount,
		Area:            0.25 * 3.14159 * orkParachute.Diameter * orkParachute.Diameter,
		Trigger:         ParachuteTrigger(orkParachute.DeployEvent),
	}, nil
}

// Update updates the parachute component
func (p *Parachute) Update(dt float64) error {
	// INFO: Empty, just meeting interface requirements
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

// GetMass returns the mass of the Parachute
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
