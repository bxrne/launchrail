package utils_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name      string
		valueStr  string
		fieldName string
		expected  float64
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid float",
			valueStr:  "123.45",
			fieldName: "testField",
			expected:  123.45,
			wantErr:   false,
		},
		{
			name:      "empty string",
			valueStr:  "",
			fieldName: "testField",
			expected:  0.0,
			wantErr:   true,
			errMsg:    "testField is required",
		},
		{
			name:      "invalid float",
			valueStr:  "abc",
			fieldName: "testField",
			expected:  0.0,
			wantErr:   true,
			errMsg:    "testField must be a valid number: strconv.ParseFloat: parsing \"abc\": invalid syntax",
		},
		{
			name:      "integer as float",
			valueStr:  "123",
			fieldName: "testField",
			expected:  123.0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.ParseFloat(tt.valueStr, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name      string
		valueStr  string
		fieldName string
		expected  int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid int",
			valueStr:  "123",
			fieldName: "testField",
			expected:  123,
			wantErr:   false,
		},
		{
			name:      "empty string",
			valueStr:  "",
			fieldName: "testField",
			expected:  0,
			wantErr:   true,
			errMsg:    "testField is required",
		},
		{
			name:      "invalid int - text",
			valueStr:  "abc",
			fieldName: "testField",
			expected:  0,
			wantErr:   true,
			errMsg:    "testField must be a valid integer: strconv.Atoi: parsing \"abc\": invalid syntax",
		},
		{
			name:      "invalid int - float",
			valueStr:  "123.45",
			fieldName: "testField",
			expected:  0,
			wantErr:   true,
			errMsg:    "testField must be a valid integer: strconv.Atoi: parsing \"123.45\": invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.ParseInt(tt.valueStr, tt.fieldName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
