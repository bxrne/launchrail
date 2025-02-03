package designation

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// Designation represents a motor designation string
type Designation string

// NOTE: TotalImpulse-Class-AverageThrust-DelayTime-Variant (e.g. "269H110-14A" is a valid designation)
var schema = `^(\d+)([A-Z]+)(\d+)-(\d+)([A-Z]+)$`

// New creates a new designation from a string
func New(designation string) (Designation, error) {
	d := Designation(designation)
	valid, err := d.validate()
	if !valid {
		return "", errors.New("invalid designation")
	}
	if err != nil {
		return "", err
	}
	return d, nil
}

// Validate the designation string
func (d Designation) validate() (bool, error) {
	// NOTE: TotalImpulse-Class-AverageThrust-DelayTime-Variant (e.g. "269H110-14A" is a valid designation)
	exp := regexp.MustCompile(schema)
	if !exp.MatchString(string(d)) {
		return false, nil
	}

	return true, nil
}

// Describe the designation
func (d *Designation) Describe() (string, error) {
	var totalImpulse float64
	var class string
	var averageThrust float64
	var delayTime float64
	var variant string

	exp := regexp.MustCompile(schema)
	matches := exp.FindStringSubmatch(string(*d))
	if len(matches) != 6 {
		return "", errors.New("failed to parse designation")
	}

	totalImpulse, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return "", err
	}
	class = matches[2]
	averageThrust, err = strconv.ParseFloat(matches[3], 64)
	if err != nil {
		return "", err
	}

	delayTime, err = strconv.ParseFloat(matches[4], 64)
	if err != nil {
		return "", err
	}
	variant = matches[5]

	return fmt.Sprintf("TotalImpulse=%.2f, Class=%s, AverageThrust=%.2f, DelayTime=%.2f, Variant=%s", totalImpulse, class, averageThrust, delayTime, variant), nil
}
