package entities

import (
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/openrocket"
)

// RocketEntity represents a complete rocket with all its components
type RocketEntity struct {
	*ecs.BasicEntity
	*components.Position
	*components.Velocity
	*components.Acceleration
	*components.Mass
	components map[string]interface{} // Change to map for easier lookup
	mu         sync.RWMutex
}

// NewRocketEntity creates a new rocket entity from OpenRocket data
func NewRocketEntity(world *ecs.World, orkData *openrocket.RocketDocument, motor *components.Motor) *RocketEntity {
	if orkData == nil || motor == nil {
		return nil
	}

	basic := ecs.NewBasic()

	// Create base rocket entity
	rocket := &RocketEntity{
		BasicEntity:  &basic,
		Position:     &components.Position{BasicEntity: basic},
		Velocity:     &components.Velocity{BasicEntity: basic},
		Acceleration: &components.Acceleration{BasicEntity: basic},
		Mass:         &components.Mass{BasicEntity: basic},
		components:   make(map[string]interface{}),
	}

	// Store components with proper error handling
	rocket.components["motor"] = motor

	bodytube, err := components.NewBodytubeFromORK(ecs.NewBasic(), orkData)
	if err != nil {
		return nil
	}
	rocket.components["bodytube"] = bodytube

	nosecone := components.NewNoseconeFromORK(ecs.NewBasic(), orkData)
	if nosecone == nil {
		return nil
	}
	rocket.components["nosecone"] = nosecone

	finset := components.NewTrapezoidFinsetFromORK(ecs.NewBasic(), orkData)
	if finset == nil {
		return nil
	}
	rocket.components["finset"] = finset

	// Calculate total mass
	rocket.Mass.Value = calculateTotalMass(orkData)

	return rocket
}

// Calculate total mass by summing all components from OpenRocket data
func calculateTotalMass(orkData *openrocket.RocketDocument) float64 {
	var totalMass float64

	// Get masses from OpenRocket components
	nosecone := orkData.Subcomponents.Stages[0].SustainerSubcomponents.Nosecone
	bodytube := orkData.Subcomponents.Stages[0].SustainerSubcomponents.BodyTube

	totalMass += nosecone.GetMass() + bodytube.GetMass()
	// Add other component masses...

	return totalMass
}

// AddComponent adds a component to the entity
func (r *RocketEntity) GetComponent(name string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.components[name]
}
