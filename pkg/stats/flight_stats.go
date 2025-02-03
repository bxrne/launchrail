package stats

import (
	"fmt"
	"math"
	"sync"
)

// FlightStats represents statistics for a rocket flight
type FlightStats struct {
	mu                sync.RWMutex
	Apogee            float64
	MaxVelocity       float64
	MaxAccel          float64
	BurnTime          float64
	TimeToApogee      float64
	TotalFlightTime   float64
	MaxMach           float64
	GroundHitVelocity float64
}

// NewFlightStats creates a new FlightStats object
func NewFlightStats() *FlightStats {
	return &FlightStats{
		Apogee:      -math.MaxFloat64,
		MaxVelocity: -math.MaxFloat64,
		MaxAccel:    -math.MaxFloat64,
		MaxMach:     -math.MaxFloat64,
	}
}

// Update updates the flight statistics with new data
func (s *FlightStats) Update(time, altitude, velocity, accel, mach float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if altitude > s.Apogee {
		s.Apogee = altitude
		s.TimeToApogee = time
	}

	s.MaxVelocity = math.Max(s.MaxVelocity, math.Abs(velocity))
	s.MaxAccel = math.Max(s.MaxAccel, math.Abs(accel))
	s.MaxMach = math.Max(s.MaxMach, mach)
	s.TotalFlightTime = time

	if altitude <= 0 {
		s.GroundHitVelocity = velocity
	}
}

// String returns a string representation of the flight statistics
func (s *FlightStats) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return fmt.Sprintf("Apogee=%.2fm, MaxVelocity=%.2fm/s, MaxAccel=%.2fm/s², MaxMach=%.2f, GroundHitVelocity=%.2fm/s", s.Apogee, s.MaxVelocity, s.MaxAccel, s.MaxMach, s.GroundHitVelocity)
}
