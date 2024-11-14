package utils_test

import (
	"strings"
	"testing"

	"github.com/bxrne/launchrail/pkg/utils"
)

// TEST: GIVEN a list of strings WHEN n is less than the length of the list THEN it returns the last n elements joined by newline
func TestTail_LessThanLength(t *testing.T) {
	strs := []string{"one", "two", "three", "four", "five"}
	n := 3
	expected := strings.Join([]string{"three", "four", "five"}, "\n")

	result := utils.Tail(strs, n)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TEST: GIVEN a list of strings WHEN n is equal to the length of the list THEN it returns the entire list joined by newline
func TestTail_EqualToLength(t *testing.T) {
	strs := []string{"one", "two", "three", "four", "five"}
	n := 5
	expected := strings.Join(strs, "\n")

	result := utils.Tail(strs, n)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TEST: GIVEN a list of strings WHEN n is greater than the length of the list THEN it returns the entire list joined by newline
func TestTail_GreaterThanLength(t *testing.T) {
	strs := []string{"one", "two", "three", "four", "five"}
	n := 10
	expected := strings.Join(strs, "\n")

	result := utils.Tail(strs, n)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TEST: GIVEN an empty list of strings WHEN n is any value THEN it returns an empty string
func TestTail_EmptyList(t *testing.T) {
	strs := []string{}
	n := 3
	expected := ""

	result := utils.Tail(strs, n)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TEST: GIVEN a list of strings WHEN n is zero THEN it returns an empty string
func TestTail_ZeroN(t *testing.T) {
	strs := []string{"one", "two", "three", "four", "five"}
	n := 0
	expected := ""

	result := utils.Tail(strs, n)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
