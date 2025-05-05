package types

import "fmt"

// Event represents significant simulation milestones.
type Event int

const (
	None Event = iota
	Liftoff
	Apogee
	Land
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
	default:
		return fmt.Sprintf("UNKNOWN_EVENT(%d)", e)
	}
}
