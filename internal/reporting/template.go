package reporting

import (
	"bytes"
	"fmt"
	"html/template"
	"image/color"
	"os"
	"path/filepath"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"github.com/zerodha/logf"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// TemplateRenderer handles report template processing and rendering
type TemplateRenderer struct {
	log            *logf.Logger
	templates      *template.Template
	assetsDir      string
	reportTemplate *template.Template
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

	// Parse all templates in the directory
	templatePattern := filepath.Join(templatesDir, "*.tmpl")
	log.Debug("Loading templates", "pattern", templatePattern)
	tmpl, err := template.ParseGlob(templatePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	renderer := &TemplateRenderer{
		log:            log,
		templates:      tmpl,
		assetsDir:      assetsDir,
		reportTemplate: tmpl, // For backward compatibility
	}

	return renderer, nil
}

// RenderReport renders the report template with the provided data
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

	// Render the template
	var buf bytes.Buffer
	if err := tr.reportTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
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

	// Create asset paths for templates
	// If assetsDir is provided, use it to generate paths to assets
	if tr.assetsDir != "" && data.Plots != nil {
		// Process plot paths to make them accessible
		for key, relPath := range data.Plots {
			// Create an absolute path for the asset (tr.assetsDir already points to the assets directory)
			// Do NOT add another "assets" to the path!
			absPath := filepath.Join(tr.assetsDir, relPath)

			// Log the path construction for debugging
			tr.log.Debug("Constructed asset path", "key", key, "assetsDir", tr.assetsDir, "relPath", relPath, "fullPath", absPath)

			// Verify the file exists
			if _, err := os.Stat(absPath); err == nil {
				// Update the path in the Plots map - use relative path for template
				data.Plots[key] = filepath.Join("assets", relPath)
				tr.log.Debug("Updated plot path", "key", key, "path", data.Plots[key])
			} else {
				tr.log.Warn("Plot file not found", "key", key, "path", absPath, "error", err)
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

// RenderToHTML renders the report data to HTML
func (tr *TemplateRenderer) RenderToHTML(data *ReportData, templateName string) (string, error) {
	// First render to Markdown
	md, err := tr.RenderReport(data)
	if err != nil {
		return "", err
	}

	// Create HTML rendering extensions - render with SVG support
	exts := blackfriday.WithExtensions(blackfriday.CommonExtensions | blackfriday.HardLineBreak)

	// Create HTML renderer with features we need
	opts := blackfriday.WithRenderer(
		blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
			Flags: blackfriday.CommonHTMLFlags | blackfriday.HrefTargetBlank,
		}),
	)

	// Convert Markdown to HTML
	unsafe := blackfriday.Run([]byte(md), opts, exts)

	// Create a sanitization policy that allows SVG elements
	p := bluemonday.UGCPolicy()

	// Allow SVG elements and attributes for our plots
	p.AllowElements("svg", "rect", "circle", "path", "line", "polyline", "polygon", "text", "g")
	p.AllowAttrs("width", "height", "viewBox", "xmlns", "fill", "stroke", "x", "y", "cx", "cy", "r", "d", "x1", "y1", "x2", "y2", "points").OnElements("svg", "rect", "circle", "path", "line", "polyline", "polygon")
	p.AllowAttrs("text-anchor", "font-size", "font-family", "font-weight", "transform").OnElements("text", "g")

	// Sanitize HTML
	html := p.SanitizeBytes(unsafe)

	return string(html), nil
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
			data.Plots[plotKey] = filepath.ToSlash(filepath.Join("assets", plotKey+".svg"))
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
	if err := tr.generatePlots(data, assetsDir); err != nil {
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

// generatePlots creates SVG plots for the report
func (tr *TemplateRenderer) generatePlots(data *ReportData, outputDir string) error {
	// Ensure we have motion data to plot
	if len(data.MotionData) == 0 {
		tr.log.Warn("No motion data available for plotting")
		return nil
	}

	// Generate plots using the plot_transformer package or other plotting library
	// This is a placeholder - actual implementation would depend on your plotting library
	plotGenerators := map[string]func(*ReportData, string) error{
		"altitude_vs_time":     tr.generateAltitudeVsTimePlot,
		"velocity_vs_time":     tr.generateVelocityVsTimePlot,
		"acceleration_vs_time": tr.generateAccelerationVsTimePlot,
	}

	for plotKey, generator := range plotGenerators {
		plotPath := filepath.Join(outputDir, plotKey+".svg")
		if err := generator(data, plotPath); err != nil {
			tr.log.Warn("Failed to generate plot", "plot", plotKey, "error", err)
			// Continue with other plots even if one fails, but don't update data.Plots for this key
		} else {
			// Update the plot path in the data
			// Ensure data.Plots is initialized if it's nil
			if data.Plots == nil {
				data.Plots = make(map[string]string)
			}
			data.Plots[plotKey] = filepath.ToSlash(filepath.Join("assets", plotKey+".svg"))
			tr.log.Info("Successfully generated plot", "plot", plotKey, "path", data.Plots[plotKey])
		}
	}

	return nil
}

// extractPlotData is a helper function to extract X and Y data series from motionData
func (tr *TemplateRenderer) extractPlotData(motionData []*plotSimRecord, xKey, yKey string) (plotter.XYs, error) {
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

		xValRaw, xOk := (*record)[xKey]
		yValRaw, yOk := (*record)[yKey]

		if !xOk || !yOk {
			// Log only once per key pair if missing for all records, or be more verbose if needed
			// For now, let's assume it's fine if some records don't have all keys, but we need at least some.
			continue
		}

		xVal, xAssertOk := xValRaw.(float64)
		yVal, yAssertOk := yValRaw.(float64)

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
	p.X.Label.Text = "Time (s)"
	p.Y.Label.Text = "Altitude (m)"

	line, err := plotter.NewLine(pts)
	if err != nil {
		return fmt.Errorf("failed to create new line plot: %w", err)
	}
	line.Color = color.RGBA{B: 255, A: 255} // Blue line

	p.Add(line)
	p.Add(plotter.NewGrid()) // Add a grid for better readability

	// Save the plot to a SVG file.
	if err := p.Save(6*vg.Inch, 4*vg.Inch, outputPath); err != nil {
		return fmt.Errorf("failed to save plot: %w", err)
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
	p.X.Label.Text = "Time (s)"
	p.Y.Label.Text = "Velocity (m/s)"

	line, err := plotter.NewLine(pts)
	if err != nil {
		return fmt.Errorf("failed to create new line plot: %w", err)
	}
	line.Color = color.RGBA{R: 255, A: 255} // Red line

	p.Add(line)
	p.Add(plotter.NewGrid())

	if err := p.Save(6*vg.Inch, 4*vg.Inch, outputPath); err != nil {
		return fmt.Errorf("failed to save plot: %w", err)
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
	p.X.Label.Text = "Time (s)"
	p.Y.Label.Text = "Acceleration (m/s²)"

	line, err := plotter.NewLine(pts)
	if err != nil {
		return fmt.Errorf("failed to create new line plot: %w", err)
	}
	line.Color = color.RGBA{G: 200, A: 255} // Green line

	p.Add(line)
	p.Add(plotter.NewGrid())

	if err := p.Save(6*vg.Inch, 4*vg.Inch, outputPath); err != nil {
		return fmt.Errorf("failed to save plot: %w", err)
	}
	tr.log.Info("Successfully generated acceleration vs time plot", "path", outputPath)
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
