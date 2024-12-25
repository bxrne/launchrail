package thrustcurves

import (
	"github.com/bxrne/launchrail/internal/logger"
)

// NOTE: Assemble motor data from the ThrustCurve API.
func Load(designation string) *MotorData {
	log := logger.GetLogger()

	valid, err := validateDesignation(designation)
	if !valid {
		log.Fatal("Invalid motor designation", "designation", designation)
	}

	specification, err := designationToSpecification(designation)
	if err != nil {
		log.Fatal("Failed to parse motor designation", "error", err)
	}

	id, err := getMotorID(designation)
	if err != nil {
		log.Fatal("Failed to get motor ID", "error", err)
	}

	curve, err := getMotorCurve(id)
	if err != nil {
		log.Fatal("Failed to get motor curve", "error", err)
	}

	motor := &MotorData{
		Designation:   designation,
		ID:            id,
		Thrust:        curve,
		Specification: specification,
	}

	log.Info("Motor data loaded", "description", motor.String())

	return motor

}
