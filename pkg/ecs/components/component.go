package components

// Component interface that all components must implement
type Component interface {
	String() string
	Update(dt float64)
}
