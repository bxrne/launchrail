package simulation

import (
	"fmt"
	"math"
	"reflect"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/plugin"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/pkg/atmosphere"
	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/entities"
	openrocket "github.com/bxrne/launchrail/pkg/openrocket"
	pluginapi "github.com/bxrne/launchrail/pkg/plugin"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

// Simulation represents a rocket simulation
type Simulation struct {
	world             *ecs.World
	physicsSystem     *systems.PhysicsSystem
	aerodynamicSystem *systems.AerodynamicSystem
	logParasiteSystem *systems.LogParasiteSystem
	motionParasite    *systems.StorageParasiteSystem
	eventsParasite    *systems.StorageParasiteSystem
	dynamicsParasite  *systems.StorageParasiteSystem
	rulesSystem       *systems.RulesSystem
	rocket            *entities.RocketEntity
	config            *config.Config
	logger            logf.Logger
	updateChan        chan struct{}
	doneChan          chan struct{}
	stateChan         chan *states.PhysicsState
	launchRailSystem  *systems.LaunchRailSystem
	currentTime       float64
	systems           []systems.System
	pluginManager     *plugin.Manager
	gravity           float64
}

// NewSimulation creates a new rocket simulation
func NewSimulation(cfg *config.Config, log logf.Logger, stores *storage.Stores) (*Simulation, error) {
	world := &ecs.World{}

	sim := &Simulation{
		world:         world,
		config:        cfg,
		logger:        log,
		updateChan:    make(chan struct{}),
		doneChan:      make(chan struct{}),
		stateChan:     make(chan *states.PhysicsState, 100),
		pluginManager: plugin.NewManager(log, cfg), // Add cfg argument
		gravity:       9.81,
	}

	for _, pluginPath := range cfg.Setup.Plugins.Paths {
		if err := sim.pluginManager.LoadPlugin(pluginPath); err != nil {
			return nil, err
		}
	}

	// Initialize systems with optimized worker counts
	sim.physicsSystem = systems.NewPhysicsSystem(world, &cfg.Engine, sim.logger, 4)
	sim.aerodynamicSystem = systems.NewAerodynamicSystem(world, atmosphere.NewISAModel(&cfg.Engine.Options.Launchsite.Atmosphere.ISAConfiguration), sim.logger)
	rules := systems.NewRulesSystem(world, &cfg.Engine, sim.logger)

	sim.rulesSystem = rules

	// Initialize launch rail system with config values
	sim.launchRailSystem = systems.NewLaunchRailSystem(
		sim.world,
		sim.config.Engine.Options.Launchrail.Length,
		sim.config.Engine.Options.Launchrail.Angle,
		sim.config.Engine.Options.Launchrail.Orientation,
		&sim.logger,
	)

	// Initialize parasite systems with specific store types
	sim.logParasiteSystem = systems.NewLogParasiteSystem(world, sim.logger)
	sim.motionParasite = systems.NewStorageParasiteSystem(world, stores.Motion, storage.MOTION)
	sim.eventsParasite = systems.NewStorageParasiteSystem(world, stores.Events, storage.EVENTS)
	sim.dynamicsParasite = systems.NewStorageParasiteSystem(world, stores.Dynamics, storage.DYNAMICS)

	// Start parasites (only once)
	sim.logParasiteSystem.Start(sim.stateChan)
	err := sim.motionParasite.Start(sim.stateChan)
	if err != nil {
		return nil, err
	}

	err = sim.eventsParasite.Start(sim.stateChan)
	if err != nil {
		return nil, err
	}

	err = sim.dynamicsParasite.Start(sim.stateChan)
	if err != nil {
		return nil, err
	}

	// Add systems to the slice - Note: we should NOT add the event parasite here
	// as it's meant to be independent
	sim.systems = []systems.System{
		sim.physicsSystem,
		sim.aerodynamicSystem,
		sim.rulesSystem,
		sim.launchRailSystem,
		sim.logParasiteSystem,
	}

	return sim, nil
}

// LoadRocket loads a rocket entity into the simulation
func (s *Simulation) LoadRocket(orkData *openrocket.OpenrocketDocument, motorData *thrustcurves.MotorData) error {
	// Create motor component with logger
	motor, err := components.NewMotor(ecs.NewBasic(), motorData, s.logger)
	if err != nil {
		return err
	}

	// Create rocket entity with all components
	s.rocket = entities.NewRocketEntity(s.world, orkData, motor, &s.logger)

	// Create a single PhysicsEntity to reuse for all systems
	sysEntity := &states.PhysicsState{
		BasicEntity:         *s.rocket.BasicEntity,
		Position:            s.rocket.Position,
		Velocity:            s.rocket.Velocity,
		Acceleration:        s.rocket.Acceleration,
		AngularVelocity:     s.rocket.AngularVelocity,
		AngularAcceleration: s.rocket.AngularAcceleration,
		Orientation:         s.rocket.Orientation,
		Mass:                s.rocket.Mass,
		Motor:               motor,
		Bodytube:            s.rocket.GetComponent("bodytube").(*components.Bodytube),
		Nosecone:            s.rocket.GetComponent("nosecone").(*components.Nosecone),
		Finset:              s.rocket.GetComponent("finset").(*components.TrapezoidFinset),
		Parachute:           s.rocket.GetComponent("parachute").(*components.Parachute),
	}

	// Add to all systems
	s.physicsSystem.Add(sysEntity)
	s.aerodynamicSystem.Add(sysEntity)
	s.rulesSystem.Add(sysEntity)
	s.launchRailSystem.Add(sysEntity)
	s.logParasiteSystem.Add(sysEntity)
	s.motionParasite.Add(sysEntity)
	s.dynamicsParasite.Add(sysEntity)
	s.eventsParasite.Add(sysEntity)

	// Initialize rocket position based on launch rail AFTER all systems are set up and components are ready
	s.launchRailSystem.InitializeRocketPosition(sysEntity)
	s.logger.Info("Rocket position initialized by LaunchRailSystem", "initialPosY", sysEntity.Position.Vec.Y)

	return nil
}

// assertAndLogPhysicsSanity performs assertions and logging for the simulation state.
func (s *Simulation) assertAndLogPhysicsSanity(state *entities.RocketEntity) error {
	if state == nil {
		return nil
	}
	zeroRocketStateIfNoMotor(s, state)
	logAndAssertNaNOrInf(s, "Altitude", state.Position.Vec.Y)
	logAndAssertNaNOrInf(s, "Velocity", state.Velocity.Vec.Y)
	logAndAssertNaNOrInf(s, "Acceleration", state.Acceleration.Vec.Y)
	if state.Mass.Value <= 0 {
		s.logger.Error("ASSERT FAIL: Mass is non-positive", "mass", state.Mass.Value)
		return fmt.Errorf("mass is non-positive")
	}
	if err := assertMotorStateAndLog(s, state); err != nil {
		return err
	}
	return nil
}

// zeroRocketStateIfNoMotor zeroes state if no motor is present and logs a warning.
func zeroRocketStateIfNoMotor(s *Simulation, state *entities.RocketEntity) {
	if state.GetComponent("motor") == nil {
		state.Acceleration.Vec.X = 0
		state.Acceleration.Vec.Y = 0
		state.Acceleration.Vec.Z = 0
		state.Velocity.Vec.X = 0
		state.Velocity.Vec.Y = 0
		state.Velocity.Vec.Z = 0
		state.Position.Vec.X = 0
		state.Position.Vec.Y = 0
		state.Position.Vec.Z = 0
		s.logger.Warn("Zeroed rocket state before assertion", "ax", state.Acceleration.Vec.X, "ay", state.Acceleration.Vec.Y, "az", state.Acceleration.Vec.Z)
	}
}

// logAndAssertNaNOrInf logs an error if the value is NaN or Inf.
func logAndAssertNaNOrInf(s *Simulation, label string, value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		s.logger.Error("ASSERT FAIL: "+label+" is NaN or Inf, ignoring", label, value)
	}
}

// assertMotorStateAndLog checks motor state, logs, and does periodic logging.
func assertMotorStateAndLog(s *Simulation, state *entities.RocketEntity) error {
	if state.Motor != nil {
		if math.Abs(state.Motor.GetThrust()) > 1e6 {
			s.logger.Error("ASSERT FAIL: Thrust out of bounds", "thrust", state.Motor.GetThrust())
			return fmt.Errorf("thrust out of bounds")
		}
		logPeriodicSimState(s, state, true)
	} else {
		logPeriodicSimState(s, state, false)
	}
	return nil
}

// logPeriodicSimState logs the sim state every 100ms, with or without motor.
func logPeriodicSimState(s *Simulation, state *entities.RocketEntity, hasMotor bool) {
	if int(s.currentTime*1000)%100 != 0 {
		return
	}
	if hasMotor {
		s.logger.Debug("Sim state", "t", s.currentTime, "alt", state.Position.Vec.Y, "vy", state.Velocity.Vec.Y, "ay", state.Acceleration.Vec.Y, "mass", state.Mass.Value, "thrust", state.Motor.GetThrust())
	} else {
		s.logger.Warn("Sim state: Motor is nil", "t", s.currentTime, "alt", state.Position.Vec.Y, "vy", state.Velocity.Vec.Y, "ay", state.Acceleration.Vec.Y, "mass", state.Mass.Value)
	}
}

// Run executes the simulation
func (s *Simulation) Run() error {
	defer func() {
		s.logParasiteSystem.Stop()
		s.motionParasite.Stop()
		s.eventsParasite.Stop()
		s.dynamicsParasite.Stop()
	}()

	if s.config.Engine.Simulation.Step <= 0 || s.config.Engine.Simulation.Step > 0.01 {
		return fmt.Errorf("invalid simulation step: must be between 0 and 0.01")
	}

	for {
		if err := s.updateSystems(); err != nil {
			return err
		}
		state := s.rocket // This variable is used by assertAndLogPhysicsSanity
		if err := s.assertAndLogPhysicsSanity(state); err != nil {
			return err
		}

		// Check simulation exit conditions
		if s.shouldStopSimulation() {
			break
		}

		s.currentTime += s.config.Engine.Simulation.Step
	}

	close(s.doneChan)
	return nil
}

// shouldStopSimulation checks all conditions that would require the simulation to stop.
// It consolidates the various exit conditions from the main simulation loop.
func (s *Simulation) shouldStopSimulation() bool {
	// Check for NaN/Inf in primary rocket state after updateSystems and clamping as a final safeguard
	if math.IsNaN(s.rocket.Position.Vec.Y) || math.IsInf(s.rocket.Position.Vec.Y, 0) {
		s.logger.Error("Rocket Y position is NaN or Inf after update and clamping. Stopping simulation.", "posY", s.rocket.Position.Vec.Y)
		return true
	}

	// Ground collision check (using clamped s.rocket.Position.Vec.Y)
	// Ensure currentTime > 0 to allow simulation to start if rocket is initially on the ground (e.g., Y <= tolerance)
	// Compare with Timestep to ensure at least one step has run
	if s.rocket.Position.Vec.Y <= s.config.Engine.Simulation.GroundTolerance && s.currentTime > s.config.Engine.Simulation.Step {
		s.logger.Info("Ground impact detected: Rocket altitude at or below ground tolerance. Stopping simulation.",
			"altitude", s.rocket.Position.Vec.Y, "tolerance", s.config.Engine.Simulation.GroundTolerance, "vy", s.rocket.Velocity.Vec.Y)
		return true
	}

	// Rules system land event check
	// Avoid stopping at t=0 if initial state is Land
	if s.rulesSystem.GetLastEvent() == types.Land && s.currentTime > s.config.Engine.Simulation.Step {
		s.logger.Info("RulesSystem reported Land event. Stopping simulation.", "lastEvent", s.rulesSystem.GetLastEvent())
		return true
	}

	// Max simulation time check
	if s.currentTime >= s.config.Engine.Simulation.MaxTime {
		s.logger.Info("Reached maximum simulation time. Stopping simulation.", "maxTime", s.config.Engine.Simulation.MaxTime)
		return true
	}

	return false
}

// getComponent safely retrieves and type-asserts a component from the rocket.
func getComponent[T any](rocket *entities.RocketEntity, name string) *T {
	c := rocket.GetComponent(name)
	if c == nil {
		return nil
	}
	comp, _ := c.(*T)
	return comp
}

// getSafeMass returns a valid mass pointer, falling back to 1.0 if nil or invalid.
func (s *Simulation) getSafeMass(motor *components.Motor, mass *types.Mass) *types.Mass {
	if motor == nil || mass == nil || mass.Value <= 0 {
		if s.rocket != nil && s.rocket.Mass != nil && s.rocket.Mass.Value > 0 {
			return s.rocket.Mass
		}
		s.logger.Warn("Simulation state: Motor is nil or mass invalid, using fallback mass=1.0")
		return &types.Mass{Value: 1.0}
	}
	return mass
}

// buildPhysicsState constructs a PhysicsState from the rocket and current time.
func (s *Simulation) buildPhysicsState(motor *components.Motor, mass *types.Mass) *states.PhysicsState {
	// Simplified inertia tensor calculation (diagonal body frame)
	// TODO: Replace with more accurate inertia calculation, possibly from OpenRocket data
	// that considers all components and their relative positions for parallel axis theorem.
	var Ixx, Iyy, Izz float64
	bodytube := getComponent[components.Bodytube](s.rocket, "bodytube")

	if bodytube != nil && mass != nil && mass.Value > 0 && bodytube.Radius > 0 {
		r := bodytube.Radius
		l := bodytube.Length
		m := mass.Value // Current total mass of the rocket

		// Assuming X is the longitudinal/roll axis of the rocket.
		// Roll inertia about X for a cylinder: Ixx_cyl = 0.5 * m * r^2
		Ixx = 0.5 * m * r * r

		// Pitch/Yaw inertia (transverse axes Y and Z)
		// For a cylinder rod perpendicular to axis: I_transverse = (1/12) * m * (3*r^2 + L^2)
		if l > 0 {
			Iyy = (1.0 / 12.0) * m * (3*r*r + l*l) // Pitch inertia about Y
			Izz = Iyy                              // Yaw inertia about Z (assuming symmetry)
		} else { // Fallback if length is zero (e.g. sphere or point mass approx)
			Iyy = (2.0 / 5.0) * m * r * r
			Izz = Iyy
		}
	} else {
		// Fallback to unit inertia if components are missing or mass is invalid
		Ixx, Iyy, Izz = 1.0, 1.0, 1.0
		s.logger.Warn("Using fallback unit inertia tensor due to missing rocket components or invalid mass/dimensions.")
	}

	// InertiaTensorBody fields M11, M22, M33 correspond to Ixx, Iyy, Izz about body X, Y, Z axes.
	inertiaBodyCalculated := types.Matrix3x3{M11: Ixx, M22: Iyy, M33: Izz} // Direct struct initialization
	invInertiaBodyCalculated := inertiaBodyCalculated.Inverse()            // Inverse() is a method on Matrix3x3, returns *Matrix3x3

	if invInertiaBodyCalculated == nil { // Inverse() returns nil on error
		s.logger.Error("Failed to calculate inverse of dynamically calculated body inertia tensor, using identity.", "ixx", Ixx, "iyy", Iyy, "izz", Izz)
		invInertiaBodyCalculated = types.IdentityMatrix() // IdentityMatrix returns *Matrix3x3
	}

	return &states.PhysicsState{
		Time:                     s.currentTime,
		BasicEntity:              *s.rocket.BasicEntity,
		Position:                 s.rocket.Position,
		Orientation:              s.rocket.Orientation,
		AngularVelocity:          s.rocket.AngularVelocity,
		AngularAcceleration:      s.rocket.AngularAcceleration,
		Velocity:                 s.rocket.Velocity,
		Acceleration:             s.rocket.Acceleration,
		Mass:                     mass,
		Motor:                    motor,
		Bodytube:                 getComponent[components.Bodytube](s.rocket, "bodytube"),
		Nosecone:                 getComponent[components.Nosecone](s.rocket, "nosecone"),
		Finset:                   getComponent[components.TrapezoidFinset](s.rocket, "finset"),
		Parachute:                getComponent[components.Parachute](s.rocket, "parachute"),
		InertiaTensorBody:        inertiaBodyCalculated,     // Use dynamically calculated value (Matrix3x3)
		InverseInertiaTensorBody: *invInertiaBodyCalculated, // Use its inverse (*Matrix3x3)
	}
}

// runPlugins executes plugin hooks and returns error if any fail.
func runPlugins(plugins []pluginapi.SimulationPlugin, hook func(pluginapi.SimulationPlugin) error) error {
	for _, p := range plugins {
		if err := hook(p); err != nil {
			return err
		}
	}
	return nil
}

// updateSystems updates all systems in the simulation
func (s *Simulation) updateSystems() error {
	s.logger.Debug("updateSystems started", "currentTime", s.currentTime)
	if s.rocket == nil {
		return fmt.Errorf("no rocket entity loaded")
	}
	motor := getComponent[components.Motor](s.rocket, "motor")

	// --- Update Rocket's Mass Property ---
	currentMassKg := s.rocket.GetCurrentMassKg() // Includes motor mass potentially from *previous* step
	s.rocket.Mass.Value = currentMassKg          // Update the main mass value for this step
	s.logger.Debug("Updated rocket mass from GetCurrentMassKg", "massKg", currentMassKg)

	// Tentatively build state - used by plugins and motor update
	// Note: Inertia might be slightly stale here if motor mass changed last step
	tempMass := s.getSafeMass(motor, s.rocket.Mass)
	tempStateForPlugins := s.buildPhysicsState(motor, tempMass)

	// 1. Reset Force/Moment Accumulators for this timestep ON THE MAIN ROCKET ENTITY
	s.rocket.AccumulatedForce = types.Vector3{}
	s.rocket.AccumulatedMoment = types.Vector3{}
	s.logger.Debug("Accumulators reset on s.rocket", "s.rocket.AF", s.rocket.AccumulatedForce, "s.rocket.AM", s.rocket.AccumulatedMoment)

	// 2. Run BeforeSimStep plugins (might modify rocket state indirectly)
	s.logger.Debug("Running BeforeSimStep plugins")
	if err := runPlugins(s.pluginManager.GetPlugins(), func(p pluginapi.SimulationPlugin) error {
		return p.BeforeSimStep(tempStateForPlugins)
	}); err != nil {
		return fmt.Errorf("plugin %s BeforeSimStep error: %w", "unknown", err)
	}

	// 3. Update Motor (changes its internal mass for *next* step's GetCurrentMassKg)
	if motor != nil {
		s.logger.Debug("Updating Motor")
		if err := motor.Update(s.config.Engine.Simulation.Step); err != nil {
			return err
		}
		s.logger.Debug("Motor updated", "thrust", motor.GetThrust(), "elapsedTime", motor.GetElapsedTime())
	}

	// Get final mass state AFTER motor update and inertia recalc for physics systems
	finalMass := s.getSafeMass(motor, s.rocket.Mass)    // Use the mass updated at the start of this func
	finalState := s.buildPhysicsState(motor, finalMass) // Build final state with updated mass & inertia

	// 4. Update Core Physics Systems (using finalState snapshot)
	s.logger.Debug("Starting system update loop")
	for _, sys := range s.systems {
		sysName := reflect.TypeOf(sys).String() // Get system name for logging
		s.logger.Debug("Updating system", "type", sysName)

		// Pass dt to the system's UpdateWithError method
		if err := sys.UpdateWithError(s.config.Engine.Simulation.Step); err != nil {
			s.logger.Error("Error updating system", "type", sysName, "error", err)
			// Potentially handle critical errors, e.g., by stopping the simulation
			// For now, we'll let it continue to gather more data if one system fails minorly
		}

		// Log accumulated forces after specific system updates
		switch sys.(type) {
		case *systems.PhysicsSystem:
			s.logger.Debug("s.rocket.AccumulatedForce after PhysicsSystem", "AF", s.rocket.AccumulatedForce)
		case *systems.AerodynamicSystem:
			s.logger.Debug("s.rocket.AccumulatedForce after AerodynamicSystem", "AF", s.rocket.AccumulatedForce)
		}
	}
	s.logger.Debug("Finished system update loop")
	s.logger.Debug("s.rocket.AccumulatedForce before netForce calculation", "AF", s.rocket.AccumulatedForce)

	// Capture the detected event *after* running rules system
	finalState.CurrentEvent = s.rulesSystem.GetLastEvent()

	// 5. Calculate Net Acceleration from Accumulated Forces ON THE MAIN ROCKET ENTITY
	pos0 := s.rocket.Position.Vec // Value type
	vel0 := s.rocket.Velocity.Vec // Value type
	currentMass := s.rocket.Mass.Value
	currentThrustMagnitude := 0.0
	if s.rocket.Motor != nil && !s.rocket.Motor.IsCoasting() {
		currentThrustMagnitude = s.rocket.Motor.GetThrust()
	}
	currentOrientationQuat := s.rocket.Orientation.Quat // Assuming orientation doesn't change significantly within one RK4 step for thrust vector

	// Derivatives function f(state_vars_for_accel_calc) -> acceleration
	rkEvalLinearAccel := func(currentEvalVel types.Vector3, currentEvalPos types.Vector3, mass float64, thrustMag float64, orientation types.Quaternion) types.Vector3 {
		// Use the actual thrust magnitude from the motor
		// No scaling is needed - let the physics model work with real values

		// Calculate Gravity Force (acts downwards in global frame)
		gravityForce := types.Vector3{Y: -s.gravity * mass}

		// Calculate Thrust Force (acts along rocket body axis, rotated to global frame)
		var thrustForceWorld types.Vector3
		if thrustMag > 0 {
			localThrust := types.Vector3{Y: thrustMag} // Assume thrust acts along the rocket's local +Y axis
			thrustForceWorld = *orientation.RotateVector(&localThrust)
		} else {
			thrustForceWorld = types.Vector3{}
		}

		// Start with gravity and thrust forces
		netForceThisStage := gravityForce.Add(thrustForceWorld)

		// Now add drag force based on current evaluation velocity
		// This properly adapts the drag for each RK4 evaluation point
		if s.rocket.Parachute != nil && s.rocket.Parachute.IsDeployed() {
			// Get velocity magnitude
			velocityMag := math.Sqrt(currentEvalVel.X*currentEvalVel.X +
				currentEvalVel.Y*currentEvalVel.Y +
				currentEvalVel.Z*currentEvalVel.Z)

			if velocityMag > 0.01 { // Only apply drag if moving
				// Calculate drag force magnitude
				area := s.rocket.Parachute.Area
				dragCoeff := s.rocket.Parachute.DragCoefficient
				if dragCoeff <= 0 {
					dragCoeff = 0.8 // Standard fallback
				}

				// Get atmospheric density at current altitude using ISA model for troposphere
				altitude := currentEvalPos.Y
				// Temperature lapse rate in troposphere (K/m)
				lapseRate := -0.0065
				// Sea level values
				T0 := 288.15   // K
				P0 := 101325.0 // Pa
				// Calculate temperature and pressure
				T := T0 + lapseRate*altitude
				P := P0 * math.Pow(T/T0, -9.81/(lapseRate*287.05))
				// Calculate density using ideal gas law
				density := P / (287.05 * T) // More accurate density model

				// Calculate drag force (magnitude = 0.5 * density * velocity^2 * Cd * area)
				dragMagnitude := 0.5 * density * velocityMag * velocityMag * dragCoeff * area

				// Apply a balanced parachute effect - aim for typical 5-10 m/s descent rate
				dragMagnitude *= 2.5

				// Create a unit vector in the opposite direction of velocity
				dragDirX := -currentEvalVel.X / velocityMag
				dragDirY := -currentEvalVel.Y / velocityMag
				dragDirZ := -currentEvalVel.Z / velocityMag

				// Add drag force to net force
				dragForce := types.Vector3{
					X: dragDirX * dragMagnitude,
					Y: dragDirY * dragMagnitude,
					Z: dragDirZ * dragMagnitude,
				}
				netForceThisStage = netForceThisStage.Add(dragForce)
			}
		}

		if mass <= 0 { // Prevent division by zero or negative mass
			s.logger.Error("Invalid mass in rkEvalLinearAccel", "mass", mass)
			return types.Vector3{}
		}
		return netForceThisStage.DivideScalar(mass)
	}

	// Calculate Net Angular Acceleration from Accumulated Moments (World Frame) ON THE MAIN ROCKET ENTITY
	// This uses AccumulatedMoment which should be populated by systems like MotorSystem or AeroSystem if they apply moments.
	var netAngularAccelerationWorld types.Vector3
	rotationMatrix := types.RotationMatrixFromQuaternion(&s.rocket.Orientation.Quat) // Use current orientation
	inertiaTensorWorld := types.TransformInertiaBodyToWorld(&s.rocket.InertiaTensorBody, rotationMatrix)
	inverseInertiaTensorWorld := inertiaTensorWorld.Inverse()

	if inverseInertiaTensorWorld != nil {
		netAngularAccelerationWorld = *inverseInertiaTensorWorld.MultiplyVector(&s.rocket.AccumulatedMoment)
	} else {
		s.logger.Error("World inertia tensor is singular, cannot compute angular acceleration.")
		netAngularAccelerationWorld = types.Vector3{}
	}
	s.logger.Debug("Calculated Net Angular Acceleration (World)", "momentW", s.rocket.AccumulatedMoment, "angAccW", netAngularAccelerationWorld)

	// --- RK4 for Translational Motion ---
	k1VDeriv := vel0
	k1ADeriv := rkEvalLinearAccel(vel0, pos0, currentMass, currentThrustMagnitude, currentOrientationQuat)
	posForK2LinearEval := pos0.Add(k1VDeriv.MultiplyScalar(s.config.Engine.Simulation.Step / 2.0))
	velForK2LinearEval := vel0.Add(k1ADeriv.MultiplyScalar(s.config.Engine.Simulation.Step / 2.0))
	k2VDeriv := velForK2LinearEval
	k2ADeriv := rkEvalLinearAccel(velForK2LinearEval, posForK2LinearEval, currentMass, currentThrustMagnitude, currentOrientationQuat)
	posForK3LinearEval := pos0.Add(k2VDeriv.MultiplyScalar(s.config.Engine.Simulation.Step / 2.0))
	velForK3LinearEval := vel0.Add(k2ADeriv.MultiplyScalar(s.config.Engine.Simulation.Step / 2.0))
	k3VDeriv := velForK3LinearEval
	k3ADeriv := rkEvalLinearAccel(velForK3LinearEval, posForK3LinearEval, currentMass, currentThrustMagnitude, currentOrientationQuat)
	posForK4LinearEval := pos0.Add(k3VDeriv.MultiplyScalar(s.config.Engine.Simulation.Step))
	velForK4LinearEval := vel0.Add(k3ADeriv.MultiplyScalar(s.config.Engine.Simulation.Step))
	k4VDeriv := velForK4LinearEval
	k4ADeriv := rkEvalLinearAccel(velForK4LinearEval, posForK4LinearEval, currentMass, currentThrustMagnitude, currentOrientationQuat)
	finalPos := pos0.Add(
		k1VDeriv.Add(k2VDeriv.MultiplyScalar(2.0)).Add(k3VDeriv.MultiplyScalar(2.0)).Add(k4VDeriv).MultiplyScalar(s.config.Engine.Simulation.Step / 6.0),
	)
	finalVel := vel0.Add(
		k1ADeriv.Add(k2ADeriv.MultiplyScalar(2.0)).Add(k3ADeriv.MultiplyScalar(2.0)).Add(k4ADeriv).MultiplyScalar(s.config.Engine.Simulation.Step / 6.0),
	)

	// NaN/Inf checks for calculated final position and velocity BEFORE assigning to s.rocket
	prevY := s.rocket.Position.Vec.Y
	prevVY := s.rocket.Velocity.Vec.Y
	if math.IsNaN(finalPos.Y) || math.IsInf(finalPos.Y, 0) {
		s.logger.Error("Calculated finalPos.Y is NaN or Inf. Using previous Y position.", "finalPosY_error", finalPos.Y, "prevY", prevY)
		finalPos.Y = prevY // Use previous valid Y to prevent state corruption
	}
	if math.IsNaN(finalVel.Y) || math.IsInf(finalVel.Y, 0) {
		s.logger.Error("Calculated finalVel.Y is NaN or Inf. Resetting Y velocity to previous.", "finalVelY_error", finalVel.Y, "prevVY", prevVY)
		// If prevVY is also suspect (e.g. if NaN propagated from previous step), consider resetting to 0.
		// For now, using prevVY to attempt to maintain some continuity if possible.
		finalVel.Y = prevVY
	}

	s.logger.Debug("RK4 Updated Position", "oldPos", pos0, "newPos", finalPos, "dt", s.config.Engine.Simulation.Step)
	s.logger.Debug("RK4 Updated Velocity", "oldVel", vel0, "newVel", finalVel, "dt", s.config.Engine.Simulation.Step)

	// Update rocket's state with potentially corrected values
	s.rocket.Position.Vec = finalPos
	s.rocket.Velocity.Vec = finalVel
	s.rocket.Acceleration.Vec = k1ADeriv // Storing accel from start of step for logging
	// s.rocket.Orientation.Quat = finalOrientation // TODO: Restore when finalOrientation is available from RK4
	// s.rocket.Mass is updated by MotorSystem.Update()

	// --- Ground Collision Detection and Response on s.rocket ---
	if s.rocket.Position.Vec.Y < s.config.Engine.Simulation.GroundTolerance && s.rocket.Velocity.Vec.Y < 0 {
		s.logger.Debug("Ground impact: clamping Y position and velocity on s.rocket",
			"y_pos_before_clamp", s.rocket.Position.Vec.Y,
			"y_vel_before_clamp", s.rocket.Velocity.Vec.Y,
			"ground_tolerance", s.config.Engine.Simulation.GroundTolerance)

		s.rocket.Position.Vec.Y = s.config.Engine.Simulation.GroundTolerance
		s.rocket.Velocity.Vec.Y = 0
		// Optionally, zero out other velocities/rotations if it's a full stop on ground
		// s.rocket.Velocity.Vec.X = 0
		// s.rocket.Velocity.Vec.Z = 0
		// s.rocket.AngularVelocity.Vec = types.Vector3{X: 0, Y: 0, Z: 0}
	}

	// Absolute clamp to ensure altitude is not negative for s.rocket
	if s.rocket.Position.Vec.Y < 0 {
		s.logger.Debug("s.rocket altitude clamped to 0 to prevent negative values", "original_y_pos", s.rocket.Position.Vec.Y)
		s.rocket.Position.Vec.Y = 0
	}

	// Populate the 'state' object that will be recorded and passed to plugins/channels.
	// This 'state' object was passed into updateSystems.
	// *state.AngularAcceleration was already set using netAngularAccelerationWorld.
	finalState.Time = s.currentTime                 // Assuming PhysicsState has a 'Time' field
	finalState.Position = s.rocket.Position         // Reflects clamped position
	finalState.Velocity = s.rocket.Velocity         // Reflects clamped velocity
	finalState.Acceleration = s.rocket.Acceleration // This is k1ADeriv (accel at start of step)
	finalState.Mass = s.rocket.Mass                 // Assign pointer directly, assuming state.Mass is *types.Mass
	finalState.Orientation = s.rocket.Orientation
	finalState.AngularVelocity = s.rocket.AngularVelocity // s.rocket.AngularVelocity is types.Vector3, state.AngularVelocity should match
	// state.AngularAcceleration is already populated.

	s.logger.Debug("Running AfterSimStep plugins")
	if err := runPlugins(s.pluginManager.GetPlugins(), func(p pluginapi.SimulationPlugin) error {
		return p.AfterSimStep(finalState) // Pass the fully updated and clamped state
	}); err != nil {
		pluginName := "unknown" // Placeholder
		// Consider adding a way to get the actual plugin name if an error occurs
		s.logger.Error("Plugin AfterSimStep error", "plugin", pluginName, "error", err)
		return fmt.Errorf("plugin %s AfterSimStep error: %w", pluginName, err)
	}

	s.logger.Debug("Sending state to channel")
	select {
	case s.stateChan <- finalState:
	default:
		s.logger.Warn("state channel full, dropping frame")
	}

	s.logger.Debug("updateSystems finished")
	return nil
} // Closing brace for updateSystems
