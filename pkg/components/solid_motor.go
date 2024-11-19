package components

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type ThrustPoint struct {
	Time   time.Duration
	Thrust float64
}

type SolidMotor struct {
	Manufacturer string
	Designation  string

	DryMass        float64
	PropellantMass float64

	// parsed
	Diameter      float64
	Length        float64
	TotalImpulse  float64
	Propellant    string
	AverageThrust float64
	MaxThrust     float64
	BurnTime      time.Duration
	ThrustCurve   []ThrustPoint
}

func NewSolidMotor(thrustCurveFilePath string, dryMass float64, propellantMass float64) (*SolidMotor, error) {
	file, err := os.Open(thrustCurveFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var motor SolidMotor

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
	}

	if dryMass <= 0 {
		return nil, fmt.Errorf("Dry mass must be greater than 0")
	}
	motor.DryMass = dryMass

	if propellantMass <= 0 {
		return nil, fmt.Errorf("Propellant mass must be greater than 0")
	}
	motor.PropellantMass = propellantMass

	return &motor, nil
}

func (m *SolidMotor) String() string {
	return fmt.Sprintf("Manufacturer: %s\nDesignation: %s\nDiameter: %.2f mm\nLength: %.2f mm\nPropellant: %s\nTotal Impulse: %.2f Ns\nAverage Thrust: %.2f N\nBurn Time: %s\n",
		m.Manufacturer, m.Designation, m.Diameter, m.Length, m.Propellant, m.TotalImpulse, m.AverageThrust, m.BurnTime)
}
