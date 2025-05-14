package reporting

import (
	"strconv"
	"strings"

	"github.com/zerodha/logf"
)

// findColumnIndices locates important column indices in event data
func findColumnIndices(headers []string, log *logf.Logger) (timeIdx, eventNameIdx, statusIdx, parachuteStatusIdx, parachuteTypeIdx int) {
	timeIdx, eventNameIdx, statusIdx, parachuteStatusIdx, parachuteTypeIdx = -1, -1, -1, -1, -1

	for i, header := range headers {
		headerLower := strings.ToLower(header)
		// First check for exact matches
		if headerLower == "time" || headerLower == "time (s)" {
			timeIdx = i
			continue
		}
		if headerLower == "event" || headerLower == "name" {
			eventNameIdx = i
			continue
		}
		if headerLower == "status" || headerLower == "state" {
			statusIdx = i
			continue
		}
		if headerLower == "parachute_status" || headerLower == "chute status" {
			parachuteStatusIdx = i
			continue
		}
		if headerLower == "parachute_type" || headerLower == "chute type" {
			parachuteTypeIdx = i
			continue
		}

		// Then check for partial matches if exact matches weren't found
		switch {
		case timeIdx == -1 && strings.Contains(headerLower, "time"):
			timeIdx = i
		case eventNameIdx == -1 && (strings.Contains(headerLower, "event") || strings.Contains(headerLower, "name")):
			eventNameIdx = i
		case statusIdx == -1 && strings.Contains(headerLower, "status") && !strings.Contains(headerLower, "parachute") && !strings.Contains(headerLower, "chute"):
			statusIdx = i
		case parachuteStatusIdx == -1 && ((strings.Contains(headerLower, "parachute") || strings.Contains(headerLower, "chute")) && strings.Contains(headerLower, "status")):
			parachuteStatusIdx = i
		case parachuteTypeIdx == -1 && ((strings.Contains(headerLower, "parachute") || strings.Contains(headerLower, "chute")) && strings.Contains(headerLower, "type")):
			parachuteTypeIdx = i
		}
	}

	// Use defaults if columns not found
	if timeIdx == -1 {
		timeIdx = 1 // Typically second column
		log.Warn("Could not find time column in events data, using default (column 1)")
	}
	if eventNameIdx == -1 {
		eventNameIdx = 0 // Typically first column
		log.Warn("Could not find event name column in events data, using default (column 0)")
	}

	return timeIdx, eventNameIdx, statusIdx, parachuteStatusIdx, parachuteTypeIdx
}

// parseDeploymentTime attempts to parse a time value from a string
func parseDeploymentTime(timeValue string, log *logf.Logger) (float64, bool) {
	deploymentTime, err := strconv.ParseFloat(timeValue, 64)
	if err != nil {
		log.Debug("Failed to parse parachute deployment time", "time_value", timeValue, "error", err)
		return 0, false
	}
	return deploymentTime, true
}

// determineParachuteType identifies the type of parachute from event data
func determineParachuteType(row []string, eventNameIdx, parachuteTypeIdx int, defaultType string) string {
	parachuteType := defaultType

	// Try to get specific type if we have that column
	if parachuteTypeIdx >= 0 && parachuteTypeIdx < len(row) {
		typeValue := strings.TrimSpace(row[parachuteTypeIdx])
		if typeValue != "" {
			parachuteType = typeValue
		}
	}

	// Fall back to detecting from event name if still using default
	if parachuteType == defaultType && eventNameIdx >= 0 && eventNameIdx < len(row) {
		eventName := strings.ToLower(row[eventNameIdx])
		if strings.Contains(eventName, "drogue") {
			parachuteType = RecoverySystemDrogue
		} else if strings.Contains(eventName, "main") {
			parachuteType = RecoverySystemMain
		}
	}

	return parachuteType
}

// getDescentRate determines the descent rate based on parachute type
func getDescentRate(parachuteType string) float64 {
	lowerType := strings.ToLower(parachuteType)
	if strings.Contains(lowerType, "drogue") {
		return DefaultDescentRateDrogue
	} else if strings.Contains(lowerType, "main") {
		return DefaultDescentRateMain
	}
	return 15.0 // Default fallback
}

// processParachuteDeployment creates or updates recovery system data in the map
func processParachuteDeployment(parachuteType string, deploymentTime float64, log *logf.Logger, parachuteMap map[string]RecoverySystemData) {
	descentRate := getDescentRate(parachuteType)

	log.Info("Found parachute deployment event", "type", parachuteType, "time", deploymentTime)

	parachuteMap[parachuteType] = RecoverySystemData{
		Type:        parachuteType,
		Deployment:  deploymentTime,
		DescentRate: descentRate,
	}
}

// isDeployedStatus checks if a status string indicates deployment
func isDeployedStatus(status string) bool {
	statusUpper := strings.ToUpper(strings.TrimSpace(status))
	return statusUpper == StatusDeployed || statusUpper == "TRUE" || statusUpper == "1"
}
