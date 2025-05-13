package systems

import (
	"math"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/atmosphere"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

// AerodynamicSystem calculates aerodynamic forces on entities
type AerodynamicSystem struct {
	world    *ecs.World
	entities []*states.PhysicsState
	workers  int
	isa      *atmosphere.ISAModel
	log      logf.Logger
}

// Ensure AerodynamicSystem implements both System interfaces
var _ ecs.System = (*AerodynamicSystem)(nil)
var _ System = (*AerodynamicSystem)(nil)

// NewAerodynamicSystem creates a new aerodynamic system
func NewAerodynamicSystem(world *ecs.World, isa *atmosphere.ISAModel, log logf.Logger) *AerodynamicSystem {
	return &AerodynamicSystem{
		world:    world,
		entities: make([]*states.PhysicsState, 0),
		workers:  1,
		isa:      isa,
		log:      log,
	}
}

// Add implements System interface
func (a *AerodynamicSystem) Add(entity *states.PhysicsState) {
	if entity != nil {
		a.entities = append(a.entities, entity)
	}
}

// Remove implements System interface
func (a *AerodynamicSystem) Remove(basic ecs.BasicEntity) {
	var del int = -1
	for i, e := range a.entities {
		if e.ID() == basic.ID() {
			del = i
			break
		}
	}
	if del >= 0 {
		a.entities = append(a.entities[:del], a.entities[del+1:]...)
	}
}

// Update implements ecs.System interface
func (a *AerodynamicSystem) Update(dt float32) {
	_ = a.update(float64(dt))
}

// UpdateWithError implements System interface
func (a *AerodynamicSystem) UpdateWithError(dt float64) error {
	return a.update(dt)
}

// update is the internal update method
func (a *AerodynamicSystem) update(dt float64) error {
	a.log.Debug("AerodynamicSystem Update started", "entity_count", len(a.entities), "dt", dt)

	for _, entity := range a.entities {
		if entity == nil {
			continue
		}

		// Calculate Drag Force (World Frame)
		dragForce := a.CalculateDrag(entity)

		// Calculate Aerodynamic Moment (Body Frame)
		momentBodyFrame := a.CalculateAerodynamicMoment(*entity)
		var momentWorldFrame types.Vector3

		// Rotate Moment to World Frame if orientation is valid
		if entity.Orientation != nil && entity.Orientation.Quat != (types.Quaternion{}) {
			momentWorldFrame = *entity.Orientation.Quat.RotateVector(&momentBodyFrame)
		}

		// Accumulate forces and moments
		entity.AccumulatedForce = entity.AccumulatedForce.Add(dragForce)
		entity.AccumulatedMoment = entity.AccumulatedMoment.Add(momentWorldFrame)
	}

	return nil
}

// CalculateDrag calculates drag force in world frame
func (a *AerodynamicSystem) CalculateDrag(entity *states.PhysicsState) types.Vector3 {
	// Get atmospheric data at current altitude
	atmData := a.isa.GetAtmosphere(entity.Position.Vec.Y)

	// Calculate velocity magnitude
	velocity := math.Sqrt(entity.Velocity.Vec.X*entity.Velocity.Vec.X +
		entity.Velocity.Vec.Y*entity.Velocity.Vec.Y +
		entity.Velocity.Vec.Z*entity.Velocity.Vec.Z)

	if velocity < 1e-10 {
		return types.Vector3{}
	}

	// Calculate reference area
	refArea := 0.0
	if entity.Bodytube != nil {
		refArea = math.Pi * entity.Bodytube.Radius * entity.Bodytube.Radius
	}

	// Calculate drag coefficient (simplified for now)
	cd := 0.75 // Higher value for small rockets due to relatively larger surface area and subsonic speeds

	// Calculate drag force magnitude
	dragMagnitude := 0.5 * atmData.Density * velocity * velocity * refArea * cd

	// Calculate drag force direction (opposite to velocity)
	dragDir := types.Vector3{
		X: -entity.Velocity.Vec.X / velocity,
		Y: -entity.Velocity.Vec.Y / velocity,
		Z: -entity.Velocity.Vec.Z / velocity,
	}

	// Calculate final drag force
	return types.Vector3{
		X: dragDir.X * dragMagnitude,
		Y: dragDir.Y * dragMagnitude,
		Z: dragDir.Z * dragMagnitude,
	}
}

// CalculateAerodynamicMoment calculates the aerodynamic moments on the entity
func (a *AerodynamicSystem) CalculateAerodynamicMoment(entity states.PhysicsState) types.Vector3 {
	// Get atmospheric data
	atmData := a.isa.GetAtmosphere(entity.Position.Vec.Y)

	// Calculate velocity magnitude
	velocity := math.Sqrt(entity.Velocity.Vec.X*entity.Velocity.Vec.X +
		entity.Velocity.Vec.Y*entity.Velocity.Vec.Y +
		entity.Velocity.Vec.Z*entity.Velocity.Vec.Z)

	if velocity < 0.01 {
		return types.Vector3{} // No moment at very low speeds
	}

	// Calculate reference area
	area := 0.0
	if entity.Bodytube != nil {
		area = math.Pi * entity.Bodytube.Radius * entity.Bodytube.Radius
	}

	// Calculate moment coefficient based on angle of attack
	if entity.Orientation != nil {
		// Get the angle between velocity and rocket's axis
		velocityVec := types.Vector3{X: entity.Velocity.Vec.X, Y: entity.Velocity.Vec.Y, Z: entity.Velocity.Vec.Z}
		rocketAxis := entity.Orientation.Quat.RotateVector(&types.Vector3{Y: 1.0})

		// Calculate angle of attack
		velMag := math.Sqrt(velocityVec.X*velocityVec.X + velocityVec.Y*velocityVec.Y + velocityVec.Z*velocityVec.Z)
		if velMag > 0.1 {
			// Normalize velocity vector
			velUnit := types.Vector3{
				X: velocityVec.X / velMag,
				Y: velocityVec.Y / velMag,
				Z: velocityVec.Z / velMag,
			}

			// Calculate angle of attack using dot product
			cosAngle := velUnit.X*rocketAxis.X + velUnit.Y*rocketAxis.Y + velUnit.Z*rocketAxis.Z
			cosAngle = math.Max(-1.0, math.Min(1.0, cosAngle))
			aoa := math.Acos(cosAngle)

			// Calculate moment coefficient (simplified)
			cm := -0.1 * math.Sin(2*aoa) // Basic stability moment

			// Calculate characteristic length from bodytube
			length := 1.0 // Default length
			if entity.Bodytube != nil {
				length = entity.Bodytube.Length
			}

			// Calculate moment magnitude
			momentMag := 0.5 * atmData.Density * velocity * velocity * area * length * cm

			// Calculate cross product of velocity and rocket axis for moment direction
			cross := types.Vector3{
				X: velUnit.Y*rocketAxis.Z - velUnit.Z*rocketAxis.Y,
				Y: velUnit.Z*rocketAxis.X - velUnit.X*rocketAxis.Z,
				Z: velUnit.X*rocketAxis.Y - velUnit.Y*rocketAxis.X,
			}

			// Return moment vector
			return types.Vector3{
				X: cross.X * momentMag,
				Y: cross.Y * momentMag,
				Z: cross.Z * momentMag,
			}
		}
	}

	return types.Vector3{} // No moment if no orientation or low velocity
}

// CalculateInertia returns a simplified moment of inertia value for pitch/yaw
func CalculateInertia(entity *states.PhysicsState) float64 {
	if entity == nil || entity.Bodytube == nil || entity.Mass == nil || entity.Bodytube.Radius <= 0 || entity.Bodytube.Length <= 0 || entity.Mass.Value <= 0 {
		return 0 // Return 0 for invalid inputs to prevent NaN/Inf later
	}
	// Moment of inertia for a cylinder about an axis perpendicular to the length through the center
	// I = (1/12) * m * (3*r^2 + L^2)
	r := entity.Bodytube.Radius
	l := entity.Bodytube.Length
	m := entity.Mass.Value
	inertia := (1.0 / 12.0) * m * (3*r*r + l*l)

	// Double-check for NaN/Inf just in case, although input checks should prevent this
	if math.IsNaN(inertia) || math.IsInf(inertia, 0) {
		return 0
	}
	return inertia
}
