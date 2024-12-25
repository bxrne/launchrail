package thrustcurves

import (
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

var (
	designationSchema = `^(\d+)([A-Z]+)(\d+)-(\d+)([A-Z]+)$`
)

type MotorSpecification struct {
	TotalImpulse  float64
	Class         string
	AverageThrust float64
	DelayTime     float64
	Variant       string
}

func designationToSpecification(designation string) (*MotorSpecification, error) {
	// INFO: Designation String can be broken down into the following fields:
	// TotalImpulse-Class-AverageThrust-DelayTime-Variant (e.g. "269H110-14A")
	var totalImpulse float64
	var class string
	var averageThrust float64
	var delayTime float64
	var variant string

	exp := regexp.MustCompile(designationSchema)
	matches := exp.FindStringSubmatch(designation)
	if len(matches) != 6 {
		return nil, errors.New("failed to parse designation")
	}

	totalImpulse, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse total impulse")
	}

	class = matches[2]
	averageThrust, err = strconv.ParseFloat(matches[3], 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse average thrust")
	}

	delayTime, err = strconv.ParseFloat(matches[4], 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse delay time")
	}

	variant = matches[5]

	return &MotorSpecification{
		TotalImpulse:  totalImpulse,
		Class:         class,
		AverageThrust: averageThrust,
		DelayTime:     delayTime,
		Variant:       variant,
	}, nil
}

func validateDesignation(designation string) (bool, error) {
	// NOTE: TotalImpulse-Class-AverageThrust-DelayTime-Variant (e.g. "269H110-14A" is a valid designation)
	exp := regexp.MustCompile(designationSchema)
	if !exp.MatchString(designation) {
		return false, nil
	}

	return true, nil
}
