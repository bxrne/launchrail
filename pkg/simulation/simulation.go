package simulation

import (
	"fmt"
	"math"
	"reflect"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/plugin"
	"github.com/bxrne/launchrail/internal/storage"
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
	}

	for _, pluginPath := range cfg.Setup.Plugins.Paths {
		if err := sim.pluginManager.LoadPlugin(pluginPath); err != nil {
			return nil, err
		}
	}

	// Initialize systems with optimized worker counts
	sim.physicsSystem = systems.NewPhysicsSystem(world, &cfg.Engine, sim.logger, 4)
	sim.aerodynamicSystem = systems.NewAerodynamicSystem(world, 4, &cfg.Engine, sim.logger)
	rules := systems.NewRulesSystem(world, &cfg.Engine, sim.logger)

	sim.rulesSystem = rules

	// Initialize launch rail system with config values
	sim.launchRailSystem = systems.NewLaunchRailSystem(
		world,
		cfg.Engine.Options.Launchrail.Length,
		cfg.Engine.Options.Launchrail.Angle,
		cfg.Engine.Options.Launchrail.Orientation,
	)

	// Initialize parasite systems with specific store types
	sim.logParasiteSystem = systems.NewLogParasiteSystem(world, log)
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
func (s *Simulation) LoadRocket(orkData *openrocket.RocketDocument, motorData *thrustcurves.MotorData) error {
	// Create motor component with logger
	motor, err := components.NewMotor(ecs.NewBasic(), motorData, s.logger)
	if err != nil {
		return err
	}

	// Create rocket entity with all components
	s.rocket = entities.NewRocketEntity(s.world, orkData, motor, &s.logger)

	// Create a single PhysicsEntity to reuse for all systems
	sysEntity := &states.PhysicsState{
		Entity:              s.rocket.BasicEntity,
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

	return nil
}

// assertAndLogPhysicsSanity performs assertions and logging for the simulation state.
func (s *Simulation) assertAndLogPhysicsSanity(state *entities.RocketEntity) error {
	if state == nil {
		return nil
	}
	zeroRocketStateIfNoMotor(s, state)
	s.logger.Warn("Pre-assert acceleration", "ax", state.Acceleration.Vec.X, "ay", state.Acceleration.Vec.Y, "az", state.Acceleration.Vec.Z)
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
		s.logger.Info("Sim state", "t", s.currentTime, "alt", state.Position.Vec.Y, "vy", state.Velocity.Vec.Y, "ay", state.Acceleration.Vec.Y, "mass", state.Mass.Value, "thrust", state.Motor.GetThrust())
	} else {
		s.logger.Warn("Sim state: Motor is nil", "t", s.currentTime, "alt", state.Position.Vec.Y, "vy", state.Velocity.Vec.Y, "ay", state.Acceleration.Vec.Y, "mass", state.Mass.Value)
	}
}

// shouldStopSimulation checks if the simulation should stop and logs the reason.
func (s *Simulation) shouldStopSimulation() bool {
	if s.rulesSystem.GetLastEvent() == types.Land {
		s.logger.Info("Rocket has landed; stopping simulation")
		return true
	}
	if s.currentTime >= s.config.Engine.Simulation.MaxTime {
		s.logger.Info("Reached maximum simulation time")
		return true
	}
	return false
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
		state := s.rocket
		if err := s.assertAndLogPhysicsSanity(state); err != nil {
			return err
		}
		if s.shouldStopSimulation() {
			break
		}
		s.currentTime += s.config.Engine.Simulation.Step
	}

	close(s.doneChan)
	return nil
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
	var Ixx, Iyy, Izz float64
	bodytube := getComponent[components.Bodytube](s.rocket, "bodytube")

	if bodytube != nil && mass != nil && mass.Value > 0 && bodytube.Radius > 0 {
		r := bodytube.Radius
		l := bodytube.Length
		m := mass.Value

		// Roll inertia (longitudinal axis, assuming X)
		Ixx = 0.5 * m * r * r
		// Pitch/Yaw inertia (transverse axes, assuming Y and Z are similar for a cylinder)
		// Using the simplified inertia from aerodynamics.go for cylinder rod perpendicular to axis
		// I = (1/12) * m * (3*r^2 + L^2) for pitch/yaw about CG
		// This CalculateInertia might need to be adapted or made more accessible
		// For now, let's call a static version if available, or use the formula.
		// This calculation is a placeholder as `systems.CalculateInertia` expects a full PhysicsState.
		if l > 0 {
			Iyy = (1.0 / 12.0) * m * (3*r*r + l*l)
			Izz = Iyy // Assume symmetry for pitch and yaw inertia
		} else { // Fallback if length is zero (e.g. sphere)
			Iyy = (2.0 / 5.0) * m * r * r
			Izz = Iyy
		}
	} else {
		// Fallback to unit inertia if components are missing
		Ixx, Iyy, Izz = 1.0, 1.0, 1.0
		s.logger.Warn("Using fallback unit inertia tensor due to missing rocket components or mass.")
	}

	inertiaBody := types.NewMatrix3x3([]float64{
		Ixx, 0, 0,
		0, Iyy, 0,
		0, 0, Izz,
	})
	invInertiaBody := types.NewMatrix3x3([]float64{
		1.0 / Ixx, 0, 0,
		0, 1.0 / Iyy, 0,
		0, 0, 1.0 / Izz,
	})
	if Ixx == 0 || Iyy == 0 || Izz == 0 { // Avoid division by zero if any inertia is zero
		s.logger.Error("Zero component in body inertia tensor, using identity for inverse.", "ixx", Ixx, "iyy", Iyy, "izz", Izz)
		invInertiaBody = types.IdentityMatrix()
	}

	return &states.PhysicsState{
		Time:                     s.currentTime,
		Entity:                   s.rocket.BasicEntity,
		Position:                 s.rocket.Position,
		Orientation:              s.rocket.Orientation,
		AngularVelocity:          s.rocket.AngularVelocity,
		AngularAcceleration:      s.rocket.AngularAcceleration,
		Velocity:                 s.rocket.Velocity,
		Acceleration:             s.rocket.Acceleration,
		Mass:                     mass,
		Motor:                    motor,
		Bodytube:                 bodytube, // Use already fetched bodytube
		Nosecone:                 getComponent[components.Nosecone](s.rocket, "nosecone"),
		Finset:                   getComponent[components.TrapezoidFinset](s.rocket, "finset"),
		Parachute:                getComponent[components.Parachute](s.rocket, "parachute"),
		InertiaTensorBody:        *inertiaBody,
		InverseInertiaTensorBody: *invInertiaBody,
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
	mass := s.getSafeMass(motor, s.rocket.Mass)
	state := s.buildPhysicsState(motor, mass) // state is a snapshot for systems to READ from

	// 1. Reset Force/Moment Accumulators for this timestep ON THE MAIN ROCKET ENTITY
	s.rocket.AccumulatedForce = types.Vector3{}
	s.rocket.AccumulatedMoment = types.Vector3{}
	s.logger.Debug("Accumulators reset on s.rocket", "s.rocket.AF", s.rocket.AccumulatedForce, "s.rocket.AM", s.rocket.AccumulatedMoment)

	// 2. Run BeforeSimStep plugins
	s.logger.Debug("Running BeforeSimStep plugins")
	if err := runPlugins(s.pluginManager.GetPlugins(), func(p pluginapi.SimulationPlugin) error {
		return p.BeforeSimStep(state)
	}); err != nil {
		return fmt.Errorf("plugin %s BeforeSimStep error: %w", "unknown", err)
	}

	if state.Motor != nil {
		s.logger.Debug("Updating Motor")
		if err := state.Motor.Update(s.config.Engine.Simulation.Step); err != nil {
			return err
		}
	}

	s.logger.Debug("Starting system update loop")
	for _, sys := range s.systems {
		sysName := reflect.TypeOf(sys).String() // Get system name for logging
		s.logger.Debug("Updating system", "type", sysName)

		// Pass dt to the system's Update method
		if err := sys.Update(s.config.Engine.Simulation.Step); err != nil {
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
	state.CurrentEvent = s.rulesSystem.GetLastEvent()

	// 3. Calculate Net Acceleration from Accumulated Forces ON THE MAIN ROCKET ENTITY
	netForce := s.rocket.AccumulatedForce // CRITICAL: Use s.rocket's accumulator
	var netAcceleration types.Vector3
	if s.rocket.Mass.Value <= 0 { // CRITICAL: Use s.rocket's mass
		s.logger.Error("Invalid mass for acceleration calculation", "mass", s.rocket.Mass.Value)
		netAcceleration = types.Vector3{}
	} else {
		netAcceleration = netForce.DivideScalar(s.rocket.Mass.Value)
	}
	s.logger.Debug("Calculated Net Acceleration", "netForce", netForce, "mass", s.rocket.Mass.Value, "netAcc", netAcceleration)

	// 3b. Calculate Net Angular Acceleration from Accumulated Moments (World Frame) ON THE MAIN ROCKET ENTITY
	var netAngularAccelerationWorld types.Vector3
	// CRITICAL: Use s.rocket for orientation and inertia tensor
	rotationMatrix := types.RotationMatrixFromQuaternion(&s.rocket.Orientation.Quat)
	inertiaTensorWorld := types.TransformInertiaBodyToWorld(&s.rocket.InertiaTensorBody, rotationMatrix)
	inverseInertiaTensorWorld := inertiaTensorWorld.Inverse()

	if inverseInertiaTensorWorld != nil {
		netAngularAccelerationWorld = *inverseInertiaTensorWorld.MultiplyVector(&s.rocket.AccumulatedMoment) // CRITICAL: Use s.rocket's accumulator
	} else {
		s.logger.Error("World inertia tensor is singular, cannot compute angular acceleration.")
		netAngularAccelerationWorld = types.Vector3{}
	}
	s.logger.Debug("Calculated Net Angular Acceleration (World)", "momentW", s.rocket.AccumulatedMoment, "angAccW", netAngularAccelerationWorld)

	// --- START TEMPORARY DEBUG: Disable angular motion for altitude diagnosis ---
	// netAngularAccelerationWorld = types.Vector3{}
	// s.logger.Debug("TEMPORARY DEBUG: Angular acceleration zeroed for testing linear motion.")
	// --- END TEMPORARY DEBUG ---

	// --- RK4 for Translational Motion ---
	// y = [position, velocity]
	// y_dot = [velocity, acceleration]

	// Initial state for the RK4 step, taken directly from s.rocket
	pos0 := s.rocket.Position.Vec
	vel0 := s.rocket.Velocity.Vec
	angVel0Val := *s.rocket.AngularVelocity // Value for calculations
	orient0Val := s.rocket.Orientation.Quat // Value for calculations

	// Derivatives function f(state_vars_for_accel_calc) -> acceleration
	rkEvalLinearAccel := func(currentEvalVel types.Vector3, currentEvalPos types.Vector3) types.Vector3 {
		return netAcceleration
	}
	// Simplified derivative function for angular acceleration (world frame)
	/*
		rkEvalAngularAccelWorld := func(currentEvalAngVel types.Vector3, currentEvalOrient types.Quaternion) types.Vector3 {
			return netAngularAccelerationWorld
		}
	*/
	// Derivative function for quaternion: dQ/dt = 0.5 * Q * omega_q_body. Returns *Quaternion.
	/*
		rkEvalQuaternionDeriv := func(qEvalVal types.Quaternion, omegaWorldEval types.Vector3) *types.Quaternion {
			qEvalPtr := &qEvalVal // Operate with a pointer if methods expect it
			qEvalInv := qEvalPtr.Inverse()
			omegaBodyVec := qEvalInv.RotateVector(&omegaWorldEval)
			omegaQBody := types.NewQuaternion(0, omegaBodyVec.X, omegaBodyVec.Y, omegaBodyVec.Z)

			// qEvalPtr.Multiply(omegaQBody) returns *Quaternion
			// .Scale(0.5) returns *Quaternion
			return qEvalPtr.Multiply(omegaQBody).Scale(0.5)
		}
	*/

	// --- RK4 for Translational Motion ---
	k1VDeriv := vel0
	k1ADeriv := rkEvalLinearAccel(vel0, pos0)
	posForK2LinearEval := pos0.Add(k1VDeriv.MultiplyScalar(s.config.Engine.Simulation.Step / 2.0))
	velForK2LinearEval := vel0.Add(k1ADeriv.MultiplyScalar(s.config.Engine.Simulation.Step / 2.0))
	k2VDeriv := velForK2LinearEval
	k2ADeriv := rkEvalLinearAccel(velForK2LinearEval, posForK2LinearEval)
	posForK3LinearEval := pos0.Add(k2VDeriv.MultiplyScalar(s.config.Engine.Simulation.Step / 2.0))
	velForK3LinearEval := vel0.Add(k2ADeriv.MultiplyScalar(s.config.Engine.Simulation.Step / 2.0))
	k3VDeriv := velForK3LinearEval
	k3ADeriv := rkEvalLinearAccel(velForK3LinearEval, posForK3LinearEval)
	posForK4LinearEval := pos0.Add(k3VDeriv.MultiplyScalar(s.config.Engine.Simulation.Step))
	velForK4LinearEval := vel0.Add(k3ADeriv.MultiplyScalar(s.config.Engine.Simulation.Step))
	k4VDeriv := velForK4LinearEval
	k4ADeriv := rkEvalLinearAccel(velForK4LinearEval, posForK4LinearEval)
	finalPos := pos0.Add(
		k1VDeriv.Add(k2VDeriv.MultiplyScalar(2.0)).Add(k3VDeriv.MultiplyScalar(2.0)).Add(k4VDeriv).MultiplyScalar(s.config.Engine.Simulation.Step / 6.0),
	)
	finalVel := vel0.Add(
		k1ADeriv.Add(k2ADeriv.MultiplyScalar(2.0)).Add(k3ADeriv.MultiplyScalar(2.0)).Add(k4ADeriv).MultiplyScalar(s.config.Engine.Simulation.Step / 6.0),
	)

	// --- RK4 for Angular Velocity (World Frame) ---
	/*
		// angVel0Val is types.Vector3
		k1AngADeriv := rkEvalAngularAccelWorld(angVel0Val, orient0Val)
		// angVelForKXEval are types.Vector3
		angVelForK2Eval := angVel0Val.Add(k1AngADeriv.MultiplyScalar(s.config.Engine.Simulation.Step / 2.0))
		k2AngADeriv := rkEvalAngularAccelWorld(angVelForK2Eval, orient0Val)
		angVelForK3Eval := angVel0Val.Add(k2AngADeriv.MultiplyScalar(s.config.Engine.Simulation.Step / 2.0))
		k3AngADeriv := rkEvalAngularAccelWorld(angVelForK3Eval, orient0Val)
		angVelForK4Eval := angVel0Val.Add(k3AngADeriv.MultiplyScalar(s.config.Engine.Simulation.Step))
		k4AngADeriv := rkEvalAngularAccelWorld(angVelForK4Eval, orient0Val)

		*state.AngularVelocity = angVel0Val.Add(
			k1AngADeriv.Add(k2AngADeriv.MultiplyScalar(2.0)).Add(k3AngADeriv.MultiplyScalar(2.0)).Add(k4AngADeriv).MultiplyScalar(s.config.Engine.Simulation.Step / 6.0),
		)
	*/

	// --- RK4 for Orientation (Quaternion) ---
	/*
		// kXQDeriv will be *types.Quaternion because rkEvalQuaternionDeriv returns *types.Quaternion
		k1QDerivPtr := rkEvalQuaternionDeriv(orient0Val, angVel0Val)

		// qForKXEval will be values (types.Quaternion) after dereferencing and normalizing pointer results
		// orient0Val is types.Quaternion. Add method is on *Quaternion. Scale is on *Quaternion.
		// Need to convert orient0Val to pointer for Add, or ensure Add can take value + pointer.
		// Let's assume Quaternion methods Add, Scale, Normalize always return new *Quaternion.
		tempQ1 := (&orient0Val).Add(k1QDerivPtr.Scale(s.config.Engine.Simulation.Step / 2.0))
		qForK2EvalVal := *tempQ1.Normalize()
		k2QDerivPtr := rkEvalQuaternionDeriv(qForK2EvalVal, angVelForK2Eval)

		tempQ2 := (&orient0Val).Add(k2QDerivPtr.Scale(s.config.Engine.Simulation.Step / 2.0))
		qForK3EvalVal := *tempQ2.Normalize()
		k3QDerivPtr := rkEvalQuaternionDeriv(qForK3EvalVal, angVelForK3Eval)

		tempQ3 := (&orient0Val).Add(k3QDerivPtr.Scale(s.config.Engine.Simulation.Step))
		qForK4EvalVal := *tempQ3.Normalize()
		k4QDerivPtr := rkEvalQuaternionDeriv(qForK4EvalVal, angVelForK4Eval)

		// Sum of quaternion derivatives (all are *Quaternion, Scale returns *Quaternion, Add returns *Quaternion)
		sumQDerivPtr := k1QDerivPtr.Scale(1.0).Add(k2QDerivPtr.Scale(2.0)).Add(k3QDerivPtr.Scale(2.0)).Add(k4QDerivPtr.Scale(1.0))

		finalOrientationPtr := (&orient0Val).Add(sumQDerivPtr.Scale(s.config.Engine.Simulation.Step / 6.0))
		state.Orientation.Quat = *finalOrientationPtr.Normalize()
	*/

	s.logger.Debug("RK4 Updated Position", "oldPos", pos0, "newPos", finalPos, "dt", s.config.Engine.Simulation.Step)
	s.logger.Debug("RK4 Updated Velocity", "oldVel", vel0, "newVel", finalVel, "dt", s.config.Engine.Simulation.Step)
	s.logger.Debug("RK4 Updated Angular Velocity", "oldAngVel", angVel0Val, "newAngVel", s.rocket.AngularVelocity, "dt", s.config.Engine.Simulation.Step)
	s.logger.Debug("RK4 Updated Orientation", "oldOrient", orient0Val, "newOrient", s.rocket.Orientation.Quat, "dt", s.config.Engine.Simulation.Step)

	// Update Acceleration state for logging/output (with the acceleration at the START of the step)
	state.Acceleration.Vec = netAcceleration
	*state.AngularAcceleration = netAngularAccelerationWorld
	s.logger.Debug("Final State Accel set (RK4 context)", "accel", state.Acceleration.Vec, "angAccel", state.AngularAcceleration)

	// 5. Handle Ground Collision (Simplified: check *after* integration)
	if state.Position.Vec.Y <= s.config.Engine.Simulation.GroundTolerance {
		s.logger.Debug("Ground collision detected", "posY", state.Position.Vec.Y, "velY", state.Velocity.Vec.Y)
		state.Position.Vec.Y = 0 // Clamp to ground
		if state.Velocity.Vec.Y < 0 {
			state.Velocity.Vec.Y = 0 // Stop downward motion
		}
		// Also zero out acceleration? Prevents 'bouncing' calculation on next step if net force is still downwards
		if state.Acceleration.Vec.Y < 0 {
			state.Acceleration.Vec.Y = 0
		}
		s.logger.Debug("State after ground collision adjustment", "pos", state.Position.Vec, "vel", state.Velocity.Vec, "acc", state.Acceleration.Vec)
	}

	s.logger.Debug("Running AfterSimStep plugins")
	if err := runPlugins(s.pluginManager.GetPlugins(), func(p pluginapi.SimulationPlugin) error {
		return p.AfterSimStep(state)
	}); err != nil {
		return fmt.Errorf("plugin %s AfterSimStep error: %w", "unknown", err)
	}

	s.logger.Debug("Updating rocket state from physics state")
	s.rocket.Position.Vec = finalPos // Update s.rocket directly
	s.rocket.Velocity.Vec = finalVel // Update s.rocket directly
	s.rocket.Acceleration.Vec = netAcceleration // Store the initial net acceleration for this step
	// s.rocket.Mass.Value is updated by motor system if fuel is consumed.
	// s.rocket.Orientation.Quat and s.rocket.AngularVelocity would be updated by angular RK4 if enabled.

	s.logger.Debug("Sending state to channel")
	select {
	case s.stateChan <- state:
	default:
		s.logger.Warn("state channel full, dropping frame")
	}

	s.logger.Debug("updateSystems finished")
	return nil
}
