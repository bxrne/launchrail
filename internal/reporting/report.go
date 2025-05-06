package reporting

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/storage"
)

// ReportData holds all the necessary data for generating the report.
type ReportData struct {
	Version            string // Application version
	RecordID           string
	AtmospherePlotPath string
	ThrustPlotPath     string
	TrajectoryPlotPath string
	DynamicsPlotPath   string
	GPSMapImagePath    string
}

const defaultReportTemplate = `
# Simulation Report: {{.RecordID}}

Version: {{.Version}}

## Plots & Data

- Atmosphere: ![]({{.AtmospherePlotPath}})
- Thrust Curve: ![]({{.ThrustPlotPath}})
- Trajectory: ![]({{.TrajectoryPlotPath}})
- Dynamics: ![]({{.DynamicsPlotPath}})
- GPS Map: ![]({{.GPSMapImagePath}})

<!-- Add more sections for tables, etc. -->
`

// createDummyAsset creates an empty file at the given path.
// In a real scenario, this would generate actual plot images.
func createDummyAsset(assetPath string) error {
	assetDir := filepath.Dir(assetPath)
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		return fmt.Errorf("failed to create asset directory %s: %w", assetDir, err)
	}
	f, err := os.Create(assetPath)
	if err != nil {
		return fmt.Errorf("failed to create dummy asset %s: %w", assetPath, err)
	}
	return f.Close()
}

// LoadSimulationData loads the necessary data for a report from storage and creates dummy assets.
func LoadSimulationData(rm *storage.RecordManager, recordID string, reportSpecificDir string) (ReportData, error) {
	log := logger.GetLogger("")

	record, err := rm.GetRecord(recordID)
	if err != nil {
		log.Error("Failed to get record for report data", "recordID", recordID, "error", err)
		return ReportData{}, fmt.Errorf("failed to load record %s: %w", recordID, err)
	}

	log.Info("Loaded record for report", "recordID", recordID, "creationTime", record.CreationTime)

	// Define relative paths for assets
	assetSubDir := "assets"
	atmoPlotRelPath := filepath.Join(assetSubDir, "atmosphere_plot.png")
	thrustPlotRelPath := filepath.Join(assetSubDir, "thrust_plot.png")
	trajectoryPlotRelPath := filepath.Join(assetSubDir, "trajectory_plot.png")
	dynamicsPlotRelPath := filepath.Join(assetSubDir, "dynamics_plot.png")
	gpsMapRelPath := filepath.Join(assetSubDir, "gps_map.png")

	// Create dummy assets in the report-specific directory
	if err := createDummyAsset(filepath.Join(reportSpecificDir, atmoPlotRelPath)); err != nil {
		return ReportData{}, err
	}
	if err := createDummyAsset(filepath.Join(reportSpecificDir, thrustPlotRelPath)); err != nil {
		return ReportData{}, err
	}
	if err := createDummyAsset(filepath.Join(reportSpecificDir, trajectoryPlotRelPath)); err != nil {
		return ReportData{}, err
	}
	if err := createDummyAsset(filepath.Join(reportSpecificDir, dynamicsPlotRelPath)); err != nil {
		return ReportData{}, err
	}
	if err := createDummyAsset(filepath.Join(reportSpecificDir, gpsMapRelPath)); err != nil {
		return ReportData{}, err
	}

	data := ReportData{
		RecordID:           record.Hash,
		// Version: Needs config or to be passed in. For now, hardcode or leave empty.
		Version:            "v0.0.0-dev",
		AtmospherePlotPath: atmoPlotRelPath,
		ThrustPlotPath:     thrustPlotRelPath,
		TrajectoryPlotPath: trajectoryPlotRelPath,
		DynamicsPlotPath:   dynamicsPlotRelPath,
		GPSMapImagePath:    gpsMapRelPath,
	}

	return data, nil
}

// Generator handles report generation.
type Generator struct {
	template *template.Template
}

// NewGenerator creates a new report generator.
func NewGenerator() (*Generator, error) {
	tmpl, err := template.New("report").Parse(defaultReportTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse default markdown template: %w", err)
	}
	return &Generator{
		template: tmpl,
	}, nil
}

// GenerateMarkdownFile creates the markdown report content and saves it to a file.
func (g *Generator) GenerateMarkdownFile(data ReportData, outputDir string) error {
	var mdOutput bytes.Buffer
	if err := g.template.Execute(&mdOutput, data); err != nil {
		return fmt.Errorf("failed to execute markdown template: %w", err)
	}

	mdFilePath := filepath.Join(outputDir, "report.md")
	if err := os.WriteFile(mdFilePath, mdOutput.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write markdown report to %s: %w", mdFilePath, err)
	}
	return nil
}

// GenerateReportPackage orchestrates the generation of a self-contained report package.
func GenerateReportPackage(rm *storage.RecordManager, recordID string, baseReportsDir string) (string, error) {
	log := logger.GetLogger("")
	reportSpecificDir := filepath.Join(baseReportsDir, recordID)

	if err := os.MkdirAll(reportSpecificDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create report directory %s: %w", reportSpecificDir, err)
	}

	data, err := LoadSimulationData(rm, recordID, reportSpecificDir)
	if err != nil {
		return "", fmt.Errorf("failed to load simulation data for report: %w", err)
	}

	gen, err := NewGenerator()
	if err != nil {
		return "", fmt.Errorf("failed to create report generator: %w", err)
	}

	if err := gen.GenerateMarkdownFile(data, reportSpecificDir); err != nil {
		return "", fmt.Errorf("failed to generate markdown report file: %w", err)
	}

	log.Info("Successfully generated report package", "recordID", recordID, "outputDir", reportSpecificDir)
	return reportSpecificDir, nil
}
