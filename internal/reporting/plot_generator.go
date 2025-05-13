package reporting

import (
	"fmt"
	"image/color"
	"path/filepath"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// GenerateAltitudeVsTimePlot generates an SVG plot of altitude vs. time.
func (tr *TemplateRenderer) GenerateAltitudeVsTimePlot(data *ReportData) error {
	if data == nil || len(data.MotionData) == 0 {
		return fmt.Errorf("cannot generate altitude plot: no motion data")
	}

	pts := make(plotter.XYs, len(data.MotionData))
	for i, record := range data.MotionData {
		timeVal, timeOk := (*record)["time"].(float64)
		altitudeVal, altOk := (*record)["altitude"].(float64)

		if !timeOk || !altOk {
			tr.log.Warn("Skipping record for altitude plot due to type assertion failure", "index", i, "record", *record)
			continue // Or handle more gracefully
		}
		pts[i].X = timeVal
		pts[i].Y = altitudeVal
	}

	p := plot.New()
	p.Title.Text = "Altitude vs. Time"
	p.X.Label.Text = "Time (s)"
	p.Y.Label.Text = "Altitude (m)"

	line, err := plotter.NewLine(pts)
	if err != nil {
		return fmt.Errorf("failed to create line plotter: %w", err)
	}
	line.Color = color.RGBA{B: 255, A: 255} // Blue line
	p.Add(line)

	plotPath := filepath.Join(tr.assetsDir, "altitude_vs_time.svg")
	if err := p.Save(4*vg.Inch, 4*vg.Inch, plotPath); err != nil {
		return fmt.Errorf("failed to save plot %s: %w", plotPath, err)
	}
	tr.log.Info("Successfully generated plot", "path", plotPath)
	return nil
}

// TODO: Add more plot generation functions here for other metrics (velocity, acceleration, etc.)
// e.g., GenerateVelocityVsTimePlot, GenerateTrajectory3DPlot, etc.
