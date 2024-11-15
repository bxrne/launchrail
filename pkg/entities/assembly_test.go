package entities_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/entities"
	"github.com/bxrne/launchrail/pkg/ork"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN a valid orkConfig and thrustCurvePath WHEN NewAssembly is called THEN it should return a valid Assembly
func TestNewAssembly_ValidInput(t *testing.T) {
	orkConfig, err := ork.Decompress("../../testdata/ULR3.ork")
	assert.NoError(t, err)

	thrustCurvePath := "../../testdata/cesaroni-l645.eng"

	assembly, err := entities.NewAssembly(*orkConfig, thrustCurvePath)
	assert.NoError(t, err)
	assert.NotNil(t, assembly)
	assert.Equal(t, "ULR3", assembly.Rocket.Name)
	assert.Equal(t, "Daire ", assembly.Rocket.Designer)
	assert.NotNil(t, assembly.Rocket.Motor)
}

// TEST: GIVEN an invalid thrustCurvePath WHEN NewAssembly is called THEN it should return an error
func TestNewAssembly_InvalidThrustCurvePath(t *testing.T) {
	orkConfig, err := ork.Decompress("../../testdata/ULR3.ork")
	assert.NoError(t, err)

	thrustCurvePath := "../../testdata/cesaroni-NOEXIST-l645.eng"

	assembly, err := entities.NewAssembly(*orkConfig, thrustCurvePath)
	assert.Error(t, err)
	assert.Nil(t, assembly)
}

// TEST: GIVEN a valid Assembly WHEN Info is called THEN it should return the correct information string
func TestAssembly_Info(t *testing.T) {
	orkConfig, err := ork.Decompress("../../testdata/ULR3.ork")
	assert.NoError(t, err)

	thrustCurvePath := "../../testdata/cesaroni-l645.eng"

	assembly, err := entities.NewAssembly(*orkConfig, thrustCurvePath)
	assert.NoError(t, err)
	assert.NotNil(t, assembly)

	info := assembly.Info()
	assert.Equal(t, "ULR3 by Daire ", info)
}
