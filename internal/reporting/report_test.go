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
	tempDir := t.TempDir()
	templateDir := filepath.Join(tempDir, "templates")

	err := os.Mkdir(templateDir, 0755)
	require.NoError(t, err)

	templateContent := "Report for {{.RecordID}} Version: {{.Version}}"
	templatePath := filepath.Join(templateDir, "report.md.tmpl")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	gen, err := NewGenerator(templateDir)
	require.NoError(t, err)
	assert.NotNil(t, gen)
	assert.NotNil(t, gen.template)
}

func TestGenerateMarkdown(t *testing.T) {
	tempDir := t.TempDir()
	templateDir := filepath.Join(tempDir, "templates")

	err := os.Mkdir(templateDir, 0755)
	require.NoError(t, err)

	templateContent := "# Report for {{.RecordID}}\nVersion: {{.Version}}"
	templatePath := filepath.Join(templateDir, "report.md.tmpl")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	gen, err := NewGenerator(templateDir)
	require.NoError(t, err)

	data := ReportData{
		RecordID: "test1234",
		Version:  "v1.0.0",
	}

	mdBytes, err := gen.GenerateMarkdown(data)
	require.NoError(t, err)
	assert.NotEmpty(t, mdBytes)

	contentString := string(mdBytes)
	assert.Contains(t, contentString, "# Report for test1234")
	assert.Contains(t, contentString, "Version: v1.0.0")
}

func TestGeneratePDF_Placeholder(t *testing.T) {
	tempDir := t.TempDir()
	templateDir := filepath.Join(tempDir, "templates")

	err := os.Mkdir(templateDir, 0755)
	require.NoError(t, err)

	templateContent := "# Report for {{.RecordID}}"
	templatePath := filepath.Join(templateDir, "report.md.tmpl")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	gen, err := NewGenerator(templateDir)
	require.NoError(t, err)

	data := ReportData{
		RecordID: "pdf_test",
		Version:  "v1.1.0",
	}

	pdfBytes, err := gen.GeneratePDF(data)
	require.NoError(t, err)
	assert.NotEmpty(t, pdfBytes)

	contentString := string(pdfBytes)
	assert.Contains(t, contentString, "--- PDF Conversion Placeholder ---")
	assert.Contains(t, contentString, "# Report for pdf_test")
	assert.Contains(t, contentString, "--- End Placeholder ---")
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

	// 3. Call LoadSimulationData
	loadedData, err := LoadSimulationData(rm, recordHash)
	require.NoError(t, err)

	// 4. Assertions
	assert.Equal(t, recordHash, loadedData.RecordID)
	// Version is set in the handler, so it will be empty here initially
	assert.Empty(t, loadedData.Version)
	// Check placeholder paths
	assert.Equal(t, "(Plot not generated)", loadedData.AtmospherePlotPath)
	assert.Equal(t, "(Plot not generated)", loadedData.ThrustPlotPath)
	assert.Equal(t, "(Plot not generated)", loadedData.TrajectoryPlotPath)
	assert.Equal(t, "(Plot not generated)", loadedData.DynamicsPlotPath)
	assert.Equal(t, "(Map not generated)", loadedData.GPSMapImagePath)
}

func TestLoadSimulationData_NotFound(t *testing.T) {
	// 1. Setup Test RecordManager
	tempDir := t.TempDir()
	rm, err := storage.NewRecordManager(tempDir)
	require.NoError(t, err)

	// 2. Attempt to load non-existent record
	nonExistentHash := "this_hash_does_not_exist"
	_, err = LoadSimulationData(rm, nonExistentHash)

	// 3. Assertions
	require.Error(t, err) // Expect an error
	assert.Contains(t, err.Error(), "failed to load record")
	assert.Contains(t, err.Error(), nonExistentHash)
}

// TODO: Add tests for PDF generation once implemented
// TODO: Add tests for plot generation placeholders/mocks
