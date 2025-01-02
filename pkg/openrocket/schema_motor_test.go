package openrocket_test

import (
	"github.com/bxrne/launchrail/pkg/openrocket"
	"testing"
)

// TEST: GIVEN a MotorMount struct WHEN calling the String method THEN return a string representation of the MotorMount struct
func TestSchemaMotorMountString(t *testing.T) {
	mm := &openrocket.MotorMount{
		IgnitionEvent:  "event",
		IgnitionDelay:  1.0,
		Overhang:       1.0,
		Motor:          openrocket.Motor{},
		IgnitionConfig: openrocket.IgnitionConfig{},
	}

	expected := "MotorMount{IgnitionEvent=event, IgnitionDelay=1.00, Overhang=1.00, Motor=Motor{ConfigID=, Type=, Manufacturer=, Digest=, Designation=, Diameter=0.00, Length=0.00, Delay=}, IgnitionConfig=IgnitionConfig{ConfigID=, IgnitionEvent=, IgnitionDelay=0.00}}"
	if mm.String() != expected {
		t.Errorf("Expected %s, got %s", expected, mm.String())
	}
}

// TEST: GIVEN a Motor struct WHEN calling the String method THEN return a string representation of the Motor struct
func TestSchemaMotorString(t *testing.T) {
	m := &openrocket.Motor{
		ConfigID:     "config",
		Type:         "type",
		Manufacturer: "manufacturer",
		Digest:       "digest",
		Designation:  "designation",
		Diameter:     1.0,
		Length:       1.0,
		Delay:        "1.00",
	}

	expected := "Motor{ConfigID=config, Type=type, Manufacturer=manufacturer, Digest=digest, Designation=designation, Diameter=1.00, Length=1.00, Delay=1.00}"
	if m.String() != expected {
		t.Errorf("Expected %s, got %s", expected, m.String())
	}
}

// TEST: GIVEN a IgnitionConfig struct WHEN calling the String method THEN return a string representation of the IgnitionConfig struct
func TestSchemaIgnitionConfigString(t *testing.T) {
	ic := &openrocket.IgnitionConfig{
		ConfigID:      "config",
		IgnitionEvent: "event",
		IgnitionDelay: 1.0,
	}

	expected := "IgnitionConfig{ConfigID=config, IgnitionEvent=event, IgnitionDelay=1.00}"
	if ic.String() != expected {
		t.Errorf("Expected %s, got %s", expected, ic.String())
	}
}
