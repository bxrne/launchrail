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

	// Calculate individual CP locations
	noseCP := c.calculateNoseCP(nose)
	bodyCP := c.calculateBodyCP(body)
	finCP := c.calculateFinCP(fins)

	// Calculate areas
	noseArea := nose.GetPlanformArea()
	bodyArea := body.GetPlanformArea()
	finArea := fins.GetPlanformArea()
	totalArea := noseArea + bodyArea + finArea

	if totalArea <= 0 {
		return 0
	}

	// Weight CP contributions by their respective areas
	cp := (noseCP*noseArea + bodyCP*bodyArea + finCP*finArea) / totalArea

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
	// Rough approximation: place fin CP at 0.75 of root chord
	return 0.75 * fins.RootChord
}

// NewCPCalculator creates a new CPCalculator
func NewCPCalculator() *CPCalculator {
	return &CPCalculator{}
}
