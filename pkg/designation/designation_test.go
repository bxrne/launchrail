package designation_test

import (
	"testing"

	"github.com/bxrne/launchrail/pkg/designation"
	"github.com/stretchr/testify/assert"
)

// TEST: GIVEN a valid designation WHEN New is called THEN the designation is returned
func TestDefaultDesignationValidator_New_Valid(t *testing.T) {
	validator := &designation.DefaultDesignationValidator{}
	input := "269H110-14A"

	d, err := validator.New(input)
	assert.NoError(t, err)
	assert.Equal(t, designation.Designation(input), d)
}

// TEST: GIVEN an invalid designation WHEN New is called THEN an error is returned
func TestDefaultDesignationValidator_New_Invalid(t *testing.T) {
	validator := &designation.DefaultDesignationValidator{}
	input := "<invalid>"

	d, err := validator.New(input)
	assert.Error(t, err)
	assert.Empty(t, d)
}

// TEST: GIVEN a valid designation WHEN Validate is called THEN true is returned
func TestDefaultDesignationValidator_Validate_Valid(t *testing.T) {
	validator := &designation.DefaultDesignationValidator{}
	d := designation.Designation("269H110-14A")

	valid, err := validator.Validate(d)
	assert.NoError(t, err)
	assert.True(t, valid)
}

// TEST: GIVEN an invalid designation WHEN Validate is called THEN false is returned
func TestDefaultDesignationValidator_Validate_Invalid(t *testing.T) {
	validator := &designation.DefaultDesignationValidator{}
	d := designation.Designation("<invalid>")

	valid, err := validator.Validate(d)
	assert.NoError(t, err)
	assert.False(t, valid)
}

// TEST: GIVEN a valid designation WHEN Describe is called THEN the description is returned
func TestDescribe_ValidDesignation(t *testing.T) {
	d := designation.Designation("269H110-14A")

	description, err := d.Describe()
	assert.NoError(t, err)
	assert.Equal(t, "TotalImpulse=269.00, Class=H, AverageThrust=110.00, DelayTime=14.00, Variant=A", description)
}

// TEST: GIVEN an invalid designation WHEN Describe is called THEN an error is returned
func TestDescribe_InvalidDesignation(t *testing.T) {
	d := designation.Designation("<invalid>")

	description, err := d.Describe()
	assert.Error(t, err)
	assert.Empty(t, description)
}

// TEST: GIVEN an empty designation WHEN New is called THEN an error is returned
func TestDefaultDesignationValidator_New_Empty(t *testing.T) {
	validator := &designation.DefaultDesignationValidator{}
	input := ""

	d, err := validator.New(input)
	assert.Error(t, err)
	assert.Empty(t, d)
}

// TEST: GIVEN a very long invalid designation WHEN New is called THEN an error is returned
func TestDefaultDesignationValidator_New_TooLong(t *testing.T) {
	validator := &designation.DefaultDesignationValidator{}
	input := "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ-1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	d, err := validator.New(input)
	assert.Error(t, err)
	assert.Empty(t, d)
}

// TEST: GIVEN a valid designation WHEN Describe is called AND parsing errors occur THEN an error is returned
func TestDescribe_ParsingError(t *testing.T) {
	// Simulate a valid-looking designation but with invalid numeric parts
	d := designation.Designation("ABCDEF-GHIJ")

	description, err := d.Describe()
	assert.Error(t, err)
	assert.Empty(t, description)
}

// TEST: GIVEN a designation with partial valid format WHEN Validate is called THEN false is returned
func TestValidate_PartialValid(t *testing.T) {
	validator := &designation.DefaultDesignationValidator{}
	d := designation.Designation("123AB")

	valid, err := validator.Validate(d)
	assert.NoError(t, err)
	assert.False(t, valid)
}
