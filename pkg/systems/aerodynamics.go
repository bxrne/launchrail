package systems

import (
	"math"

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
func (a *AerodynamicSystem) CalculateDrag(entity *states.PhysicsState) types.Vector3 {
	// Validate inputs
	if a == nil || a.isa == nil || entity == nil || entity.Position == nil || entity.Velocity == nil || entity.Nosecone == nil || entity.Bodytube == nil {
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

	// Apply force in opposite direction of velocity (prepare unit vector for both rocket and parachute drag)
	velVec := types.Vector3{X: entity.Velocity.Vec.X, Y: entity.Velocity.Vec.Y, Z: entity.Velocity.Vec.Z}
	velUnitVec := velVec.Normalize()

	// --- DIAGNOSTIC LOG ---
	if entity.Parachute != nil {
		a.log.Debug("Checking parachute status before drag calculation",
			"entity_id", entity.Entity.ID(),
			"is_deployed", entity.Parachute.IsDeployed(),
			"parachute_area", entity.Parachute.Area,
			"parachute_cd", entity.Parachute.DragCoefficient,
		)
	}
	// --- END DIAGNOSTIC LOG ---

	// Check if parachute is deployed
	if entity.Parachute != nil && entity.Parachute.IsDeployed() {
		// Parachute is deployed, calculate its specific drag
		parachuteArea := entity.Parachute.Area
		parachuteCd := entity.Parachute.DragCoefficient

		if parachuteCd <= 0 { // Fallback if Cd is not valid
			a.log.Error("Parachute drag coefficient is invalid, using fallback", "parachute_cd", parachuteCd, "fallback_cd", 0.8)
			parachuteCd = 0.8
		}

		// Calculate dynamic pressure (q = 0.5 * rho * v^2)
		q := 0.5 * atmData.density * velocity * velocity

		// Calculate parachute drag force magnitude (F_d = Cd * A * q)
		parachuteForceMagnitude := parachuteCd * parachuteArea * q

		// Apply an increased effect factor for parachute drag to ensure it significantly affects descent
		// A realistic parachute creates much more drag than the body when fully deployed
		parachuteEffectFactor := 5.0 // Increased from 1.0 for more realistic parachute effect
		parachuteForceMagnitude *= parachuteEffectFactor

		// Parachute drag details calculated

		// Parachute drag directly opposes velocity vector.
		// Set total dragForce to be the parachute drag. This effectively replaces body drag.
		dragForce.X = -velUnitVec.X * parachuteForceMagnitude
		dragForce.Y = -velUnitVec.Y * parachuteForceMagnitude
		dragForce.Z = -velUnitVec.Z * parachuteForceMagnitude
	} else {
		// Parachute not deployed, calculate body drag as before
		bodyCd := a.calculateDragCoeff(velocity/atmData.soundSpeed, entity)
		bodyArea := calculateReferenceArea(entity.Nosecone, entity.Bodytube)

		// Calculate dynamic pressure (q = 0.5 * rho * v^2)
		q := 0.5 * atmData.density * velocity * velocity

		// Calculate body drag force magnitude (F_d = Cd * A * q)
		bodyForceMagnitude := bodyCd * bodyArea * q

		a.log.Debug("Applying body drag",
			"entity_id", entity.Entity.ID(),
			"velocity", velocity,
			"altitude", entity.Position.Vec.Y,
			"density", atmData.density,
			"mach", velocity/atmData.soundSpeed,
			"body_cd", bodyCd,
			"body_area", bodyArea,
			"force_magnitude", bodyForceMagnitude)

		// Body drag opposes velocity.
		dragForce.X = -velUnitVec.X * bodyForceMagnitude
		dragForce.Y = -velUnitVec.Y * bodyForceMagnitude
		dragForce.Z = -velUnitVec.Z * bodyForceMagnitude
	}

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

// Update implements sequential force calculation and accumulation
func (a *AerodynamicSystem) Update(dt float64) error {
	a.log.Debug("AerodynamicSystem Update started", "entity_count", len(a.entities), "dt", dt)

	for _, entity := range a.entities {
		if entity == nil {
			continue
		}

		a.log.Debug("Processing entity sequentially", "entity_id", entity.Entity.ID())

		// Calculate Drag Force (World Frame)
		dragForce := a.CalculateDrag(entity)
		a.log.Debug("Calculated aerodynamic drag force",
			"entity_id", entity.Entity.ID(),
			"world_drag_force", dragForce,
		)

		// Calculate Aerodynamic Moment (Body Frame)
		momentBodyFrame := a.CalculateAerodynamicMoment(*entity)
		var momentWorldFrame types.Vector3

		// Rotate Moment to World Frame if orientation is valid
		if entity.Orientation != nil && entity.Orientation.Quat != (types.Quaternion{}) {
			momentWorldFrame = *entity.Orientation.Quat.RotateVector(&momentBodyFrame)
		} else {
			momentWorldFrame = types.Vector3{} // No rotation possible, keep moment zero in world frame
		}
		a.log.Debug("Calculated aerodynamic moment",
			"entity_id", entity.Entity.ID(),
			"body_moment", momentBodyFrame,
			"world_moment", momentWorldFrame,
		)

		// Accumulate forces and moments directly onto the entity's state (as value types)
		entity.AccumulatedForce = entity.AccumulatedForce.Add(dragForce)
		entity.AccumulatedMoment = entity.AccumulatedMoment.Add(momentWorldFrame)
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

// calculateDragCoeff calculates the drag coefficient based on Mach number and angle of attack
func (a *AerodynamicSystem) calculateDragCoeff(mach float64, entity *states.PhysicsState) float64 {
	// Base pressure drag coefficient (typical for model rockets)
	baseCd := 0.3

	// Add compressibility effects starting from Mach 0.3
	if mach > 0.3 {
		// Prandtl-Glauert compressibility correction
		compressibilityFactor := 1.0 / math.Sqrt(1.0-math.Min(0.95, mach*mach))
		baseCd *= compressibilityFactor

		// Enhanced transonic effects (M = 0.8 to 1.0)
		if mach > 0.8 && mach < 1.0 {
			// Wave drag rise
			transonicFactor := 1.0 + 5.0*math.Pow(mach-0.8, 2)
			baseCd *= transonicFactor
		}
	}

	// Supersonic drag (M >= 1.0)
	if mach >= 1.0 {
		// Sharp increase at Mach 1
		if mach < 1.2 {
			// Linear interpolation between M=1.0 and M=1.2
			baseCd = 1.2 // More realistic peak drag at Mach 1
		} else {
			// Supersonic drag reduction following 1/M^2 trend
			baseCd = 0.5 + 0.7/math.Pow(mach, 2)
		}
	}

	// Add additional form drag based on angle of attack
	// This is a simplified model that increases drag as the rocket tilts
	if entity.Orientation != nil {
		// Get the angle between velocity and rocket's axis
		velocity := types.Vector3{X: entity.Velocity.Vec.X, Y: entity.Velocity.Vec.Y, Z: entity.Velocity.Vec.Z}
		// Use the quaternion's up vector (0,1,0) as the rocket's axis
		rocketAxis := entity.Orientation.Quat.RotateVector(&types.Vector3{Y: 1.0})
		
		// Calculate angle of attack (in radians)
		velMag := math.Sqrt(velocity.X*velocity.X + velocity.Y*velocity.Y + velocity.Z*velocity.Z)
		if velMag > 0.1 { // Only consider AoA if velocity is significant
			// Normalize velocity vector manually
			velUnit := types.Vector3{
				X: velocity.X / velMag,
				Y: velocity.Y / velMag,
				Z: velocity.Z / velMag,
			}
			// Calculate dot product manually
			cosAngle := velUnit.X*rocketAxis.X + velUnit.Y*rocketAxis.Y + velUnit.Z*rocketAxis.Z
			// Clamp cosAngle to [-1, 1] to avoid acos domain errors
			cosAngle = math.Max(-1.0, math.Min(1.0, cosAngle))
			aoa := math.Acos(cosAngle)
			
			// Add form drag and induced drag from angle of attack
			// Form drag increases with sin^2(AoA)
			formDragFactor := 1.0 + 1.5*math.Sin(aoa)*math.Sin(aoa)
			
			// Induced drag from lift generation, proportional to (CL*alpha)^2
			// Using small angle approximation for lift coefficient
			liftCoeff := 2.0 * math.Sin(aoa) // Approximate CL = 2π*alpha
			inducedDragFactor := 1.0 + 0.1*liftCoeff*liftCoeff
			
			// Combined drag factors
			baseCd *= formDragFactor * inducedDragFactor
		}
	}

	return baseCd
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
