package reporting

import (
	"math"
	"strings"

	"github.com/zerodha/logf"
)

// findApogeeFromMotionData attempts to find the apogee time by examining motion data
func findApogeeFromMotionData(motionData []*PlotSimRecord, headers []string, log *logf.Logger) (float64, float64) {
	// Find altitude column index
	timeIdx, altitudeIdx := -1, -1
	for i, header := range headers {
		headerLower := strings.ToLower(header)
		if strings.Contains(headerLower, "time") {
			timeIdx = i
		}
		if strings.Contains(headerLower, "alt") || strings.Contains(headerLower, "height") {
			altitudeIdx = i
		}
	}

	if timeIdx == -1 || altitudeIdx == -1 || len(motionData) == 0 {
		log.Warn("Cannot find apogee: missing time or altitude columns in motion data")
		return 0, 0
	}

	// Extract time/altitude values for easier processing
	type dataPoint struct {
		time     float64
		altitude float64
	}

	points := make([]dataPoint, 0, len(motionData))
	for _, record := range motionData {
		// Get column keys from header names
		timeKey := headers[timeIdx]
		altKey := headers[altitudeIdx]

		// Extract values from record
		timeVal, timeOk := (*record)[timeKey].(float64)
		altVal, altOk := (*record)[altKey].(float64)

		if timeOk && altOk {
			points = append(points, dataPoint{time: timeVal, altitude: altVal})
		}
	}

	if len(points) == 0 {
		log.Warn("No valid altitude data points found")
		return 0, 0
	}

	// Find the point with maximum altitude
	maxAlt := points[0].altitude
	apogeeTime := points[0].time

	for _, p := range points {
		if p.altitude > maxAlt {
			maxAlt = p.altitude
			apogeeTime = p.time
		}
	}

	log.Info("Found apogee from motion data", "time", apogeeTime, "altitude", maxAlt)
	return apogeeTime, maxAlt
}

// calculateDescentRates determines realistic descent rates for drogue and main parachutes
// based on motion data after apogee
func calculateDescentRates(motionData []*PlotSimRecord, headers []string, apogeeTime float64, log *logf.Logger) (float64, float64) {
	// Default values in case calculation fails
	defaultDrogueRate := 20.0
	defaultMainRate := 5.0

	// Find time, altitude and velocity columns
	timeIdx, altitudeIdx, velocityIdx := -1, -1, -1
	for i, header := range headers {
		headerLower := strings.ToLower(header)
		if strings.Contains(headerLower, "time") {
			timeIdx = i
		}
		if strings.Contains(headerLower, "alt") || strings.Contains(headerLower, "height") {
			altitudeIdx = i
		}
		if strings.Contains(headerLower, "vel") || strings.Contains(headerLower, "speed") {
			velocityIdx = i
		}
	}

	if timeIdx == -1 || altitudeIdx == -1 || velocityIdx == -1 || len(motionData) == 0 {
		log.Warn("Cannot calculate descent rates: missing required columns in motion data")
		return defaultDrogueRate, defaultMainRate
	}

	// Get column keys from header names
	timeKey := headers[timeIdx]
	altKey := headers[altitudeIdx]
	velKey := headers[velocityIdx]

	// Find points after apogee time
	postApogeeVelocities := []float64{}
	lateDescentVelocities := []float64{}

	for _, record := range motionData {
		// Extract values from record
		timeVal, timeOk := (*record)[timeKey].(float64)
		altVal, altOk := (*record)[altKey].(float64)
		velVal, velOk := (*record)[velKey].(float64)

		if timeOk && altOk && velOk {
			// Only process negative velocities since we're looking at descent
			if timeVal > apogeeTime && velVal < 0 {
				// Get absolute velocity since we're interested in descent speed
				velAbs := math.Abs(velVal)

				// Collect velocities in different phases of descent
				// Drogue phase (immediately after apogee)
				if timeVal < apogeeTime+10.0 {
					postApogeeVelocities = append(postApogeeVelocities, velAbs)
				}

				// Main phase (later in the descent)
				if altVal < 300 && altVal > 50 {
					// Typically main deployed at lower altitude
					lateDescentVelocities = append(lateDescentVelocities, velAbs)
				}
			}
		}
	}

	// Calculate average descent rates if we have data
	drogueRate := defaultDrogueRate
	mainRate := defaultMainRate

	if len(postApogeeVelocities) > 0 {
		// Calculate average velocity after apogee (drogue phase)
		sum := 0.0
		for _, v := range postApogeeVelocities {
			sum += v
		}
		drogueRate = sum / float64(len(postApogeeVelocities))
		log.Info("Calculated drogue descent rate from motion data", "rate", drogueRate, "samples", len(postApogeeVelocities))
	}

	if len(lateDescentVelocities) > 0 {
		// Calculate average velocity in late descent (main phase)
		sum := 0.0
		for _, v := range lateDescentVelocities {
			sum += v
		}
		mainRate = sum / float64(len(lateDescentVelocities))
		log.Info("Calculated main descent rate from motion data", "rate", mainRate, "samples", len(lateDescentVelocities))
	}

	// Sanity check on calculated values
	if drogueRate < 2.0 || drogueRate > 100.0 {
		log.Warn("Calculated drogue descent rate outside normal range, using default", "calculated", drogueRate, "default", defaultDrogueRate)
		drogueRate = defaultDrogueRate
	}

	if mainRate < 1.0 || mainRate > 50.0 {
		log.Warn("Calculated main descent rate outside normal range, using default", "calculated", mainRate, "default", defaultMainRate)
		mainRate = defaultMainRate
	}

	return drogueRate, mainRate
}
