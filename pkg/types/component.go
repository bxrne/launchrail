package types

// Component represents a generic ECS component
type Component interface {
	Type() string
	Update(dt float64) error
	GetPlanformArea() float64
}

// System represents a generic ECS system
type System interface {
	Update(dt float32) error
	Priority() int
}
