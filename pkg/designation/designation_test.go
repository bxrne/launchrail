package designation_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/stretchr/testify/assert"
)

func TestNew_ValidDesignation(t *testing.T) {
	input := "269H110-14A"
	expected := designation.Designation(input)

	d, err := designation.New(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, d)
}

func TestNew_InvalidDesignation(t *testing.T) {
	invalidInput := "<invalid>"
	d, err := designation.New(invalidInput)
	assert.Error(t, err)
	assert.Empty(t, d)
}

func TestValidate_ValidDesignation(t *testing.T) {
	d := designation.Designation("269H110-14A")
	valid, err := d.Validate()

	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestValidate_InvalidDesignation(t *testing.T) {
	d := designation.Designation("<invalid>")
	valid, err := d.Validate()

	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestDescribe_ValidDesignation(t *testing.T) {
	d := designation.Designation("269H110-14A")
	description, err := d.Describe()

	assert.NoError(t, err)
	assert.Equal(t, "Total Impulse: 269.00 Ns, Class: H, Average Thrust: 110.00 N, Delay Time: 14.00 s, Variant: A", description)
}

func TestDescribe_InvalidDesignation(t *testing.T) {
	d := designation.Designation("Invalid123")
	description, err := d.Describe()

	assert.Error(t, err)
	assert.Empty(t, description)
}
