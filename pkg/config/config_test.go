package config

import (
	"os"
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

	configContent := `
app:
  version: "0.0.1"
  license: "GNU GPL v3"
  repo: "https://github.com/bxrne/launchrail"
logs:
  file: "launchrail.log"
`
	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(configContent))
	assert.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "0.0.1", cfg.App.Version)
	assert.Equal(t, "GNU GPL v3", cfg.App.License)
	assert.Equal(t, "https://github.com/bxrne/launchrail", cfg.App.Repo)
	assert.Equal(t, "launchrail.log", cfg.Logs.File)
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

	invalidConfigContent := `
invalid_yaml_content
`
	tmpFile, err := os.CreateTemp("", "config_test_invalid_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(invalidConfigContent))
	assert.NoError(t, err)
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	assert.Error(t, err)
}

// TEST: GIVEN a valid config file WHEN LoadConfig is called multiple times THEN it should return the same instance
func TestLoadConfig_Singleton(t *testing.T) {
	resetSingleton()

	configContent := `
app:
  version: "0.0.1"
  license: "GNU GPL v3"
  repo: "https://github.com/bxrne/launchrail"
logs:
  file: "launchrail.log"
`
	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(configContent))
	assert.NoError(t, err)
	tmpFile.Close()

	cfg1, err := LoadConfig(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, cfg1)

	cfg2, err := LoadConfig(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, cfg2)

	assert.Equal(t, cfg1, cfg2)
}

// TEST: GIVEN an invalid config file WHEN LoadConfig is called THEN it should return an error and nil instance
func TestLoadConfig_Invalid(t *testing.T) {
	resetSingleton()

	configContent := `
app
  version: "0.0.1"
  license: "GNU GPL v3"
  repo: "https://github.com/bxrne/launchrail"
logs:
  file: "launchrail.log"
`
	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(configContent))
	assert.NoError(t, err)
	tmpFile.Close()

	cfg1, err := LoadConfig(tmpFile.Name())
	assert.Error(t, err)
	assert.Nil(t, cfg1)
}
