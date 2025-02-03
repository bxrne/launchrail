package barrowman

import (
	"sync"

	"github.com/bxrne/launchrail/pkg/components"
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

// calculateNoseCP calculates the CP of a nosecone
func (c *CPCalculator) calculateNoseCP(nose *components.Nosecone) float64 {
	// Von Karman ogive approximation
	return 0.466 * nose.Length
}

// calculateBodyCP calculates the CP of a bodytube
func (c *CPCalculator) calculateBodyCP(body *components.Bodytube) float64 {
	return body.Length / 2
}

// calculateFinCP calculates the CP of a finset
func (c *CPCalculator) calculateFinCP(fins *components.TrapezoidFinset) float64 {
	// Rough approximation: place fin CP at 0.75 of root chord
	return 0.75 * fins.RootChord
}

// NewCPCalculator creates a new CPCalculator
func NewCPCalculator() *CPCalculator {
	return &CPCalculator{}
}
