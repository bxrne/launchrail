package types_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestOrientation(t *testing.T) {
	// Test creating a new orientation with quaternion values
	q := types.Quaternion{W: 1.0, X: 0.0, Y: 0.0, Z: 0.0}
	o := &types.Orientation{
		Quat: q,
	}
	
	assert.Equal(t, q, o.Quat)
	
	// Test with non-identity quaternion
	q2 := types.Quaternion{W: 0.7071, X: 0.7071, Y: 0.0, Z: 0.0}
	o2 := &types.Orientation{
		Quat: q2,
	}
	
	assert.Equal(t, q2, o2.Quat)
}
