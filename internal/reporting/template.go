package reporting

import (
	"bytes"
	"fmt"
	"html/template"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/zerodha/logf"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// Constants for plot labels and error messages
const (
	// Axis labels
	LabelTimeSeconds = "Time (s)"

	// Error messages
	ErrCreateLinePlot = "failed to create new line plot: %w"
	ErrSavePlot       = "failed to save plot: %w"
)

// TemplateRenderer handles report template processing and rendering
type TemplateRenderer struct {
	log       *logf.Logger
	templates *template.Template
	assetsDir string
	// reportTemplate *template.Template // This specific field might be less relevant if always looking up by name
}

// NewTemplateRenderer creates a new template renderer with the specified templates directory
func NewTemplateRenderer(log *logf.Logger, templatesDir, assetsDir string) (*TemplateRenderer, error) {
	if log == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Make sure the templates directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("templates directory does not exist: %w", err)
	}

	// Make sure the assets directory exists
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		// Create it if it doesn't exist
		if err := os.MkdirAll(assetsDir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create assets directory: %w", err)
		}
	}

	// Define custom template functions
	funcMap := template.FuncMap{
		"embedSVG": func(plotFileName string, altText string) (template.HTML, error) {
			if plotFileName == "" {
				log.Warn("embedSVG called with empty plotFileName")
				return "<!-- embedSVG: plotFileName was empty -->", nil
			}
			// assetsDir is the absolute path to the specific report's assets directory
			absolutePlotPath := filepath.Join(assetsDir, plotFileName)
			log.Debug("embedSVG trying to read", "path", absolutePlotPath, "inputName", plotFileName, "reportAssetsDir", assetsDir)

			content, err := os.ReadFile(absolutePlotPath)
			if err != nil {
				log.Error("embedSVG failed to read file", "path", absolutePlotPath, "error", err)
				// Return a visible error in the report instead of empty or broken image
				return template.HTML(fmt.Sprintf("<div style='color:red; border:1px solid red; padding:10px;'>Error embedding SVG '%s': %s (path: %s)</div>", altText, err.Error(), plotFileName)), nil
			}
			return template.HTML(content), nil
		},
		"formatFloat": func(value float64, precision int) string {
			return fmt.Sprintf(fmt.Sprintf("%%.%df", precision), value)
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"replace": func(input, from, to string) string {
			return strings.ReplaceAll(input, from, to)
		},
		"title": func(input string) string {
			return cases.Title(language.English).String(input)
		},
		// Add other general-purpose functions if needed
	}

	// Parse all templates in the directory, with the custom functions
	templatePattern := filepath.Join(templatesDir, "*.tmpl")
	log.Debug("Loading templates", "pattern", templatePattern)
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(templatePattern) // Apply Funcs before ParseGlob
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates with funcs: %w", err)
	}

	// Ensure the main report template is specifically looked up if needed after parsing
	// For instance, if reportTemplate is always a specific file like "report.md.tmpl"
	// mainReportTemplate := tmpl.Lookup("report.md.tmpl")
	// if mainReportTemplate == nil {
	// 	log.Warn("Main report template 'report.md.tmpl' not found after parsing glob, some specific rendering might fail if it relies on this specific name.")
	// 	// Depending on strictness, this could be an error:
	// 	// return nil, fmt.Errorf("main report template 'report.md.tmpl' not found")
	// }

	renderer := &TemplateRenderer{
		log:       log,
		templates: tmpl,
		assetsDir: assetsDir,
		// reportTemplate: mainReportTemplate, // Use the looked-up template, or tmpl if generic usage is fine
	}

	return renderer, nil
}

// RenderReport renders the markdown report template with the provided data
// This function might need to be renamed or refactored if its primary purpose changes from MD.
func (tr *TemplateRenderer) RenderReport(data *ReportData) (string, error) {
	if data == nil {
		return "", fmt.Errorf("report data cannot be nil")
	}

	// Set generation timestamp if not already set
	if data.GenerationDate == "" {
		data.GenerationDate = time.Now().Format(time.RFC1123)
	}

	// Format fields for the template that need special handling
	tr.prepareTemplateData(data)

	// Fill missing but expected fields with defaults/placeholders
	tr.ensureMandatoryFields(data)

	// Render the markdown template
	var buf bytes.Buffer
	tmpl := tr.templates.Lookup("report.md.tmpl") // Still lookup markdown for this specific func
	if tmpl == nil {
		return "", fmt.Errorf("markdown template 'report.md.tmpl' not found")
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute markdown template: %w", err)
	}

	return buf.String(), nil
}

// RenderReportToFile renders the report template and writes it to a file
func (tr *TemplateRenderer) RenderReportToFile(data *ReportData, outputPath string) error {
	renderedReport, err := tr.RenderReport(data)
	if err != nil {
		return err
	}

	// Create the output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the rendered report to the file
	if err := os.WriteFile(outputPath, []byte(renderedReport), 0644); err != nil {
		return fmt.Errorf("failed to write report to file: %w", err)
	}

	tr.log.Info("Report successfully written to file", "path", outputPath)
	return nil
}

// RenderToMarkdown renders the report data to Markdown
func (tr *TemplateRenderer) RenderToMarkdown(data *ReportData, templateName string) (string, error) {
	var buf bytes.Buffer

	// Check if template exists
	templ := tr.templates.Lookup(templateName)
	if templ == nil {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	// If assetsDir is provided, use it to generate paths to assets
	if tr.assetsDir != "" && data.Plots != nil {
		// Process plot paths to make them accessible
		for key, relPath := range data.Plots { // relPath is actually just the filename, e.g., "altitude_vs_time.svg"
			// Construct the absolute path for verification.
			absolutePlotPath := filepath.Join(tr.assetsDir, relPath)

			// Log the path construction for debugging
			tr.log.Debug("Verifying plot asset path", "key", key, "renderer.assetsDir", tr.assetsDir, "plotFileName", relPath, "CheckingFullPath", absolutePlotPath)

			// Verify the file exists (the placeholder SVG created by the handler)
			if _, err := os.Stat(absolutePlotPath); err != nil {
				tr.log.Warn("Plot file not found during RenderToMarkdown verification", "key", key, "expectedPath", absolutePlotPath, "error", err)
				// Do NOT modify data.Plots[key] here. The embedSVG func handles path resolution.
			} else {
				tr.log.Debug("Plot file verified successfully", "key", key, "path", absolutePlotPath)
			}
		}
	}

	// Execute template
	err := templ.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// RenderToHTML renders the report data to HTML using the report.html.tmpl template
func (tr *TemplateRenderer) RenderToHTML(data *ReportData, templateName string) (string, error) {
	// Prepare data (timestamps, default fields etc.)
	if data.GenerationDate == "" {
		data.GenerationDate = time.Now().Format(time.RFC1123)
	}
	tr.prepareTemplateData(data)
	tr.ensureMandatoryFields(data)

	// Lookup the specific HTML template, defaulting to "report.html.tmpl" if templateName is empty
	htmlTemplateName := templateName
	if htmlTemplateName == "" {
		htmlTemplateName = "report.html.tmpl"
	}
	htmlTempl := tr.templates.Lookup(htmlTemplateName)
	if htmlTempl == nil {
		return "", fmt.Errorf("HTML template '%s' not found", htmlTemplateName)
	}

	var buf bytes.Buffer
	if err := htmlTempl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute HTML template '%s': %w", htmlTemplateName, err)
	}

	return buf.String(), nil
}

// prepareTemplateData formats and prepares the data for template rendering
func (tr *TemplateRenderer) prepareTemplateData(data *ReportData) {
	// Ensure the Plots map exists
	if data.Plots == nil {
		data.Plots = make(map[string]string)
	}

	// Update plot paths to be relative to the report
	for key, path := range data.Plots {
		// Make sure plot paths use forward slashes
		data.Plots[key] = filepath.ToSlash(path)
	}

	// Ensure MotionMetrics exists
	if data.MotionMetrics == nil {
		data.MotionMetrics = &MotionMetrics{}
	}

	// Ensure Summary structure is properly initialized
	data.Summary.MotionMetrics = *data.MotionMetrics
}

// ensureMandatoryFields ensures all fields required by the template are at least initialized
func (tr *TemplateRenderer) ensureMandatoryFields(data *ReportData) {
	// Initialize extensions map if needed
	if data.Extensions == nil {
		data.Extensions = make(map[string]interface{})
	}

	// Set default launch site name if not present
	if _, ok := data.Extensions["LaunchSiteName"]; !ok {
		data.Extensions["LaunchSiteName"] = "Unknown Launch Site"
	}

	// Add placeholder fields if they're missing but expected by the template
	for _, plotKey := range []string{"altitude_vs_time", "velocity_vs_time", "acceleration_vs_time"} {
		if _, exists := data.Plots[plotKey]; !exists {
			data.Plots[plotKey] = plotKey + ".svg" // Store just the filename
		}
	}

	// Other field initializations as needed
	if data.ReportTitle == "" {
		data.ReportTitle = fmt.Sprintf("Simulation Report for %s", data.RecordID)
	}
}

// CreateReportBundle creates a complete report bundle with the report file and all assets
func (tr *TemplateRenderer) CreateReportBundle(data *ReportData, outputDir string) error {
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create an assets subdirectory
	assetsDir := filepath.Join(outputDir, "assets")
	if err := os.MkdirAll(assetsDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create assets directory: %w", err)
	}

	// Generate the SVG plots
	if err := tr.GeneratePlots(data); err != nil {
		tr.log.Warn("Error generating some plots", "error", err)
		// Continue with report generation even if some plots fail
	}

	// Render the report to a file
	reportPath := filepath.Join(outputDir, "report.md")
	if err := tr.RenderReportToFile(data, reportPath); err != nil {
		return fmt.Errorf("failed to render report to file: %w", err)
	}

	tr.log.Info("Report bundle created successfully", "output_dir", outputDir)
	return nil
}

// GeneratePlots generates all plots defined in the ReportData and saves them to the assets directory.
// It iterates over the plot generation functions registered in the TemplateRenderer.
func (tr *TemplateRenderer) GeneratePlots(data *ReportData) error {
	tr.log.Debug("Starting plot generation", "assetsDir", tr.assetsDir)
	if data == nil {
		return fmt.Errorf("report data is nil, cannot generate plots")
	}

	// Ensure the assets directory exists
	if _, err := os.Stat(tr.assetsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tr.assetsDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create assets directory '%s': %w", tr.assetsDir, err)
		}
	}

	// Plot generation functions
	plotFunctions := map[string]func(*ReportData, string) error{
		"altitude_vs_time":     tr.generateAltitudeVsTimePlot,
		"velocity_vs_time":     tr.generateVelocityVsTimePlot,
		"acceleration_vs_time": tr.generateAccelerationVsTimePlot,
		"thrust_vs_time":       tr.GenerateThrustVsTimePlot,
		// Add other plot functions here, e.g., for trajectory, orientation, etc.
	}

	var firstErr error
	for plotKey, plotFunc := range plotFunctions {
		if plotKey == "" { // Skip if plot key is empty (e.g. not defined in ReportData.Plots)
			tr.log.Warn("Skipping plot generation for empty plot key")
			continue
		}
		plotPath := filepath.Join(tr.assetsDir, plotKey+".svg")
		tr.log.Debug("Generating plot", "key", plotKey, "path", plotPath)
		if err := plotFunc(data, plotPath); err != nil {
			tr.log.Error("Failed to generate plot", "key", plotKey, "path", plotPath, "error", err)
			if firstErr == nil {
				firstErr = fmt.Errorf("failed to generate plot '%s': %w", plotKey, err)
			}
			// Decide if we should continue or return on first error. For now, try to generate all.
		} else {
			tr.log.Info("Successfully generated plot", "key", plotKey, "path", plotPath)
		}
	}

	return firstErr // Return the first error encountered, or nil if all successful
}

// extractPlotData is a helper function to extract X and Y data series from motionData
func (tr *TemplateRenderer) extractPlotData(motionData []*PlotSimRecord, xKey, yKey string) (plotter.XYs, error) {
	pts := make(plotter.XYs, 0, len(motionData))

	if len(motionData) == 0 {
		return nil, fmt.Errorf("motionData is empty")
	}

	var foundData bool
	for i, record := range motionData {
		if record == nil {
			tr.log.Debug("Skipping nil record in motionData", "index", i)
			continue
		}

		// Log all keys in the current record for debugging
		keysInRecord := make([]string, 0, len(*record))
		for k := range *record {
			keysInRecord = append(keysInRecord, k)
		}
		tr.log.Debug("Keys in current PlotSimRecord", "index", i, "keys", keysInRecord, "expectedX", xKey, "expectedY", yKey)

		xValRaw, xOk := (*record)[xKey]
		yValRaw, yOk := (*record)[yKey]

		if !xOk || !yOk {
			// Log only once per key pair if missing for all records, or be more verbose if needed
			// For now, let's assume it's fine if some records don't have all keys, but we need at least some.
			continue
		}

		xVal, xAssertOk := getFloat64(xValRaw)
		yVal, yAssertOk := getFloat64(yValRaw)

		if !xAssertOk || !yAssertOk {
			tr.log.Warn("Data type assertion failed for plot data", "xKey", xKey, "yKey", yKey, "recordIndex", i, "xType", fmt.Sprintf("%T", xValRaw), "yType", fmt.Sprintf("%T", yValRaw))
			continue
		}

		pts = append(pts, plotter.XY{X: xVal, Y: yVal})
		foundData = true
	}

	if !foundData {
		return nil, fmt.Errorf("no valid data points found for keys X='%s', Y='%s'", xKey, yKey)
	}
	tr.log.Debug("Extracted plot data points", "count", len(pts), "xKey", xKey, "yKey", yKey)
	return pts, nil
}

// These are placeholder implementations for the plot generators
// They would be replaced with actual plotting logic using your plotting library

func (tr *TemplateRenderer) generateAltitudeVsTimePlot(data *ReportData, outputPath string) error {
	pts, err := tr.extractPlotData(data.MotionData, "time", "altitude")
	if err != nil {
		return fmt.Errorf("failed to extract altitude vs time data: %w", err)
	}

	p := plot.New()

	p.Title.Text = "Altitude vs. Time"
	p.X.Label.Text = LabelTimeSeconds
	p.Y.Label.Text = "Altitude (m)"

	line, err := plotter.NewLine(pts)
	if err != nil {
		return fmt.Errorf(ErrCreateLinePlot, err)
	}
	line.Color = color.RGBA{B: 255, A: 255} // Blue line

	p.Add(line)
	p.Add(plotter.NewGrid()) // Add a grid for better readability

	// Save the plot to a SVG file.
	if err := p.Save(6*vg.Inch, 4*vg.Inch, outputPath); err != nil {
		return fmt.Errorf(ErrSavePlot, err)
	}
	tr.log.Info("Successfully generated altitude vs time plot", "path", outputPath)
	return nil
}

func (tr *TemplateRenderer) generateVelocityVsTimePlot(data *ReportData, outputPath string) error {
	pts, err := tr.extractPlotData(data.MotionData, "time", "velocity") // Assuming 'velocity' is the key
	if err != nil {
		return fmt.Errorf("failed to extract velocity vs time data: %w", err)
	}

	p := plot.New()

	p.Title.Text = "Velocity vs. Time"
	p.X.Label.Text = LabelTimeSeconds
	p.Y.Label.Text = "Velocity (m/s)"

	line, err := plotter.NewLine(pts)
	if err != nil {
		return fmt.Errorf(ErrCreateLinePlot, err)
	}
	line.Color = color.RGBA{R: 255, A: 255} // Red line

	p.Add(line)
	p.Add(plotter.NewGrid())

	if err := p.Save(6*vg.Inch, 4*vg.Inch, outputPath); err != nil {
		return fmt.Errorf(ErrSavePlot, err)
	}
	tr.log.Info("Successfully generated velocity vs time plot", "path", outputPath)
	return nil
}

func (tr *TemplateRenderer) generateAccelerationVsTimePlot(data *ReportData, outputPath string) error {
	pts, err := tr.extractPlotData(data.MotionData, "time", "acceleration") // Assuming 'acceleration' is the key
	if err != nil {
		return fmt.Errorf("failed to extract acceleration vs time data: %w", err)
	}

	p := plot.New()

	p.Title.Text = "Acceleration vs. Time"
	p.X.Label.Text = LabelTimeSeconds
	p.Y.Label.Text = "Acceleration (m/s²)"

	line, err := plotter.NewLine(pts)
	if err != nil {
		return fmt.Errorf(ErrCreateLinePlot, err)
	}
	line.Color = color.RGBA{G: 200, A: 255} // Green line

	p.Add(line)
	p.Add(plotter.NewGrid())

	if err := p.Save(6*vg.Inch, 4*vg.Inch, outputPath); err != nil {
		return fmt.Errorf(ErrSavePlot, err)
	}
	tr.log.Info("Successfully generated acceleration vs time plot", "path", outputPath)
	return nil
}

// GenerateThrustVsTimePlot generates a plot for thrust vs. time, if motor data is available.
func (tr *TemplateRenderer) GenerateThrustVsTimePlot(data *ReportData, outputPath string) error {
	// Check if MotorData is present and has entries
	if len(data.MotorData) == 0 {
		tr.log.Warn("No motor data available for thrust vs. time plot", "outputPath", outputPath)
		// Return a placeholder SVG or an error, depending on the desired behavior
		return nil
	}

	// Extract thrust vs time data from MotorData
	pts, err := tr.extractPlotData(data.MotorData, "time", "thrust")
	if err != nil {
		return fmt.Errorf("failed to extract thrust vs time data: %w", err)
	}

	p := plot.New()

	p.Title.Text = "Thrust vs. Time"
	p.X.Label.Text = LabelTimeSeconds
	p.Y.Label.Text = "Thrust (N)"

	line, err := plotter.NewLine(pts)
	if err != nil {
		return fmt.Errorf(ErrCreateLinePlot, err)
	}
	line.Color = color.RGBA{R: 128, G: 0, B: 128, A: 255} // Purple line

	p.Add(line)
	p.Add(plotter.NewGrid())

	if err := p.Save(6*vg.Inch, 4*vg.Inch, outputPath); err != nil {
		return fmt.Errorf(ErrSavePlot, err)
	}
	tr.log.Info("Successfully generated thrust vs time plot", "path", outputPath)
	return nil
}

// Note: ReportData, MotionMetrics and other report structures are defined in report.go
// This file extends their functionality with report rendering capabilities.

// RecoverySystem represents a recovery device like a parachute or streamer
// This is defined here if not already present in report.go
type RecoverySystem struct {
	Type        string  // Type of recovery system (parachute, streamer, etc)
	Deployment  float64 // Time of deployment in seconds
	DescentRate float64 // Descent rate in m/s
}

func getFloat64(value interface{}) (float64, bool) {
	val, ok := value.(float64)
	return val, ok
}
