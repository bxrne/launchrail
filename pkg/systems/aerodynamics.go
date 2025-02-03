package systems

import (
	"math"
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/types"
)

var (
	// Cache atmospheric calculations
	atmCache = sync.Map{}
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
	entities []physicsEntity
	workers  int
}

// physicsEntity represents an entity with physics components
func NewAerodynamicSystem(world *ecs.World, workers int) *AerodynamicSystem {
	return &AerodynamicSystem{
		world:    world,
		entities: make([]physicsEntity, 0),
		workers:  workers,
	}
}

// getAtmosphericData retrieves atmospheric data from cache or calculates it
func (a *AerodynamicSystem) getAtmosphericData(altitude float64) *atmosphericData {
	// Round to nearest meter for caching
	roundedAlt := math.Floor(altitude)

	if data, ok := atmCache.Load(roundedAlt); ok {
		return data.(*atmosphericData)
	}

	// Calculate new atmospheric data
	data := &atmosphericData{}
	data.temperature = a.calculateTemperature(altitude)
	data.pressure = a.calculatePressure(altitude, data.temperature)
	data.density = a.calculateDensity(data.pressure, data.temperature)
	data.soundSpeed = a.calculateSoundSpeed(data.temperature)

	atmCache.Store(roundedAlt, data)
	return data
}

// CalculateDrag now handles atmospheric effects and Mach number
func (a *AerodynamicSystem) CalculateDrag(entity physicsEntity) types.Vector3 {
	// Get atmospheric data
	atmData := a.getAtmosphericData(entity.Position.Y)

	// Get vector from pool
	dragForce := vectorPool.Get().(*types.Vector3)
	defer vectorPool.Put(dragForce)

	// Calculate mach number
	velocity := math.Sqrt(entity.Velocity.X*entity.Velocity.X +
		entity.Velocity.Y*entity.Velocity.Y +
		entity.Velocity.Z*entity.Velocity.Z)
	machNumber := velocity / atmData.soundSpeed

	// Calculate drag coefficient using Barrowman method
	cd := a.calculateDragCoeff(machNumber, entity)

	// Calculate reference area
	area := calculateReferenceArea(entity.Nosecone, entity.Bodytube)

	// Calculate drag force
	forceMagnitude := 0.5 * cd * atmData.density * area * velocity * velocity

	// Apply force in opposite direction of velocity
	dragForce.X = -entity.Velocity.X * forceMagnitude / velocity
	dragForce.Y = -entity.Velocity.Y * forceMagnitude / velocity
	dragForce.Z = -entity.Velocity.Z * forceMagnitude / velocity

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
func (a *AerodynamicSystem) Update(dt float32) error {
	workChan := make(chan physicsEntity, len(a.entities))
	resultChan := make(chan types.Vector3, len(a.entities))

	var wg sync.WaitGroup
	for i := 0; i < a.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entity := range workChan {
				// Perform a more accurate Mach-based drag calculation
				force := a.CalculateDrag(entity)
				resultChan <- force
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
	}()

	i := 0
	for force := range resultChan {
		entity := a.entities[i]
		acc := force.DivideScalar(entity.Mass.Value)
		entity.Acceleration.X += float64(acc.X)
		entity.Acceleration.Y += float64(acc.Y)
		entity.Acceleration.Z += float64(acc.Z)
		i++
	}
	return nil
}

// SystemEntity represents an entity with physics components (Meta rocket)
type SystemEntity struct {
	Entity   *ecs.BasicEntity
	Pos      *components.Position
	Vel      *components.Velocity
	Acc      *components.Acceleration
	Mass     *components.Mass
	Motor    *components.Motor
	Bodytube *components.Bodytube
	Nosecone *components.Nosecone
	Finset   *components.TrapezoidFinset
}

// Add adds entities to the system
func (a *AerodynamicSystem) Add(as *SystemEntity) {
	a.entities = append(a.entities, physicsEntity{as.Entity, as.Pos, as.Vel, as.Acc, as.Mass, as.Motor, as.Bodytube, as.Nosecone, as.Finset})
}

// Priority returns the system priority
func (a *AerodynamicSystem) Priority() int {
	return 2
}

// Add these methods to AerodynamicSystem
func (a *AerodynamicSystem) calculateTemperature(altitude float64) float64 {
	const (
		T0 = 288.15 // K (15°C)
		L  = 0.0065 // K/m
	)
	return T0 - L*altitude
}

// calculatePressure calculates atmospheric pressure at a given altitude
func (a *AerodynamicSystem) calculatePressure(altitude, temperature float64) float64 {
	const (
		P0 = 101325 // Pa
		g  = 9.81   // m/s²
		R  = 287.05 // J/(kg·K)
		T0 = 288.15 // K
	)
	return P0 * math.Pow(temperature/T0, -g/(R*0.0065))
}

// calculateDensity calculates air density at a given pressure and temperature
func (a *AerodynamicSystem) calculateDensity(pressure, temperature float64) float64 {
	const R = 287.05 // J/(kg·K)
	return pressure / (R * temperature)
}

// calculateSoundSpeed calculates the speed of sound at a given temperature
func (a *AerodynamicSystem) calculateSoundSpeed(temperature float64) float64 {
	const (
		gamma = 1.4    // ratio of specific heats
		R     = 287.05 // J/(kg·K)
	)
	return math.Sqrt(gamma * R * temperature)
}

// calculateDragCoeff calculates the drag coefficient based on Mach number
func (a *AerodynamicSystem) calculateDragCoeff(mach float64, entity physicsEntity) float64 {
	// Basic drag coefficient calculation
	baseCd := 0.2 // Base drag coefficient
	// Transonic drag rise
	if mach > 0.8 && mach < 1.2 {
		baseCd *= 1 + 5*(mach-0.8)
	}

	// Supersonic drag
	if mach >= 1.2 {
		baseCd = 0.3 / math.Sqrt(mach)
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
