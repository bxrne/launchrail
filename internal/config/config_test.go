package config

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// INFO: Helper to reset instance
func resetSingleton() {
	once = sync.Once{}
	instance = nil
	err = nil
}

// TEST: GIVEN a valid config file WHEN LoadConfig is called THEN it should load the config successfully
func TestLoadConfig(t *testing.T) {
	resetSingleton()

	cfg, err := LoadConfig("../../testdata/test_config.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "0.0.1", cfg.App.Version)
	assert.Equal(t, "GNU GPL v3", cfg.App.License)
	assert.Equal(t, "https://github.com/bxrne/launchrail", cfg.App.Repo)
	assert.Equal(t, "launchrail.log", cfg.Logs.File)
	assert.Equal(t, 10000000, cfg.Engine.TimeStepNS)
}

// TEST: GIVEN a non-existent config file WHEN LoadConfig is called THEN it should return an error
func TestLoadConfig_FileNotFound(t *testing.T) {
	resetSingleton()

	_, err := LoadConfig("non_existent_file.yaml")
	assert.Error(t, err)
}

// TEST: GIVEN an invalid config file WHEN LoadConfig is called THEN it should return an error
func TestLoadConfig_InvalidFormat(t *testing.T) {
	resetSingleton()

	_, err = LoadConfig("../../testdata/invalid_config.yaml")
	assert.Error(t, err)
}

// TEST: GIVEN a valid config file WHEN LoadConfig is called multiple times THEN it should return the same instance
func TestLoadConfig_Singleton(t *testing.T) {
	resetSingleton()

	cfg1, err := LoadConfig("../../testdata/test_config.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, cfg1)

	cfg2, err := LoadConfig("../../testdata/test_config.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, cfg2)

	assert.Equal(t, cfg1, cfg2)
}

// TEST: GIVEN a config file with invalid structure WHEN LoadConfig is called THEN it should return an error while unmarshalling
func TestLoadConfig_InvalidStructure(t *testing.T) {
	resetSingleton()

	_, err = LoadConfig("../../testdata/invalid_config.yaml")
	assert.Error(t, err)
}
