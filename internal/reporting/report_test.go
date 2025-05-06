package reporting

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bxrne/launchrail/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	gen, err := NewGenerator()
	require.NoError(t, err)
	assert.NotNil(t, gen)
	assert.NotNil(t, gen.template)
}

func TestLoadSimulationData(t *testing.T) {
	// 1. Setup Test RecordManager
	tempDir := t.TempDir()
	rm, err := storage.NewRecordManager(tempDir)
	require.NoError(t, err)

	// 2. Create a dummy record
	record, err := rm.CreateRecord()
	require.NoError(t, err)
	require.NotNil(t, record)
	recordHash := record.Hash
	// It's good practice to close the record resources, even in tests
	defer record.Close()

	// 3. Define report specific directory for this test
	reportSpecificDir := filepath.Join(tempDir, "test_report_pkg", recordHash)
	err = os.MkdirAll(reportSpecificDir, 0755)
	require.NoError(t, err)

	// 4. Call LoadSimulationData
	loadedData, err := LoadSimulationData(rm, recordHash, reportSpecificDir)
	require.NoError(t, err)

	// 5. Assertions on ReportData
	assert.Equal(t, recordHash, loadedData.RecordID)
	assert.Equal(t, "v0.0.0-dev", loadedData.Version) // Check hardcoded version

	assetSubDir := "assets"
	expectedAtmoPlotPath := filepath.Join(assetSubDir, "atmosphere_plot.png")
	expectedThrustPlotPath := filepath.Join(assetSubDir, "thrust_plot.png")
	expectedTrajectoryPlotPath := filepath.Join(assetSubDir, "trajectory_plot.png")
	expectedDynamicsPlotPath := filepath.Join(assetSubDir, "dynamics_plot.png")
	expectedGPSMapPath := filepath.Join(assetSubDir, "gps_map.png")

	assert.Equal(t, expectedAtmoPlotPath, loadedData.AtmospherePlotPath)
	assert.Equal(t, expectedThrustPlotPath, loadedData.ThrustPlotPath)
	assert.Equal(t, expectedTrajectoryPlotPath, loadedData.TrajectoryPlotPath)
	assert.Equal(t, expectedDynamicsPlotPath, loadedData.DynamicsPlotPath)
	assert.Equal(t, expectedGPSMapPath, loadedData.GPSMapImagePath)

	// 6. Assertions on created dummy assets
	assert.FileExists(t, filepath.Join(reportSpecificDir, expectedAtmoPlotPath))
	assert.FileExists(t, filepath.Join(reportSpecificDir, expectedThrustPlotPath))
	assert.FileExists(t, filepath.Join(reportSpecificDir, expectedTrajectoryPlotPath))
	assert.FileExists(t, filepath.Join(reportSpecificDir, expectedDynamicsPlotPath))
	assert.FileExists(t, filepath.Join(reportSpecificDir, expectedGPSMapPath))
}

func TestLoadSimulationData_NotFound(t *testing.T) {
	// 1. Setup Test RecordManager
	tempDir := t.TempDir()
	rm, err := storage.NewRecordManager(tempDir)
	require.NoError(t, err)

	// 2. Define report specific directory (even for a non-existent record)
	nonExistentHash := "this_hash_does_not_exist"
	reportSpecificDir := filepath.Join(tempDir, "test_report_pkg_notfound", nonExistentHash)
	// No need to create this dir as LoadSimulationData should fail before asset creation

	// 3. Attempt to load non-existent record
	_, err = LoadSimulationData(rm, nonExistentHash, reportSpecificDir)

	// 4. Assertions
	require.Error(t, err) // Expect an error
	assert.Contains(t, err.Error(), "failed to load record")
	assert.Contains(t, err.Error(), nonExistentHash)
}

func TestGenerateReportPackage(t *testing.T) {
	// 1. Setup Test RecordManager and base reports directory
	tempDir := t.TempDir()
	rm, err := storage.NewRecordManager(filepath.Join(tempDir, "records"))
	require.NoError(t, err)
	baseReportsDir := filepath.Join(tempDir, "reports")

	// 2. Create a dummy record
	record, err := rm.CreateRecord()
	require.NoError(t, err)
	recordHash := record.Hash
	defer record.Close()

	// 3. Call GenerateReportPackage
	generatedReportDir, err := GenerateReportPackage(rm, recordHash, baseReportsDir)
	require.NoError(t, err)

	// 4. Assertions on the generated package structure
	expectedReportDir := filepath.Join(baseReportsDir, recordHash)
	assert.Equal(t, expectedReportDir, generatedReportDir)
	assert.DirExists(t, generatedReportDir)

	mdFilePath := filepath.Join(generatedReportDir, "report.md")
	assert.FileExists(t, mdFilePath)

	assetsDir := filepath.Join(generatedReportDir, "assets")
	assert.DirExists(t, assetsDir)

	// Check for dummy asset files
	expectedAssetFiles := []string{
		"atmosphere_plot.png",
		"thrust_plot.png",
		"trajectory_plot.png",
		"dynamics_plot.png",
		"gps_map.png",
	}
	for _, assetFile := range expectedAssetFiles {
		assert.FileExists(t, filepath.Join(assetsDir, assetFile))
	}

	// 5. Assertions on report.md content (check for relative links)
	mdContentBytes, err := os.ReadFile(mdFilePath)
	require.NoError(t, err)
	mdContent := string(mdContentBytes)

	assert.Contains(t, mdContent, "# Simulation Report: "+recordHash)
	assert.Contains(t, mdContent, "Version: v0.0.0-dev")
	assert.Contains(t, mdContent, "![](assets/atmosphere_plot.png)")
	assert.Contains(t, mdContent, "![](assets/thrust_plot.png)")
	assert.Contains(t, mdContent, "![](assets/trajectory_plot.png)")
	assert.Contains(t, mdContent, "![](assets/dynamics_plot.png)")
	assert.Contains(t, mdContent, "![](assets/gps_map.png)")
}
