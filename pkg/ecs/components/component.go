package components

// Component interface that all components must implement
type Component interface {
	Type() string
	Update(dt float64)
}
