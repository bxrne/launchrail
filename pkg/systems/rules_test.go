package systems_test

import (
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/states"
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN Nothing WHEN NewRulesSystem is called THEN a new rules system is returned
func TestNewRulesSystem(t *testing.T) {
	rs := systems.NewRulesSystem(&ecs.World{}, &config.Engine{})
	assert.NotNil(t, rs)
}

// TEST: GIVEN a rules system WHEN Add is called with an entities state THEN its stored
func TestAdd(t *testing.T) {
	rs := systems.NewRulesSystem(&ecs.World{}, &config.Engine{})
	en := &states.PhysicsState{}

	rs.Add(en)
}
