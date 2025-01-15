package entities

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/ecs/components"
	"github.com/bxrne/launchrail/pkg/ecs/types"
	"github.com/bxrne/launchrail/pkg/openrocket"
)

type Rocket struct {
	ID           int
	Position     types.Vector3
	Velocity     types.Vector3
	Acceleration types.Vector3
	Mass         float64
	Forces       []types.Vector3
	Components   []components.Component
}

// NewRocket creates a new rocket instance
func NewRocket(mass float64, components ...components.Component) *Rocket {
	return &Rocket{
		ID:           0,
		Position:     types.Vector3{X: 0, Y: 0, Z: 0},
		Velocity:     types.Vector3{X: 0, Y: 0, Z: 0},
		Acceleration: types.Vector3{X: 0, Y: 0, Z: 0},
		Forces:       []types.Vector3{},
		Mass:         mass,
		Components:   components,
	}
}

// NewRocketFromORK creates a new rocket instance from an ORK Document
func NewRocketFromORK(orkData *openrocket.RocketDocument) (*Rocket, error) {
	// TODO: Decompose ork
	// subs := orkData.Subcomponents.List()
	// nosecone := subs[0].SustainerSubcomponents.Nosecone
	// bodytube := subs[0].SustainerSubcomponents.BodyTube

	return &Rocket{
		ID:           0,
		Position:     types.Vector3{X: 0, Y: 0, Z: 0},
		Velocity:     types.Vector3{X: 0, Y: 0, Z: 0},
		Acceleration: types.Vector3{X: 0, Y: 0, Z: 0},
		Forces:       []types.Vector3{},
		Mass:         1.0,
		Components:   []components.Component{},
	}, nil

}

// NewRocketWithID creates a new rocket instance with an ID
func (r *Rocket) String() string {
	return fmt.Sprintf("Rocket{ID: %d, Position: %v, Velocity: %v, Acceleration: %v, Mass: %.2f, Forces: %v, Components: %v}", r.ID, r.Position, r.Velocity, r.Acceleration, r.Mass, r.Forces, r.Components)
}

// Describe returns a string representation of the rocket
func (r *Rocket) Describe() string {
	return fmt.Sprintf("Rocket{ID: %d, Position: %v, Velocity: %v, Acceleration: %v, Mass: %.2f}", r.ID, r.Position, r.Velocity, r.Acceleration, r.Mass)
}

// AddComponent adds a component to the rocket
func (r *Rocket) AddComponent(c components.Component) {
	r.Components = append(r.Components, c)
}

// RemoveComponent removes a component from the rocket
func (r *Rocket) RemoveComponent(c components.Component) {
	for i, component := range r.Components {
		if component == c {
			r.Components = append(r.Components[:i], r.Components[i+1:]...)
		}
	}
}

// Update updates the rocket
func (r *Rocket) Update(dt float64) {
	for _, component := range r.Components {
		component.Update(dt)
	}
}
