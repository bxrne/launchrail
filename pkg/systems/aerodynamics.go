package systems

import (
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/entities"
)

// AeroSystem represents the aerodynamics system.
type AeroSystem struct {
	ecs *entities.ECS
}

// NewAeroSystem creates a new aerodynamics system.
func NewAeroSystem(ecs *entities.ECS) *AeroSystem {
	return &AeroSystem{ecs: ecs}
}

// Update updates the aerodynamics system.
func (as *AeroSystem) Update() {
	// Implement aerodynamics update logic here
	as.applyDrag()
}

// applyDrag applies aerodynamic drag to the entities.
func (as *AeroSystem) applyDrag() {
	// Drag coefficient (example value)
	const dragCoefficient = 0.5
	// Air density (example value in kg/m^3)
	const airDensity = 1.225
	// Reference area (example value in m^2)
	const referenceArea = 0.1

	for entity := entities.Entity(1); entity < as.ecs.GetNextEntity(); entity++ {
		if velocity, ok := as.ecs.GetComponent(entity, "Velocity").(*components.Velocity); ok {
			if acceleration, ok := as.ecs.GetComponent(entity, "Acceleration").(*components.Acceleration); ok {
				// Calculate drag force
				dragForce := dragCoefficient * airDensity * referenceArea * velocity.Z * velocity.Z / 2

				// Update acceleration based on drag force
				acceleration.Z -= dragForce / 1000 // Assuming mass of 1000 kg for simplicity
			}
		}
	}
}
