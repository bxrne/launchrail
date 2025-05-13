package utils

import (
	"fmt"
	"strconv"
)

// ParseFloat attempts to parse a string into a float64
func ParseFloat(valueStr string, fieldName string) (float64, error) {
	if valueStr == "" {
		return 0.0, fmt.Errorf("%s is required", fieldName)
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0.0, fmt.Errorf("%s must be a valid number: %w", fieldName, err)
	}
	return value, nil
}

// ParseInt attempts to parse a string into an int
func ParseInt(valueStr string, fieldName string) (int, error) {
	if valueStr == "" {
		return 0, fmt.Errorf("%s is required", fieldName)
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", fieldName, err)
	}
	return value, nil
}
