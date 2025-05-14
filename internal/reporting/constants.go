package reporting

// Recovery system types
const (
	RecoverySystemDrogue = "Drogue Parachute"
	RecoverySystemMain   = "Main Parachute"
)

// Event types
const (
	EventLaunch     = "Launch"
	EventRailExit   = "Rail Exit"
	EventApogee     = "Apogee"
	EventTouchdown  = "Touchdown"
	EventBurnout    = "Burnout"
	EventDeployment = "Deployment"
)

// Status values
const (
	StatusDeployed = "DEPLOYED"
	StatusSafe     = "SAFE"
	StatusArmed    = "ARMED"
)

// Column headers and labels
const (
	ColumnTimeSeconds    = "Time (s)"
	ColumnAltitude       = "Altitude (m)"
	ColumnVelocity       = "Velocity (m/s)"
	ColumnAcceleration   = "Acceleration (m/s²)"
	ColumnThrustNewtons  = "Thrust (N)"
	ColumnEventName      = "Event"
	ColumnEventStatus    = "Status"
	ColumnEventComponent = "Component"
)

// Default values
const (
	DefaultDescentRateDrogue  = 20.0  // m/s
	DefaultDescentRateMain    = 5.0   // m/s
	DefaultMainDeployAltitude = 300.0 // meters
)
