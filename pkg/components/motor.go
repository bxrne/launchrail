package components

import "time"

// Motor represents a rocket motor component
type Motor interface {
	Update(timeStep time.Duration) error
	String() string
}
