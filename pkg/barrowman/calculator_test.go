package barrowman_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bxrne/launchrail/pkg/barrowman"
	"github.com/bxrne/launchrail/pkg/components"
)

// TEST: GIVEN a new CP calculator WHEN NewCPCalculator is called THEN a new CPCalculator is returned
func TestNewCPCalculator(t *testing.T) {
	cpCalc := barrowman.NewCPCalculator()
	require.NotNil(t, cpCalc)
}

// TEST: GIVEN a nosecone, bodytube, and finset WHEN CalculateCP is called THEN the CP is calculated
func TestCalculateCP(t *testing.T) {
	cpCalc := barrowman.NewCPCalculator()
	nose := &components.Nosecone{Length: 1.0}
	body := &components.Bodytube{Length: 2.0}
	fins := &components.TrapezoidFinset{RootChord: 1.0, TipChord: 0.5, Span: 0.5}

	require.NotNil(t, nose)
	require.NotNil(t, body)
	require.NotNil(t, fins)
	assert.NotZero(t, nose.Length, "Nose length must not be zero")
	assert.NotZero(t, body.Length, "Body length must not be zero")

	noseArea := nose.GetPlanformArea()
	bodyArea := body.GetPlanformArea()
	finArea := fins.GetPlanformArea()
	totalArea := noseArea + bodyArea + finArea
	require.NotZero(t, totalArea, "Total area must not be zero")

	noseCP := 0.466 * nose.Length
	bodyCP := body.Length / 2
	finCP := 0.75 * fins.RootChord
	expectedCP := (noseCP*noseArea + bodyCP*bodyArea + finCP*finArea) / totalArea

	actualCP := cpCalc.CalculateCP(nose, body, fins)

	assert.False(t, math.IsNaN(actualCP), "CP calculation resulted in NaN")
	assert.InEpsilon(t, expectedCP, actualCP, 1e-6, "Overall CP mismatch")
}
