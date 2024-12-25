package thrustcurves

import (
	"fmt"
)

// NOTE: Assemble motor data from the ThrustCurve API.
func Load(designation string) (*MotorData, error) {
	valid, err := validateDesignation(designation)
	if !valid {
		return nil, fmt.Errorf("invalid motor designation: %s", designation)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to validate motor designation: %s", err)
	}

	specification, err := designationToSpecification(designation)
	if err != nil {
		return nil, fmt.Errorf("failed to get motor specification: %s", err)
	}

	id, err := getMotorID(designation)
	if err != nil {
		return nil, fmt.Errorf("failed to get motor ID: %s", err)
	}

	curve, err := getMotorCurve(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get motor curve: %s", err)
	}

	return &MotorData{
		Designation:   designation,
		ID:            id,
		Thrust:        curve,
		Specification: specification,
	}, nil
}
