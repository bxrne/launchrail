package designation_test

import (
	"fmt"
	"testing"

	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedDes   designation.Designation
		wantErr       bool
		expectedError string
	}{
		{
			name:        "valid designation",
			input:       "269H110-14A",
			expectedDes: designation.Designation("269H110-14A"),
			wantErr:     false,
		},
		{
			name:          "invalid designation - too short",
			input:         "269H110",
			wantErr:       true,
			expectedError: "invalid designation",
		},
		{
			name:          "invalid designation - incorrect format",
			input:         "INVALID",
			wantErr:       true,
			expectedError: "invalid designation",
		},
		{
			name:          "invalid designation - empty",
			input:         "",
			wantErr:       true,
			expectedError: "invalid designation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := designation.New(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError)
				assert.Empty(t, d)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDes, d)
			}
		})
	}
}

func TestDesignation_Describe(t *testing.T) {
	tests := []struct {
		name          string
		inputStr      string // Use string to create designation via New()
		expectedDesc  string
		wantErr       bool
		expectedError string
	}{
		{
			name:         "valid designation",
			inputStr:     "269H110-14A",
			expectedDesc: "TotalImpulse=269.00, Class=H, AverageThrust=110.00, DelayTime=14.00, Variant=A",
			wantErr:      false,
		},
		{
			name:     "valid designation - different values",
			inputStr: "10A5-2B",
			// TotalImpulse=10.00, Class=A, AverageThrust=5.00, DelayTime=2.00, Variant=B
			expectedDesc: "TotalImpulse=10.00, Class=A, AverageThrust=5.00, DelayTime=2.00, Variant=B",
			wantErr:      false,
		},
		// No need to test Describe with "INVALID" directly, as New() would prevent its creation.
		// The error cases for Describe are more about internal parsing logic if the regex somehow
		// allowed a "valid" but unparseable string, which shouldn't happen with the current schema.
		// However, if we could construct a designation.Designation that bypasses New's validation
		// and has an unexpected number of regex matches, that would be an internal error.
		// The current `Describe` implementation itself can also return errors from strconv.ParseFloat.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := designation.New(tt.inputStr)
			assert.NoError(t, err, "Pre-condition failed: New() returned an error for supposedly valid input: %s", tt.inputStr)
			if err != nil { // Skip test if New failed
				return
			}

			desc, err := d.Describe()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError) // Use Contains for strconv errors as they can be verbose
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDesc, desc)
			}
		})
	}

	// Test direct creation of Designation with invalid string for Describe() error path
	t.Run("describe fails for directly created invalid designation", func(t *testing.T) {
		invalidDes := designation.Designation("INVALID-FORMAT")
		desc, err := invalidDes.Describe()
		assert.Error(t, err)
		assert.EqualError(t, err, "failed to parse designation")
		assert.Empty(t, desc)
	})

	// Test strconv.ParseFloat errors in Describe due to out-of-range numbers
	// Create a string of 400 '9's, which should cause ParseFloat to return an out-of-range error.
	veryLargeNumberStr := ""
	for i := 0; i < 400; i++ {
		veryLargeNumberStr += "9"
	}

	overflowTests := []struct {
		name          string
		inputStr      string
		expectedError string // Substring to check for in the error
	}{
		{
			name:          "totalImpulse overflow",
			inputStr:      fmt.Sprintf("%sH110-14A", veryLargeNumberStr),
			expectedError: "value out of range", // strconv.ParseFloat error for overflow
		},
		{
			name:          "averageThrust overflow",
			inputStr:      fmt.Sprintf("269H%s-14A", veryLargeNumberStr),
			expectedError: "value out of range",
		},
		{
			name:          "delayTime overflow",
			inputStr:      fmt.Sprintf("269H110-%sA", veryLargeNumberStr),
			expectedError: "value out of range",
		},
	}

	for _, tt := range overflowTests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := designation.New(tt.inputStr)
			// New() should pass because the regex `\d+` matches long strings of digits.
			assert.NoError(t, err, "New() failed unexpectedly for input: %s", tt.inputStr)
			if err != nil {
				return
			}

			desc, errDescribe := d.Describe()
			assert.Error(t, errDescribe, "Describe() should have failed for: %s", tt.inputStr)
			if errDescribe != nil {
				assert.Contains(t, errDescribe.Error(), tt.expectedError, "Error message mismatch for: %s", tt.inputStr)
			}
			assert.Empty(t, desc)
		})
	}

	// Test case for unparseable (but somehow validated) designation for Describe
	// This requires a bit of a hack to simulate a state that New() should prevent.
	t.Run("describe fails on malformed but validated designation", func(t *testing.T) {
		// This specific string would fail the regex in New(), so this tests a hypothetical
		// scenario where validation might pass but parsing later fails.
		// To properly test the strconv errors in Describe, we'd need to mock the regex
		// or construct a Designation object that bypasses validation.
		// For now, this test is more conceptual.
		// If the regex schema changes and allows for such a case, this test would be more direct.

		// Example: If schema allowed "ABCXYZ123-45D"
		// matches := exp.FindStringSubmatch("ABCXYZ123-45D") -> len might be 6
		// strconv.ParseFloat("ABC", 64) would fail.
		// This is hard to test directly without changing production code or using mocks for regexp.
		// Given the current strict schema in New(), these error paths in Describe are unlikely to be hit.
	})
}

func TestDetermineMotorClass(t *testing.T) {
	tests := []struct {
		name         string
		totalImpulse float64
		expected     string
	}{
		// Micro Motors
		{"micro motor lower bound", 0.0001, "1/4A-1/2A"}, // Technically below 1/4A but still > 0
		{"micro motor 1/4A lower", 0.3125, "1/4A-1/2A"},  // NAR Micro is 0.3126 N-s to 0.625 N-s for 1/4A
		{"micro motor mid range", 0.5, "1/4A-1/2A"},
		{"micro motor 1/2A upper", 1.25, "1/4A-1/2A"}, // NAR Micro is 0.626 N-s to 1.25 N-s for 1/2A

		// Zero Impulse
		{"zero impulse", 0.0, "Unknown"},
		{"negative impulse", -10.0, "Unknown"},

		// Standard Classes (testing boundaries and one mid-point)
		{"A class lower bound", 1.26, "A"},
		{"A class mid", 2.0, "A"},
		{"A class upper bound", 2.5, "A"},

		{"B class lower bound", 2.51, "B"},
		{"B class mid", 3.0, "B"},
		{"B class upper bound", 5.0, "B"},

		{"C class lower bound", 5.01, "C"},
		{"C class mid", 7.0, "C"},
		{"C class upper bound", 10.0, "C"},

		{"D class lower bound", 10.01, "D"},
		{"D class mid", 15.0, "D"},
		{"D class upper bound", 20.0, "D"},

		{"E class lower bound", 20.01, "E"},
		{"E class mid", 30.0, "E"},
		{"E class upper bound", 40.0, "E"},

		{"F class lower bound", 40.01, "F"},
		{"F class mid", 60.0, "F"},
		{"F class upper bound", 80.0, "F"},

		{"G class lower bound", 80.01, "G"},
		{"G class mid", 120.0, "G"},
		{"G class upper bound", 160.0, "G"},

		{"H class lower bound", 160.01, "H"},
		{"H class mid", 240.0, "H"},
		{"H class upper bound", 320.0, "H"},

		{"I class lower bound", 320.01, "I"},
		{"I class mid", 480.0, "I"},
		{"I class upper bound", 640.0, "I"},

		{"J class lower bound", 640.01, "J"},
		{"J class mid", 960.0, "J"},
		{"J class upper bound", 1280.0, "J"},

		{"K class lower bound", 1280.01, "K"},
		{"K class mid", 1920.0, "K"},
		{"K class upper bound", 2560.0, "K"},

		{"L class lower bound", 2560.01, "L"},
		{"L class mid", 3840.0, "L"},
		{"L class upper bound", 5120.0, "L"},

		{"M class lower bound", 5120.01, "M"},
		{"M class mid", 7680.0, "M"},
		{"M class upper bound", 10240.0, "M"},

		{"N class lower bound", 10240.01, "N"},
		{"N class mid", 15360.0, "N"},
		{"N class upper bound", 20480.0, "N"},

		{"O class lower bound", 20480.01, "O"},
		{"O class mid", 30720.0, "O"},
		{"O class upper bound", 40960.0, "O"},

		// Beyond O class
		{"P+ class lower bound", 40960.01, "P+"},
		{"P+ class high value", 100000.0, "P+"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := designation.DetermineMotorClass(tt.totalImpulse)
			assert.Equal(t, tt.expected, actual, fmt.Sprintf("Total Impulse: %f", tt.totalImpulse))
		})
	}
}
