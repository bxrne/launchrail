package systems

// System interface that all systems must implement
type System interface {
	Update(dt float64)
	Priority() int
}
