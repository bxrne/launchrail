package reporting

// Plot titles
const (
	PlotTitleAltitude     = "Altitude vs Time"
	PlotTitleVelocity     = "Velocity vs Time"
	PlotTitleAcceleration = "Acceleration vs Time"
	PlotTitleThrust       = "Thrust vs Time"
)

// Plot axis labels
const (
	PlotAxisTime         = "Time (s)"
	PlotAxisAltitude     = "Altitude (m)"
	PlotAxisVelocity     = "Velocity (m/s)"
	PlotAxisAcceleration = "Acceleration (m/s²)"
	PlotAxisThrust       = "Thrust (N)"
)

// Plot filenames
const (
	PlotFileAltitude     = "altitude_vs_time.svg"
	PlotFileVelocity     = "velocity_vs_time.svg"
	PlotFileAcceleration = "acceleration_vs_time.svg"
	PlotFileThrust       = "thrust_vs_time.svg"
)

// Event types
const (
	EventTypeRailExit = "Rail Exit"
)

// Message templates
const (
	MsgSuccessPlotGeneration = "Successfully generated %s plot"
)

// Default values
const (
	DefaultPlotWidth  = 800
	DefaultPlotHeight = 600
	DefaultPadding    = 30
	DefaultLineWidth  = 2
)
