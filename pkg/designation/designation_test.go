// File path: designation/designation_test.go
package designation_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/designation"
)

// TEST: GIVEN a valid designation WHEN New is called THEN it should return a valid Designation
func TestNew_ValidDesignation(t *testing.T) {
	input := "269H110-14A"
	d, err := designation.New(input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(d) != input {
		t.Errorf("expected %s, got %s", input, d)
	}
}

// TEST: GIVEN an invalid designation WHEN New is called THEN it should return an error
func TestNew_InvalidDesignation(t *testing.T) {
	input := "INVALID"
	_, err := designation.New(input)
	if err == nil {
		t.Errorf("expected error, got none")
	}
}

// TEST: GIVEN a valid designation WHEN Describe is called THEN it should return the correct description
func TestDescribe_ValidDesignation(t *testing.T) {
	input := designation.Designation("269H110-14A")
	expected := "TotalImpulse=269.00, Class=H, AverageThrust=110.00, DelayTime=14.00, Variant=A"
	result, err := input.Describe()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// TEST: GIVEN an invalid designation WHEN Describe is called THEN it should return an error
func TestDescribe_InvalidDesignation(t *testing.T) {
	input := designation.Designation("INVALID")
	_, err := input.Describe()
	if err == nil {
		t.Errorf("expected error, got none")
	}
}
