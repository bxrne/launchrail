package types

import "fmt"

// Event represents significant simulation milestones.
type Event int

const (
	None Event = iota
	Liftoff
	Apogee
	Land
	ParachuteDeploy
	Burnout
)

// String returns a string representation of the event.
func (e Event) String() string {
	switch e {
	case None:
		return "NONE"
	case Liftoff:
		return "LIFTOFF"
	case Apogee:
		return "APOGEE"
	case Land:
		return "LAND"
	case ParachuteDeploy:
		return "PARACHUTE_DEPLOY"
	case Burnout:
		return "BURNOUT"
	default:
		return fmt.Sprintf("UNKNOWN_EVENT(%d)", e)
	}
}
