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
	"github.com/zerodha/logf"
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
	entities []*states.PhysicsState
	workers  int
	isa      *atmosphere.ISAModel
	log      logf.Logger
}

// GetAirDensity returns the air density at a given altitude
func (a *AerodynamicSystem) GetAirDensity(altitude float64) float64 {
	return float64(a.getAtmosphericData(float64(altitude)).density)
}

// NewAerodynamicSystem creates a new AerodynamicSystem
func NewAerodynamicSystem(world *ecs.World, workers int, cfg *config.Engine, log logf.Logger) *AerodynamicSystem {
	if cfg == nil || cfg.Options.Launchsite.Atmosphere.ISAConfiguration.SeaLevelDensity == 0 {
		cfg = &config.Engine{
			Options: config.Options{
				Launchsite: config.Launchsite{
					Atmosphere: config.Atmosphere{
						ISAConfiguration: config.ISAConfiguration{
							SeaLevelDensity:     1.225,  // Default sea level density in kg/m³
							SeaLevelPressure:    101325, // Default sea level pressure in Pa
							SeaLevelTemperature: 288.15, // Default sea level temperature in K
						},
					},
				},
			},
		}
	}
	return &AerodynamicSystem{
		world:    world,
		entities: make([]*states.PhysicsState, 0),
		workers:  workers,
		isa:      atmosphere.NewISAModel(&cfg.Options.Launchsite.Atmosphere.ISAConfiguration),
		log:      log,
	}
}

const minDensity = 1e-9 // Minimum physical density clamp

// getAtmosphericData retrieves atmospheric data from cache or calculates it
func (a *AerodynamicSystem) getAtmosphericData(altitude float64) *atmosphericData {
	if a.isa == nil {
		// Fallback to standard sea level values if ISA model isn't initialized
		return &atmosphericData{
			density:     1.225,
			pressure:    101325,
			temperature: 288.15,
			soundSpeed:  340.29,
		}
	}

	isaData := a.isa.GetAtmosphere(altitude)

	density := isaData.Density
	pressure := isaData.Pressure
	temperature := isaData.Temperature
	soundSpeed := isaData.SoundSpeed

	// Validate Temperature
	if temperature <= 0 || math.IsNaN(temperature) || math.IsInf(temperature, 0) {
		a.log.Warn("ISA model returned invalid temperature, using fallback", "altitude", altitude, "temp", temperature)
		temperature = 288.15 // Fallback temperature (sea level)
		// Invalidate related values that depend on temperature if they weren't already bad
		if pressure > 0 {
			pressure = 101325
		}
		if soundSpeed > 0 {
			soundSpeed = 0
		} // Mark sound speed for recalculation/fallback below
	}

	// Validate Density - clamp to minimum, don't use sea level fallback
	if density <= 0 || math.IsNaN(density) || math.IsInf(density, 0) {
		a.log.Warn("ISA model returned invalid density, clamping to minimum", "altitude", altitude, "density", density)
		density = minDensity // Clamp to small positive value
		if pressure > 0 {
			pressure = 1e-5
		} // Also adjust pressure if density was bad? Or leave it if T was okay?
	}

	// Validate Pressure (less critical for drag, but good practice)
	if pressure <= 0 || math.IsNaN(pressure) || math.IsInf(pressure, 0) {
		a.log.Warn("ISA model returned invalid pressure, using fallback", "altitude", altitude, "pressure", pressure)
		pressure = 1e-5 // Fallback to a very small positive pressure
	}

	// Validate or Recalculate Sound Speed
	if soundSpeed <= 0 || math.IsNaN(soundSpeed) || math.IsInf(soundSpeed, 0) {
		// Try recalculating from validated temperature
		// Using isa package constants (assuming they exist: Gamma=1.4, R=287.05)
		recalcSoundSpeed := math.Sqrt(1.4 * 287.05 * temperature)
		if !math.IsNaN(recalcSoundSpeed) && recalcSoundSpeed > 0 {
			a.log.Warn("ISA model returned invalid sound speed, using recalculated value", "altitude", altitude, "isaSoundSpeed", isaData.SoundSpeed, "recalcSoundSpeed", recalcSoundSpeed)
			soundSpeed = recalcSoundSpeed
		} else {
			// Final fallback if recalculation fails
			a.log.Warn("ISA model returned invalid sound speed, recalculation failed, using sea level fallback", "altitude", altitude, "isaSoundSpeed", isaData.SoundSpeed)
			soundSpeed = 340.29
		}
	}

	return &atmosphericData{
		density:     density,
		pressure:    pressure,
		temperature: temperature,
		soundSpeed:  soundSpeed,
	}
}

// GetTemperature calculates the temperature at a given altitude
func (a *AerodynamicSystem) getTemperature(altitude float64) float64 {
	return float64(a.isa.GetTemperature(float64(altitude)))
}

// CalculateDrag now handles atmospheric effects and Mach number
func (a *AerodynamicSystem) CalculateDrag(entity states.PhysicsState) types.Vector3 {
	// Validate inputs
	if a == nil || a.isa == nil || entity.Position == nil || entity.Velocity == nil || entity.Nosecone == nil || entity.Bodytube == nil {
		return types.Vector3{}
	}

	if entity == (states.PhysicsState{}) {
		return types.Vector3{}
	}

	// Get atmospheric data
	atmData := a.getAtmosphericData(entity.Position.Vec.Y)

	// Get vector from pool
	dragForce := vectorPool.Get().(*types.Vector3)
	defer vectorPool.Put(dragForce)

	// Calculate velocity magnitude
	velocity := math.Sqrt(entity.Velocity.Vec.X*entity.Velocity.Vec.X +
		entity.Velocity.Vec.Y*entity.Velocity.Vec.Y +
		entity.Velocity.Vec.Z*entity.Velocity.Vec.Z)
	if math.IsNaN(velocity) || math.IsInf(velocity, 0) || velocity < 0.01 {
		return types.Vector3{} // No force if velocity is invalid or too low
	}

	// Prevent division by zero if sound speed is invalid
	if atmData.soundSpeed <= 0 {
		return types.Vector3{} // Cannot calculate Mach, return zero drag
	}

	// Calculate drag coefficient using Barrowman method
	mach := velocity / atmData.soundSpeed
	cd := a.calculateDragCoeff(mach, entity)

	// Calculate reference area
	area := calculateReferenceArea(entity.Nosecone, entity.Bodytube)

	// Calculate dynamic pressure
	q := 0.5 * atmData.density * velocity * velocity

	// Calculate force magnitude (Cd * q * area)
	forceMagnitude := cd * q * area
	if math.IsNaN(forceMagnitude) || math.IsInf(forceMagnitude, 0) {
		return types.Vector3{} // No force if magnitude calculation is invalid
	}

	// Apply force in opposite direction of velocity
	velVec := types.Vector3{X: entity.Velocity.Vec.X, Y: entity.Velocity.Vec.Y, Z: entity.Velocity.Vec.Z}
	velUnitVec := velVec.Normalize()
	dragForce.X = -velUnitVec.X * forceMagnitude
	dragForce.Y = -velUnitVec.Y * forceMagnitude
	dragForce.Z = -velUnitVec.Z * forceMagnitude

	return *dragForce
}

// calculateReferenceArea calculates the reference area for drag calculations
func calculateReferenceArea(nosecone *components.Nosecone, bodytube *components.Bodytube) float64 {
	if nosecone == nil || bodytube == nil {
		panic("missing geometry components: Nosecone or Bodytube is nil")
	}

	// Use the largest cross-sectional area
	noseArea := math.Pi * nosecone.Radius * nosecone.Radius
	tubeArea := math.Pi * bodytube.Radius * bodytube.Radius
	return math.Max(noseArea, tubeArea)
}

// Update implements parallel force calculation and accumulation
func (a *AerodynamicSystem) Update(dt float64) error {
	a.log.Debug("AerodynamicSystem Update started", "entity_count", len(a.entities), "dt", dt)
	workChan := make(chan *states.PhysicsState, len(a.entities))
	resultChan := make(chan types.Vector3, len(a.entities))
	momentChan := make(chan types.Vector3, len(a.entities))

	var wg sync.WaitGroup
	for i := 0; i < a.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for entity := range workChan {
				if entity == nil {
					continue
				}
				a.log.Debug("Worker processing entity", "worker_id", workerID, "entity_id", entity.Entity.ID())
				force := a.CalculateDrag(*entity)
				moment := a.CalculateAerodynamicMoment(*entity)
				a.log.Debug("Worker calculated force/moment", "worker_id", workerID, "entity_id", entity.Entity.ID(), "force", force, "moment", moment)
				resultChan <- force
				momentChan <- moment
			}
		}(i) // Pass worker ID
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

		a.log.Debug("Calculating aero force",
			"entity_id", entity.Entity.ID(),
			"body_force", force,
			"orientation", entity.Orientation.Quat,
			"global_force", globalForce,
			"mass", entity.Mass.Value,
		)

		// Add the calculated global force to the accumulator
		entity.AccumulatedForce = entity.AccumulatedForce.Add(globalForce)

		// Apply angular accelerations from moments
		moment := <-momentChan
		// Add the calculated moment to the accumulator
		entity.AccumulatedMoment = entity.AccumulatedMoment.Add(moment)

		i++
	}
	return nil
}

// Add adds entities to the system
func (a *AerodynamicSystem) Add(as *states.PhysicsState) {
	a.entities = append(a.entities, as) // Store pointer directly
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

	// Limit altitude to avoid numerical instability
	if altitude > 80000 {
		return 0
	}

	if altitude < 11000 { // troposphere
		return rho0 * math.Pow(1-(L*altitude)/T0, g/(R*L)-1)
	}
	// Add stratosphere calculations if needed
	return rho0 * math.Exp(-g*altitude/(R*T0))
}

// calculateAerodynamicMoment calculates the aerodynamic moments on the entity
func (a *AerodynamicSystem) CalculateAerodynamicMoment(entity states.PhysicsState) types.Vector3 {
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
