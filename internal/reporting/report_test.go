package reporting

import (
	"os"
	"path/filepath"
	"testing"

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

// TODO: Add tests for PDF generation once implemented
// TODO: Add tests for LoadSimulationData once implemented
// TODO: Add tests for plot generation placeholders/mocks
