package components

import "time"

type Component interface {
	Update(timeStep time.Duration) error
	String() string
}
