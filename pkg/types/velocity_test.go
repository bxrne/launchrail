package types_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestVelocity(t *testing.T) {
	// Test creating a new velocity with vector values
	v := &types.Velocity{
		Vec: types.Vector3{X: 15.0, Y: 25.0, Z: 35.0},
	}
	
	assert.Equal(t, 15.0, v.Vec.X)
	assert.Equal(t, 25.0, v.Vec.Y)
	assert.Equal(t, 35.0, v.Vec.Z)
	
	// Test zero velocity
	zero := &types.Velocity{
		Vec: types.Vector3{},
	}
	
	assert.Equal(t, 0.0, zero.Vec.X)
	assert.Equal(t, 0.0, zero.Vec.Y)
	assert.Equal(t, 0.0, zero.Vec.Z)
}
