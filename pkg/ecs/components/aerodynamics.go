package components

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/ecs"
	"github.com/bxrne/launchrail/pkg/ecs/types"
)

// Aerodynamics represents the aerodynamics component of an entity.
type Aerodynamics struct {
	Area            float64
	DragCoefficient float64
}

// NewAerodynamics creates a new aerodynamics component.
func NewAerodynamics(dragCoefficient, area float64) *Aerodynamics {
	return &Aerodynamics{
		Area:            area,
		DragCoefficient: dragCoefficient,
	}
}

// CalculateDrag calculates the drag force based on the rocket's velocity and nosecone properties.
func (a *Aerodynamics) CalculateDrag(velocity types.Vector3) types.Vector3 {
	const airDensity = 1.225
	speed := velocity.Magnitude() // use the full magnitude

	if speed < 1e-10 {
		// If velocity is nearly zero, no meaningful drag
		return types.Vector3{}
	}

	dragForceMagnitude := 0.5 * a.DragCoefficient * airDensity * a.Area * speed * speed

	// Normalize velocity to get the direction
	normalizedVel := velocity.DivideScalar(speed)

	// Build drag in the opposite direction
	dragForce := normalizedVel.MultiplyScalar(-dragForceMagnitude)
	return dragForce
}

// String returns a string representation of the Aerodynamics component.
func (a *Aerodynamics) String() string {
	return fmt.Sprintf("Aerodynamics{DragCoefficient: %.2f}", a.DragCoefficient)
}

// Type returns the component type
func (a *Aerodynamics) Type() string {
	return ecs.ComponentAerodynamics
}

// Update updates the aerodynamics component (currently does nothing)
// INFO: Empty, just meeting interface requirements
func (a *Aerodynamics) Update(dt float64) error {
	return nil
}
