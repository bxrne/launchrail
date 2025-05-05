package reporting

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/storage"
)

// ReportData holds all the necessary data for generating the report.
type ReportData struct {
	// TODO: Define fields for simulation results, atmospheric data, motor data, etc.
	Version string // Application version
	RecordID string
	// ... add other fields like plot paths/data
	AtmospherePlotPath string // Example placeholder
	ThrustPlotPath     string // Example placeholder
	TrajectoryPlotPath string // Example placeholder
	DynamicsPlotPath   string // Example placeholder
	GPSMapImagePath    string // Path to generated GPS map image
}

// LoadSimulationData loads the necessary data for a report from storage.
func LoadSimulationData(rm *storage.RecordManager, recordID string) (ReportData, error) {
	log := logger.GetLogger("") // Consider passing logger or config

	record, err := rm.GetRecord(recordID)
	if err != nil {
		log.Error("Failed to get record for report data", "recordID", recordID, "error", err)
		return ReportData{}, fmt.Errorf("failed to load record %s: %w", recordID, err)
	}

	// TODO: Load actual data from record.Motion, record.Events, record.Dynamics
	//       using record.Motion.ReadHeadersAndData(), etc.
	// TODO: Generate plots/maps using the loaded data and store paths.

	log.Info("Loaded record for report", "recordID", recordID, "creationTime", record.CreationTime)

	// Populate ReportData with basic info and placeholders
	data := ReportData{
		RecordID: record.Hash,
		// Version: Needs config access or to be passed in.
		// Plot Paths - Keep as placeholders for now:
		AtmospherePlotPath: "(Plot not generated)",
		ThrustPlotPath:     "(Plot not generated)",
		TrajectoryPlotPath: "(Plot not generated)",
		DynamicsPlotPath:   "(Plot not generated)",
		GPSMapImagePath:    "(Map not generated)",
	}

	return data, nil
}

// Generator handles report generation.
type Generator struct {
	template *template.Template
}

// NewGenerator creates a new report generator.
func NewGenerator(templateDir string) (*Generator, error) {
	tmplPath := filepath.Join(templateDir, "report.md.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse report template '%s': %w", tmplPath, err)
	}

	return &Generator{
		template: tmpl,
	}, nil
}

// GenerateMarkdown creates the markdown report content for the given data.
func (g *Generator) GenerateMarkdown(data ReportData) ([]byte, error) {
	var mdOutput bytes.Buffer
	if err := g.template.Execute(&mdOutput, data); err != nil {
		return nil, fmt.Errorf("failed to execute markdown template: %w", err)
	}
	return mdOutput.Bytes(), nil
}

// GeneratePDF generates the final PDF report.
func (g *Generator) GeneratePDF(data ReportData) ([]byte, error) {
	// 2. Generate Plots (Needs implementation - create temporary image files)
	// data.AtmospherePlotPath, err = generateAtmospherePlot(data) ...
	// data.ThrustPlotPath, err = generateThrustPlot(data) ...
	// ... etc for other plots
	// Remember to clean up temporary plot files later.

	// 3. Generate Markdown from template
	mdBytes, err := g.GenerateMarkdown(data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate markdown content: %w", err)
	}

	// 4. Convert Markdown to PDF (Needs implementation)
	pdfBytes, err := convertMarkdownToPDF(mdBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert markdown to PDF: %w", err)
	}

	// TODO: Clean up temporary plot files if they were created.

	return pdfBytes, nil
}

// convertMarkdownToPDF placeholder function
func convertMarkdownToPDF(md []byte) ([]byte, error) {
	// TODO: Implement actual Markdown to PDF conversion
	// This might involve using a library like gofpdf + a markdown parser,
	// or shelling out to pandoc or wkhtmltopdf.
	// For now, just return the markdown wrapped in a placeholder message.
	pdfContent := fmt.Sprintf("--- PDF Conversion Placeholder ---\n\n%s\n\n--- End Placeholder ---", string(md))
	return []byte(pdfContent), nil
	// return nil, fmt.Errorf("PDF conversion not yet implemented")
}

// TODO: Add placeholder functions for plot generation
// func generateAtmospherePlot(data ReportData) (string, error) { ... return plotFilePath, nil }
// func generateThrustPlot(data ReportData) (string, error) { ... return plotFilePath, nil }
// ... etc ...
