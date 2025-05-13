package types_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestAcceleration(t *testing.T) {
	// Test creating a new acceleration with values
	a := &types.Acceleration{
		Vec: types.Vector3{X: 1.0, Y: 2.0, Z: 3.0},
	}

	assert.Equal(t, 1.0, a.Vec.X)
	assert.Equal(t, 2.0, a.Vec.Y)
	assert.Equal(t, 3.0, a.Vec.Z)

	// Test zero acceleration
	zero := &types.Acceleration{
		Vec: types.Vector3{},
	}

	assert.Equal(t, 0.0, zero.Vec.X)
	assert.Equal(t, 0.0, zero.Vec.Y)
	assert.Equal(t, 0.0, zero.Vec.Z)
}
