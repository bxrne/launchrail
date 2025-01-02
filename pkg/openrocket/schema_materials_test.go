package openrocket_test

import (
	"github.com/bxrne/launchrail/pkg/openrocket"
	"testing"
)

// TEST: GIVEN a Material struct WHEN calling the String method THEN return a string representation of the Material struct
func TestSchemaMaterialString(t *testing.T) {
	m := &openrocket.Material{
		Type:    "type1",
		Density: 1.0,
		Name:    "name1",
	}

	expected := "Material{Type=type1, Density=1.00, Name=name1}"
	if m.String() != expected {
		t.Errorf("Expected %s, got %s", expected, m.String())
	}
}

// TEST: GIVEN a FilletMaterial struct WHEN calling the String method THEN return a string representation of the FilletMaterial struct
func TestSchemaFilletMaterialString(t *testing.T) {
	fm := &openrocket.FilletMaterial{
		Type:    "type1",
		Density: 1.0,
		Name:    "name1",
	}

	expected := "FilletMaterial{Type=type1, Density=1.00, Name=name1}"
	if fm.String() != expected {
		t.Errorf("Expected %s, got %s", expected, fm.String())
	}
}

// TEST: GIVEN a LineMaterial struct WHEN calling the String method THEN return a string representation of the LineMaterial struct
func TestSchemaLineMaterialString(t *testing.T) {
	lm := &openrocket.LineMaterial{
		Type:    "type1",
		Density: 1.0,
		Name:    "name1",
	}

	expected := "LineMaterial{Type=type1, Density=1.00, Name=name1}"
	if lm.String() != expected {
		t.Errorf("Expected %s, got %s", expected, lm.String())
	}
}
