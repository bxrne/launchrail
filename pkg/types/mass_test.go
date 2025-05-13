package types_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestMass(t *testing.T) {
	// Test creating a new mass with value
	m := &types.Mass{
		Value: 10.5,
	}
	
	assert.Equal(t, 10.5, m.Value)
	
	// Test zero mass
	zero := &types.Mass{
		Value: 0,
	}
	
	assert.Equal(t, 0.0, zero.Value)
}
