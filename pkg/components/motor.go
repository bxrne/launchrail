package components

import (
	"context"
	"fmt"
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/pkg/thrustcurves"
	"github.com/bxrne/launchrail/pkg/types"
	"github.com/zerodha/logf"
)

type Motor struct {
	ID          ecs.BasicEntity
	Position    types.Vector3
	Thrustcurve [][]float64
	Mass        float64 // Current total mass (casing + current propellant)
	Length      float64 // Length of the motor casing
	Diameter    float64 // Diameter of the motor casing
	thrust      float64
	Props       *thrustcurves.MotorData
	FSM         *MotorFSM
	elapsedTime float64
	mu          sync.RWMutex
	burnTime    float64
	logger      logf.Logger

	// New mass fields
	initialPropellantMass float64
	casingMass            float64
	currentPropellantMass float64

	// Efficiency factors
	nozzleEff     float64 // Nozzle efficiency (typical range 0.85-0.98)
	combustionEff float64 // Combustion efficiency (typical range 0.90-0.98)
	frictionEff   float64 // Friction losses (typical range 0.97-0.99)
}

// NewMotor creates a new motor component from thrust curve data
func NewMotor(id ecs.BasicEntity, md *thrustcurves.MotorData, logger logf.Logger) (*Motor, error) {
	if md == nil || len(md.Thrust) == 0 {
		return nil, fmt.Errorf("thrust curve data is required")
	}

	if md.MaxThrust <= 0 {
		md.MaxThrust = findMaxThrust(md.Thrust)
	}

	thrustcurve := validateThrustCurve(md.Thrust)
	burnTime := md.BurnTime
	if len(thrustcurve) > 0 {
		lastCurveTime := thrustcurve[len(thrustcurve)-1][0]
		// Only warn if thrust curve ends before burnTime
		if lastCurveTime < burnTime-1e-6 {
			logger.Warn("Thrust curve ends before official burnTime; thrust will be zero after curve ends", "curveEnd", lastCurveTime, "burnTime", burnTime)
		}
	}

	// Calculate initial propellant and casing mass
	// md.WetMass is propellant mass in kg (from thrustcurves.Load)
	// md.TotalMass is initial total motor mass in kg (from thrustcurves.Load)
	initialPropellantMass := md.WetMass
	casingMass := md.TotalMass - md.WetMass
	if casingMass < 0 {
		logger.Warn("Calculated casing mass is negative. Clamping to zero.", "totalMassKg", md.TotalMass, "propellantMassKg", md.WetMass)
		casingMass = 0 // Or handle as an error if preferred
	}

	m := &Motor{
		ID:          id,
		Position:    types.Vector3{},
		Thrustcurve: thrustcurve,
		Mass:        md.TotalMass, // Initial total mass
		Length:      0,            // Initialize length to 0, will be set from ORK data
		Diameter:    0,            // Initialize diameter to 0, will be set from ORK data
		Props:       md,
		thrust:      0, // Initialize thrust to 0, will be set by FSM/Update
		burnTime:    burnTime,
		logger:      logger,

		initialPropellantMass: initialPropellantMass,
		casingMass:            casingMass,
		currentPropellantMass: initialPropellantMass,

		// Set default efficiency factors
		nozzleEff:     0.95, // 95% nozzle efficiency
		combustionEff: 0.95, // 95% combustion efficiency
		frictionEff:   0.95, // 95% friction efficiency
	}

	// Initialize thrust to first data point if available and burn time > 0
	if len(m.Thrustcurve) > 0 && m.burnTime > 0 {
		// Apply efficiency factors to initial thrust
		efficiencyFactor := m.nozzleEff * m.combustionEff * m.frictionEff
		m.thrust = m.Thrustcurve[0][1] * efficiencyFactor
	} else {
		m.thrust = 0
	}

	m.FSM = NewMotorFSM(m, logger) // Initialize FSM here, passing the motor instance

	m.logger.Info("Motor created", "ID", m.ID.ID(),
		"InitialTotalMassKg", m.Mass,
		"InitialPropellantMassKg", m.initialPropellantMass,
		"CasingMassKg", m.casingMass,
		"BurnTimeSec", m.burnTime)

	err := m.FSM.Event(context.Background(), "ignite")
	if err != nil {
		return nil, fmt.Errorf("failed to transition motor state to ignited: %w", err)
	}
	return m, nil
}

// updateBurningState handles the motor's logic when it's in the burning state.
// This includes interpolating thrust, calculating mass loss, and updating mass.
// This method assumes it's called when m.FSM.Current() == StateBurning,
// m.currentPropellantMass > 0, and m.burnTime > 0.
func (m *Motor) updateBurningState(dt float64) {
	// Interpolate thrust based on elapsed time
	m.thrust = m.interpolateThrust(m.elapsedTime)

	// Calculate mass loss based on elapsed time and total burn time
	// For constant thrust, mass decreases linearly with time
	// Mass(t) = InitialMass - (InitialMass - FinalMass) * (t/BurnTime)
	// where FinalMass = CasingMass

	// If we're going to exceed burn time in this step, consume all propellant
	if m.elapsedTime >= m.burnTime {
		m.currentPropellantMass = 0
		m.Mass = m.casingMass
		m.thrust = 0
		return
	}

	// Calculate what the mass should be at this point in time
	targetMass := (m.initialPropellantMass + m.casingMass) - (m.initialPropellantMass * (m.elapsedTime / m.burnTime))

	// If this update will take us past burn time, set mass to casing mass
	if m.elapsedTime+dt >= m.burnTime {
		targetMass = m.casingMass
	}

	// Update current propellant mass and total mass
	m.currentPropellantMass = targetMass - m.casingMass
	m.Mass = targetMass
}

func (m *Motor) Update(dt float64) error {
	if dt < 0 {
		return fmt.Errorf("invalid negative timestep")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update elapsed time
	m.elapsedTime += dt

	// Update the motor FSM state first based on new elapsed time
	err := m.FSM.UpdateState(m.Mass, m.elapsedTime, m.burnTime)
	if err != nil {
		return fmt.Errorf("failed to update motor state: %w", err)
	}

	currentState := m.FSM.Current()

	if currentState == StateBurning && m.currentPropellantMass > 0 && m.burnTime > 0 {
		m.updateBurningState(dt)
	} else if currentState == StateCoasting {
		// When transitioning to coasting, consume all remaining propellant
		m.currentPropellantMass = 0
		m.Mass = m.casingMass
		m.thrust = 0
	} else {
		// If not burning (e.g., pre-ignition), thrust is zero
		m.thrust = 0
	}
	return nil
}

func (m *Motor) interpolateThrust(totalDt float64) float64 {
	// Combined efficiency
	efficiencyFactor := m.nozzleEff * m.combustionEff * m.frictionEff // About 0.86

	// If before burn start, use initial thrust
	if totalDt <= m.Thrustcurve[0][0] {
		return m.Thrustcurve[0][1] * efficiencyFactor
	}

	// If past burn time, return 0
	if totalDt > m.burnTime {
		return 0
	}

	// Find the surrounding data points for interpolation
	for i := 0; i < len(m.Thrustcurve)-1; i++ {
		t1, thrust1 := m.Thrustcurve[i][0], m.Thrustcurve[i][1]
		t2, thrust2 := m.Thrustcurve[i+1][0], m.Thrustcurve[i+1][1]

		if totalDt >= t1 && totalDt <= t2 {
			// Linear interpolation
			ratio := (totalDt - t1) / (t2 - t1)
			thrust := (thrust1 + (ratio * (thrust2 - thrust1))) * efficiencyFactor
			m.logger.Debug("interpolateThrust", "totalDt", totalDt, "t1", t1, "t2", t2, "thrust1", thrust1, "thrust2", thrust2, "thrust", thrust)
			return thrust
		}
	}

	// If we're between last data point and burn time
	// Use the last thrust value
	thrust := m.Thrustcurve[len(m.Thrustcurve)-1][1] * efficiencyFactor
	m.logger.Debug("interpolateThrust (last value)", "totalDt", totalDt, "thrust", thrust)
	return thrust
}

func (m *Motor) GetThrust() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.thrust
}

func (m *Motor) IsCoasting() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.FSM.Current() == "coast"
}

func (m *Motor) GetState() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.FSM.Current()
}

func (m *Motor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.elapsedTime = 0
	m.Mass = m.Props.TotalMass
	m.thrust = m.Thrustcurve[0][1]
	// m.coasting = false // Removed, FSM handles state
	m.FSM = NewMotorFSM(m, m.logger) // Pass motor and logger
	err := m.FSM.Event(context.Background(), "ignite")
	if err != nil {
		m.logger.Error("failed to transition to idle state", "error", err)
	}
}

func (m *Motor) SetState(state string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FSM.SetState(state)
}

func (m *Motor) GetMass() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Mass
}

func (m *Motor) GetCasingMass() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.casingMass
}

func (m *Motor) Type() string {
	return "Motor"
}

func (m *Motor) String() string {
	return fmt.Sprintf("Motor{ID: %d, Position: %s, Mass: %f, Thrust: %f}", m.ID.ID(), m.Position.String(), m.Mass, m.thrust)
}

func (m *Motor) GetPlanformArea() float64 {
	return 0
}

func (m *Motor) GetElapsedTime() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.elapsedTime
}

func (m *Motor) GetPosition() types.Vector3 {
	// TODO: Ensure m.Position is correctly set during NewMotorFromORK relative to a common rocket origin.
	return m.Position
}

func (m *Motor) GetCenterOfMassLocal() types.Vector3 {
	// Assuming motor is aligned with local Z-axis, base at Z=0, tip at Z=m.Length.
	// Simplified: CG is at the geometric center of the motor casing.
	// More accurate model would account for propellant burn-off and shift in CG.
	if m.Length == 0 {
		// m.logger is a struct, so it cannot be nil.
		// It's assumed to be initialized by NewMotor. If its internal writer is nil,
		// the logf methods should handle that or it's an initialization bug in NewMotor.
		m.logger.Warn("Motor.GetCenterOfMassLocal() called with zero length. Returning zero vector.")
		return types.Vector3{X: 0, Y: 0, Z: 0}
	}
	return types.Vector3{X: 0, Y: 0, Z: m.Length / 2.0}
}

func (m *Motor) GetInertiaTensorLocal() types.Matrix3x3 {
	// Approximating the motor as a solid cylinder of total current mass `m.Mass`,
	// length `m.Length`, and radius `R = m.Diameter / 2`.
	// The inertia tensor is calculated about its CG (assumed at m.Length/2 along Z-axis),
	// with Z-axis aligned with the motor's length.

	if m.Mass <= 1e-9 { // Effectively zero mass
		m.logger.Warn("Motor.GetInertiaTensorLocal() called with near-zero mass. Returning zero matrix.")
		return types.Matrix3x3{}
	}
	if m.Length == 0 || m.Diameter == 0 {
		m.logger.Warn("Motor.GetInertiaTensorLocal() called with zero length or diameter. Returning zero matrix.")
		return types.Matrix3x3{}
	}

	mass := m.GetMass() // Use GetMass() to ensure thread-safety if underlying mass changes
	length := m.Length
	radius := m.Diameter / 2.0

	// Formulas for a solid cylinder about its CG:
	// Ixx = Iyy = (1/12) * mass * (3*radius^2 + length^2)
	// Izz = (1/2) * mass * radius^2
	// Assuming Z is the longitudinal axis of the motor.

	ixx := (1.0 / 12.0) * mass * (3*radius*radius + length*length)
	iyy := ixx // Symmetric for a cylinder about X and Y axes perpendicular to Z
	izz := (1.0 / 2.0) * mass * radius * radius

	return types.Matrix3x3{
		M11: ixx, M12: 0, M13: 0,
		M21: 0, M22: iyy, M23: 0,
		M31: 0, M32: 0, M33: izz,
	}
}

func validateThrustCurve(curve [][]float64) [][]float64 {
	if len(curve) < 2 {
		panic("thrust curve must have at least 2 points")
	}

	// Ensure time points are monotonically increasing
	for i := 1; i < len(curve); i++ {
		if curve[i][0] <= curve[i-1][0] {
			panic("thrust curve time points must be strictly increasing")
		}
	}

	// Ensure no negative thrust values
	for _, point := range curve {
		if point[1] < 0 {
			panic("negative thrust values are invalid")
		}
	}

	return curve
}

func findMaxThrust(thrustData [][]float64) float64 {
	maxVal := 0.0
	for _, point := range thrustData {
		if point[1] > maxVal {
			maxVal = point[1]
		}
	}
	return maxVal
}
