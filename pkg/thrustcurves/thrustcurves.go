package thrustcurves

import (
	"fmt"

	"github.com/bxrne/launchrail/pkg/designation"
)

// NOTE: Assemble motor data from the ThrustCurve API.
func Load(raw_designation string) (*MotorData, error) {
	designation, err := designation.New(raw_designation)
	if err != nil {
		return nil, fmt.Errorf("failed to create motor designation: %s", err)
	}

	valid, err := designation.Validate()
	if !valid {
		return nil, fmt.Errorf("invalid motor designation: %s", designation)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to validate motor designation: %s", err)
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
		Designation: designation,
		ID:          id,
		Thrust:      curve,
	}, nil
}
