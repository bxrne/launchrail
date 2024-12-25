package thrustcurves

import (
	"fmt"
	"strings"
)

type MotorData struct {
	Designation   string
	ID            string
	Thrust        [][]float64
	Specification *MotorSpecification
}

func (m *MotorData) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Motor: %s, ", m.Designation))
	sb.WriteString(fmt.Sprintf("Total Impulse: %.2f Ns, ", m.Specification.TotalImpulse))
	sb.WriteString(fmt.Sprintf("Class: %s, ", m.Specification.Class))
	sb.WriteString(fmt.Sprintf("Average Thrust: %.2f N, ", m.Specification.AverageThrust))
	sb.WriteString(fmt.Sprintf("Delay Time: %.2f s, ", m.Specification.DelayTime))
	sb.WriteString(fmt.Sprintf("Variant: %s", m.Specification.Variant))

	return sb.String()
}
