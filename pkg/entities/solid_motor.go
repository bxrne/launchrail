package entities

import (
	"bufio"
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
	Manufacturer  string
	Designation   string
	Diameter      float64
	Length        float64
	TotalImpulse  float64
	Propellant    string
	AverageThrust float64
	MaxThrust     float64
	BurnTime      time.Duration
	ThrustCurve   []ThrustPoint
}

func NewSolidMotor(filePath string) (*SolidMotor, error) {
	file, err := os.Open(filePath)
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

	return &motor, nil
}
