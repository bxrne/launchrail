package entities

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/types"
	"github.com/bxrne/launchrail/pkg/openrocket"
)

// Nosecone represents the nosecone entity of a rocket
type Nosecone struct {
	ID             ecs.EntityID
	Position       types.Vector3
	Radius         float64
	Length         float64
	Mass           float64
	ShapeParameter float64
}

// NewNosecone creates a new nosecone instance
func NewNosecone(id ecs.EntityID, radius, length, mass, shapeParameter float64) *Nosecone {
	return &Nosecone{
		ID:             id,
		Position:       types.Vector3{X: 0, Y: 0, Z: 0},
		Radius:         radius,
		Length:         length,
		Mass:           mass,
		ShapeParameter: shapeParameter,
	}
}

// NewNoseconeFromORK creates a new nosecone instance from an ORK Document
func NewNoseconeFromORK(id ecs.EntityID, orkData *openrocket.RocketDocument) *Nosecone {
	orkNosecone := orkData.Subcomponents.Stages[0].SustainerSubcomponents.Nosecone
	return NewNosecone(id, orkNosecone.AftRadius, orkNosecone.Length, orkNosecone.GetMass(), orkNosecone.ShapeParameter)
}

// String returns a string representation of the Nosecone
func (n *Nosecone) String() string {
	return fmt.Sprintf("Nosecone{ID: %d, Position: %v, Radius: %.2f, Length: %.2f, Mass: %.2f, ShapeParameter: %.2f}", n.ID, n.Position, n.Radius, n.Length, n.Mass, n.ShapeParameter)
}

// Update updates the nosecone (currently does nothing)
func (n *Nosecone) Update(dt float64) error {
	// INFO: Empty, just meeting interface requirements
	return nil
}

// Type returns the type of the component
func (n *Nosecone) Type() string {
	return "nosecone"
}
