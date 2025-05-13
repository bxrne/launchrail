package types_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestPosition(t *testing.T) {
	// Test creating a new position with vector values
	p := &types.Position{
		Vec: types.Vector3{X: 10.0, Y: 20.0, Z: 30.0},
	}

	assert.Equal(t, 10.0, p.Vec.X)
	assert.Equal(t, 20.0, p.Vec.Y)
	assert.Equal(t, 30.0, p.Vec.Z)

	// Test zero position
	zero := &types.Position{
		Vec: types.Vector3{},
	}

	assert.Equal(t, 0.0, zero.Vec.X)
	assert.Equal(t, 0.0, zero.Vec.Y)
	assert.Equal(t, 0.0, zero.Vec.Z)
}
