package systems

import (
	"math"
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/atmosphere"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/types"
)

// atmosphericData stores atmospheric data at a given altitude
type atmosphericData struct {
	density     float64
	pressure    float64
	temperature float64
	soundSpeed  float64
}

// AerodynamicSystem calculates aerodynamic forces on entities
type AerodynamicSystem struct {
	world    *ecs.World
	entities []*states.PhysicsState // Change to pointer slice
	workers  int
	isa      *atmosphere.ISAModel
}

// GetAirDensity returns the air density at a given altitude
func (a *AerodynamicSystem) GetAirDensity(altitude float64) float64 {
	return float64(a.getAtmosphericData(float64(altitude)).density)
}

// NewAerodynamicSystem creates a new AerodynamicSystem
func NewAerodynamicSystem(world *ecs.World, workers int, cfg *config.Config) *AerodynamicSystem {
	return &AerodynamicSystem{
		world:    world,
		entities: make([]*states.PhysicsState, 0),
		workers:  workers,
		isa:      atmosphere.NewISAModel(&cfg.Options.Launchsite.Atmosphere.ISAConfiguration),
	}
}

// getAtmosphericData retrieves atmospheric data from cache or calculates it
func (a *AerodynamicSystem) getAtmosphericData(altitude float64) *atmosphericData {
	isaData := a.isa.GetAtmosphere(altitude)
	return &atmosphericData{
		density:     isaData.Density,
		pressure:    isaData.Pressure,
		temperature: isaData.Temperature,
		soundSpeed:  a.isa.GetSpeedOfSound(altitude),
	}
}

// GetTemperature calculates the temperature at a given altitude
func (a *AerodynamicSystem) getTemperature(altitude float64) float64 {
	return float64(a.isa.GetTemperature(float64(altitude)))
}

// CalculateDrag now handles atmospheric effects and Mach number
func (a *AerodynamicSystem) CalculateDrag(entity states.PhysicsState) types.Vector3 {
	// Get atmospheric data
	atmData := a.getAtmosphericData(entity.Position.Vec.Y)

	// Get vector from pool
	dragForce := vectorPool.Get().(*types.Vector3)
	defer vectorPool.Put(dragForce)

	// Calculate mach number
	velocity := math.Sqrt(entity.Velocity.Vec.X*entity.Velocity.Vec.X +
		entity.Velocity.Vec.Y*entity.Velocity.Vec.Y +
		entity.Velocity.Vec.Z*entity.Velocity.Vec.Z)
	machNumber := velocity / atmData.soundSpeed

	// Calculate drag coefficient using Barrowman method
	cd := a.calculateDragCoeff(machNumber, entity)

	// Calculate reference area
	area := calculateReferenceArea(entity.Nosecone, entity.Bodytube)

	// Calculate drag force
	forceMagnitude := 0.5 * cd * atmData.density * area * velocity * velocity

	// Apply force in opposite direction of velocity
	dragForce.X = -entity.Velocity.Vec.X * forceMagnitude / velocity
	dragForce.Y = -entity.Velocity.Vec.Y * forceMagnitude / velocity
	dragForce.Z = -entity.Velocity.Vec.Z * forceMagnitude / velocity

	return *dragForce
}

// calculateReferenceArea calculates the reference area for drag calculations
func calculateReferenceArea(nosecone *components.Nosecone, bodytube *components.Bodytube) float64 {
	// Use the largest cross-sectional area
	noseArea := math.Pi * nosecone.Radius * nosecone.Radius
	tubeArea := math.Pi * bodytube.Radius * bodytube.Radius
	return math.Max(noseArea, tubeArea)
}

// Update implements parallel force calculation and application
func (a *AerodynamicSystem) Update(dt float64) error {
	workChan := make(chan *states.PhysicsState, len(a.entities))
	resultChan := make(chan types.Vector3, len(a.entities))
	momentChan := make(chan types.Vector3, len(a.entities))

	var wg sync.WaitGroup
	for i := 0; i < a.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entity := range workChan {
				if entity == nil {
					continue
				}
				force := a.CalculateDrag(*entity)
				moment := a.calculateAerodynamicMoment(*entity)
				resultChan <- force
				momentChan <- moment
			}
		}()
	}

	for _, entity := range a.entities {
		workChan <- entity
	}
	close(workChan)

	go func() {
		wg.Wait()
		close(resultChan)
		close(momentChan)
	}()

	i := 0
	for force := range resultChan {
		entity := a.entities[i]
		if entity == nil || entity.Orientation == nil {
			i++
			<-momentChan // Consume the moment to keep channels in sync
			continue
		}

		var globalForce types.Vector3
		if entity.Orientation.Quat != (types.Quaternion{}) {
			// Transform force to global coordinates using current orientation
			globalForce = *entity.Orientation.Quat.RotateVector(&force)
		} else {
			globalForce = force // Use untransformed force if no valid orientation
		}

		acc := globalForce.DivideScalar(entity.Mass.Value)

		entity.Acceleration.Vec.X += float64(acc.X)
		entity.Acceleration.Vec.Y += float64(acc.Y)
		entity.Acceleration.Vec.Z += float64(acc.Z)

		// Apply angular accelerations from moments
		moment := <-momentChan
		if entity.AngularAcceleration != nil {
			inertia := calculateInertia(entity)
			angAcc := moment.DivideScalar(inertia)

			entity.AngularAcceleration.X = float64(angAcc.X)
			entity.AngularAcceleration.Y = float64(angAcc.Y)
			entity.AngularAcceleration.Z = float64(angAcc.Z)
		}

		i++
	}
	return nil
}

// Add adds entities to the system
func (a *AerodynamicSystem) Add(as *states.PhysicsState) {
	a.entities = append(a.entities, as) // Store pointer directly
}

// Priority returns the system priority
func (a *AerodynamicSystem) Priority() int {
	return 2
}

// calculateSoundSpeed calculates the speed of sound at a given temperature
func (a *AerodynamicSystem) GetSpeedOfSound(altitude float64) float64 {
	temperature := a.getTemperature(altitude)
	if temperature <= 0 {
		return 340.29 // Return sea level speed of sound as fallback
	}
	return float64(math.Sqrt(float64(1.4 * 287.05 * temperature)))
}

// calculateDragCoeff calculates the drag coefficient based on Mach number
func (a *AerodynamicSystem) calculateDragCoeff(mach float64, entity states.PhysicsState) float64 {
	// More accurate drag coefficient calculation
	baseCd := 0.2 // Subsonic base drag

	_ = entity // Placeholder for future drag coefficient calculations

	// Add wave drag in transonic region
	if mach > 0.8 && mach < 1.2 {
		// Prandtl-Glauert compressibility correction
		baseCd *= 1 / math.Sqrt(1-math.Pow(mach, 2))
	}

	// Supersonic drag
	if mach >= 1.2 {
		baseCd = 0.2 + 0.6*math.Exp(-0.6*(mach-1.2))
	}

	return baseCd
}

// getAtmosphericDensity implements the International Standard Atmosphere model
func getAtmosphericDensity(altitude float64) float64 {
	// Constants for ISA model
	const (
		rho0 = 1.225     // sea level density in kg/m^3
		T0   = 288.15    // sea level temperature in K
		L    = 0.0065    // temperature lapse rate in K/m
		g    = 9.80665   // gravitational acceleration in m/s^2
		R    = 287.05287 // specific gas constant for air in J/(kg·K)
	)

	if altitude < 11000 { // troposphere
		return rho0 * math.Pow(1-(L*altitude)/T0, g/(R*L)-1)
	}
	// Add stratosphere calculations if needed
	return rho0 * math.Exp(-g*altitude/(R*T0))
}

// calculateAerodynamicMoment calculates the aerodynamic moments on the entity
func (a *AerodynamicSystem) calculateAerodynamicMoment(entity states.PhysicsState) types.Vector3 {
	// Get atmospheric data
	atmData := a.getAtmosphericData(entity.Position.Vec.Y)

	// Calculate velocity magnitude
	velocity := math.Sqrt(entity.Velocity.Vec.X*entity.Velocity.Vec.X +
		entity.Velocity.Vec.Y*entity.Velocity.Vec.Y +
		entity.Velocity.Vec.Z*entity.Velocity.Vec.Z)

	if velocity < 0.01 {
		return types.Vector3{} // No moment at very low speeds
	}

	// Calculate angle of attack
	alpha := math.Atan2(entity.Velocity.Vec.Y, entity.Velocity.Vec.X)

	// Calculate moment coefficient (simplified)
	cm := -0.1 * math.Sin(2*alpha) // Basic stability moment

	// Calculate reference area and length
	area := calculateReferenceArea(entity.Nosecone, entity.Bodytube)
	length := entity.Bodytube.Length

	// Calculate moment magnitude
	momentMag := 0.5 * atmData.density * velocity * velocity * area * length * cm

	// Return moment vector (primarily around pitch axis)
	return types.Vector3{
		X: 0,
		Y: momentMag,
		Z: 0,
	}
}

// calculateInertia returns a simplified moment of inertia value
func calculateInertia(entity *states.PhysicsState) float64 {
	// Simple approximation using cylinder formula
	radius := entity.Bodytube.Radius
	length := entity.Bodytube.Length
	mass := entity.Mass.Value

	// I = (1/12) * m * (3r² + l²) for a cylinder
	return (1.0 / 12.0) * mass * (3*radius*radius + length*length)
}
