package ecs

// ECS represents all non fixed objects
type ECS struct {
	World      *World
	Launchrail *Launchrail
	Launchsite *Launchsite
}

// New creates a new ECS instance
func New() *ECS {
	return &ECS{
		World:      NewWorld(),
		Launchrail: NewLaunchrail(),
		Launchsite: NewLaunchsite(),
	}
}
