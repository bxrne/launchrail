package components

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type ThrustPoint struct {
	Time   time.Duration
	Thrust float64
}

type MotorGrain struct {
	InitialLength float64
	CurrentLength float64
	Diameter      float64
	BurnRate      float64
}

type SolidMotorState struct {
	ElapsedTime      time.Duration
	CurrentThrust    float64
	RemainingMass    float64
	GrainLengths     []float64
	BurnedPropellant float64
}

type SolidMotor struct {
	Manufacturer          string
	Designation           string
	DryMass               float64
	InitialPropellantMass float64

	Diameter      float64
	Length        float64
	TotalImpulse  float64
	Propellant    string
	AverageThrust float64
	MaxThrust     float64
	BurnTime      time.Duration
	ThrustCurve   []ThrustPoint

	Grains       []MotorGrain
	CurrentState SolidMotorState
}

func NewSolidMotor(thrustCurveFilePath string, dryMass float64, propellantMass float64, numberOfGrains int) (*SolidMotor, error) {
	file, err := os.Open(thrustCurveFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var motor SolidMotor
	motor.Grains = make([]MotorGrain, numberOfGrains)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ";") {
			continue // Skip comments
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if len(fields) == 7 {
			motor.Designation = fields[0]
			motor.Diameter, _ = strconv.ParseFloat(fields[1], 64)
			motor.Length, _ = strconv.ParseFloat(fields[2], 64)
			motor.Propellant = fields[3]
			motor.TotalImpulse, _ = strconv.ParseFloat(fields[4], 64)
			motor.AverageThrust, _ = strconv.ParseFloat(fields[5], 64)
			motor.Manufacturer = fields[6]

			// Initialize grains equally
			for i := 0; i < numberOfGrains; i++ {
				motor.Grains[i] = MotorGrain{
					InitialLength: motor.Length / float64(numberOfGrains),
					CurrentLength: motor.Length / float64(numberOfGrains),
					Diameter:      motor.Diameter,
					BurnRate:      0.1, // Example burn rate, adjust based on propellant
				}
			}
		} else if len(fields) == 2 {
			timeSeconds, err1 := strconv.ParseFloat(fields[0], 64)
			thrust, err2 := strconv.ParseFloat(fields[1], 64)
			if err1 == nil && err2 == nil {
				timeDuration := time.Duration(timeSeconds * float64(time.Second))
				motor.ThrustCurve = append(motor.ThrustCurve, ThrustPoint{Time: timeDuration, Thrust: thrust})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(motor.ThrustCurve) > 0 {
		motor.BurnTime = motor.ThrustCurve[len(motor.ThrustCurve)-1].Time
		motor.MaxThrust = motor.ThrustCurve[0].Thrust
		for _, tp := range motor.ThrustCurve {
			if tp.Thrust > motor.MaxThrust {
				motor.MaxThrust = tp.Thrust
			}
		}
	}

	if dryMass <= 0 {
		return nil, fmt.Errorf("dry mass must be greater than 0")
	}
	motor.DryMass = dryMass

	if propellantMass <= 0 {
		return nil, fmt.Errorf("propellant mass must be greater than 0")
	}
	motor.InitialPropellantMass = propellantMass

	motor.CurrentState = SolidMotorState{
		ElapsedTime:      0,
		CurrentThrust:    0,
		RemainingMass:    propellantMass,
		GrainLengths:     make([]float64, numberOfGrains),
		BurnedPropellant: 0,
	}
	for i := range motor.CurrentState.GrainLengths {
		motor.CurrentState.GrainLengths[i] = motor.Grains[i].InitialLength
	}

	return &motor, nil
}

func (m *SolidMotor) UpdateState(timeStep time.Duration) error {
	if timeStep <= 0 {
		return fmt.Errorf("time step must be positive")
	}

	// Find the current thrust from the thrust curve
	currentThrust := 0.0
	for _, tp := range m.ThrustCurve {
		if tp.Time <= m.CurrentState.ElapsedTime+timeStep {
			currentThrust = tp.Thrust
		} else {
			break
		}
	}

	// Update grain lengths and burned propellant
	burnedPropellantThisStep := 0.0
	for i := range m.Grains {
		if m.Grains[i].CurrentLength > 0 {
			burnLength := m.Grains[i].BurnRate * timeStep.Seconds()
			m.Grains[i].CurrentLength = math.Max(0, m.Grains[i].CurrentLength-burnLength)
			m.CurrentState.GrainLengths[i] = m.Grains[i].CurrentLength
			burnedPropellantThisStep += burnLength * math.Pi *
				(m.Grains[i].Diameter / 2) * (m.Grains[i].Diameter / 2)
		}
	}

	// Update motor state
	m.CurrentState.ElapsedTime += timeStep
	m.CurrentState.CurrentThrust = currentThrust
	m.CurrentState.RemainingMass = math.Max(0, m.CurrentState.RemainingMass-burnedPropellantThisStep)
	m.CurrentState.BurnedPropellant += burnedPropellantThisStep

	return nil
}

func (m *SolidMotor) String() string {
	return fmt.Sprintf(
		"Manufacturer: %s\n"+
			"Designation: %s\n"+
			"Diameter: %.2f mm\n"+
			"Length: %.2f mm\n"+
			"Propellant: %s\n"+
			"Total Impulse: %.2f Ns\n"+
			"Average Thrust: %.2f N\n"+
			"Burn Time: %s\n"+
			"Current State:\n"+
			"  Elapsed Time: %s\n"+
			"  Current Thrust: %.2f N\n"+
			"  Remaining Mass: %.2f kg\n"+
			"  Burned Propellant: %.2f kg",
		m.Manufacturer, m.Designation, m.Diameter, m.Length,
		m.Propellant, m.TotalImpulse, m.AverageThrust, m.BurnTime,
		m.CurrentState.ElapsedTime, m.CurrentState.CurrentThrust,
		m.CurrentState.RemainingMass, m.CurrentState.BurnedPropellant,
	)
}
