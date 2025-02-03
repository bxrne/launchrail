package barrowman

import (
	"sync"

	"github.com/bxrne/launchrail/pkg/components"
)

var (
	// Object pools for frequently allocated items
	vectorPool = sync.Pool{
		New: func() interface{} {
			return make([]float64, 3)
		},
	}
)

// CPCalculator handles center of pressure calculations using Barrowman method
type CPCalculator struct {
	mu sync.RWMutex
}

// CalculateCP calculates center of pressure using Barrowman method
func (c *CPCalculator) CalculateCP(nose *components.Nosecone, body *components.Bodytube, fins *components.TrapezoidFinset) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get vector from pool
	cpContrib := vectorPool.Get().([]float64)
	defer vectorPool.Put(cpContrib)

	// Nosecone contribution
	cpContrib[0] = c.calculateNoseCP(nose)

	// Body contribution
	cpContrib[1] = c.calculateBodyCP(body)

	// Fin contribution
	// if fins != nil {
	// 	cpContrib[2] = c.calculateFinCP(fins)
	// }

	// Weight contributions by area
	totalArea := nose.GetPlanformArea() + body.GetPlanformArea()

	cp := 0.0
	for _, contrib := range cpContrib {
		cp += contrib * (contrib / totalArea)
	}

	return cp
}

func (c *CPCalculator) calculateNoseCP(nose *components.Nosecone) float64 {
	// Von Karman ogive approximation
	return 0.466 * nose.Length
}

func (c *CPCalculator) calculateBodyCP(body *components.Bodytube) float64 {
	return body.Length / 2
}

func (c *CPCalculator) calculateFinCP(fins *components.TrapezoidFinset) float64 {
	// Calculate fin CP using Barrowman method
	// See https://www.apogeerockets.com/education/downloads/Newsletter281.pdf
	// for more details

	var cp float64

	// Calculate fin moment arm
	momentArm := fins.Position.X + fins.RootChord/3

	// Calculate fin CP
	cp = momentArm + fins.RootChord/3

	return cp

}

// NewCPCalculator creates a new CPCalculator
func NewCPCalculator() *CPCalculator {
	return &CPCalculator{}
}
