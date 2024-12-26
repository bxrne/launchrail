package openrocket_test

import (
	"encoding/xml"
	"testing"

	"github.com/bxrne/launchrail/pkg/openrocket"
)

func TestOpenrocketDocumentParsing(t *testing.T) {
	// Sample OpenRocket XML data
	xmlData := `
<openrocket version="1.0" creator="TestCreator">
	<rocket>
		<name>TestRocket</name>
		<id>12345</id>
		<axialoffset method="static">0.5</axialoffset>
		<position type="absolute">1.5</position>
		<designer>John Doe</designer>
		<revision>1</revision>
		<motorconfiguration configid="config1" default="true">
			<stage number="1" active="true"/>
			<stage number="2" active="false"/>
		</motorconfiguration>
		<referencetype>someType</referencetype>
	</rocket>
</openrocket>
`

	// Unmarshal XML data into OpenrocketDocument
	var doc openrocket.OpenrocketDocument
	err := xml.Unmarshal([]byte(xmlData), &doc)
	if err != nil {
		t.Fatalf("Failed to unmarshal XML: %v", err)
	}

	// Validate top-level attributes
	if doc.Version != "1.0" {
		t.Errorf("Expected Version '1.0', got '%s'", doc.Version)
	}
	if doc.Creator != "TestCreator" {
		t.Errorf("Expected Creator 'TestCreator', got '%s'", doc.Creator)
	}

	// Validate RocketDocument
	if doc.Rocket.Name != "TestRocket" {
		t.Errorf("Expected Rocket Name 'TestRocket', got '%s'", doc.Rocket.Name)
	}
	if doc.Rocket.ID != "12345" {
		t.Errorf("Expected Rocket ID '12345', got '%s'", doc.Rocket.ID)
	}
	if doc.Rocket.AxialOffset.Method != "static" {
		t.Errorf("Expected AxialOffset Method 'static', got '%s'", doc.Rocket.AxialOffset.Method)
	}
	if doc.Rocket.AxialOffset.Value != 0.5 {
		t.Errorf("Expected AxialOffset Value '0.5', got '%f'", doc.Rocket.AxialOffset.Value)
	}
	if doc.Rocket.Position.Type != "absolute" {
		t.Errorf("Expected Position Type 'absolute', got '%s'", doc.Rocket.Position.Type)
	}
	if doc.Rocket.Position.Value != 1.5 {
		t.Errorf("Expected Position Value '1.5', got '%f'", doc.Rocket.Position.Value)
	}
	if doc.Rocket.Designer != "John Doe" {
		t.Errorf("Expected Designer 'John Doe', got '%s'", doc.Rocket.Designer)
	}
	if doc.Rocket.Revision != "1" {
		t.Errorf("Expected Revision '1', got '%s'", doc.Rocket.Revision)
	}
	if doc.Rocket.ReferenceType != "someType" {
		t.Errorf("Expected ReferenceType 'someType', got '%s'", doc.Rocket.ReferenceType)
	}

	// Validate MotorConfiguration
	if doc.Rocket.MotorConfiguration.ConfigID != "config1" {
		t.Errorf("Expected MotorConfiguration ConfigID 'config1', got '%s'", doc.Rocket.MotorConfiguration.ConfigID)
	}
	if !doc.Rocket.MotorConfiguration.Default {
		t.Errorf("Expected MotorConfiguration Default 'true', got '%t'", doc.Rocket.MotorConfiguration.Default)
	}
	if len(doc.Rocket.MotorConfiguration.Stages) != 2 {
		t.Fatalf("Expected 2 Stages, got '%d'", len(doc.Rocket.MotorConfiguration.Stages))
	}
	if doc.Rocket.MotorConfiguration.Stages[0].Number != 1 || !doc.Rocket.MotorConfiguration.Stages[0].Active {
		t.Errorf("Expected Stage 1 to be Active, got '%v'", doc.Rocket.MotorConfiguration.Stages[0])
	}
	if doc.Rocket.MotorConfiguration.Stages[1].Number != 2 || doc.Rocket.MotorConfiguration.Stages[1].Active {
		t.Errorf("Expected Stage 2 to be Inactive, got '%v'", doc.Rocket.MotorConfiguration.Stages[1])
	}
}

func TestDescribeMethod(t *testing.T) {
	doc := &openrocket.OpenrocketDocument{
		Version: "1.0",
		Creator: "TestCreator",
		Rocket: openrocket.RocketDocument{
			Name: "TestRocket",
		},
	}

	expected := "Version=1.0, Creator=TestCreator, Rocket=TestRocket"
	if doc.Describe() != expected {
		t.Errorf("Expected Describe output '%s', got '%s'", expected, doc.Describe())
	}
}

func TestStringMethod(t *testing.T) {
	doc := &openrocket.OpenrocketDocument{
		Version: "1.0",
		Creator: "TestCreator",
		Rocket: openrocket.RocketDocument{
			Name: "TestRocket",
			ID:   "12345",
			AxialOffset: openrocket.AxialOffset{
				Method: "static",
				Value:  0.5,
			},
			Position: openrocket.Position{
				Value: 1.5,
				Type:  "absolute",
			},
			Designer: "John Doe",
			Revision: "1",
			MotorConfiguration: openrocket.MotorConfiguration{
				ConfigID: "config1",
				Default:  true,
				Stages: []openrocket.Stage{
					{Number: 1, Active: true},
					{Number: 2, Active: false},
				},
			},
			ReferenceType: "someType",
		},
	}

	expected := "OpenrocketDocument{Version=1.0, Creator=TestCreator, Rocket=RocketDocument{Name=TestRocket, ID=12345, AxialOffset=AxialOffset{Method=static, Value=0.50}, Position=Position{Value=1.50, Type=absolute}, Designer=John Doe, Revision=1, MotorConfiguration=MotorConfiguration{ConfigID=config1, Default=true, Stages=(Stage{Number=1, Active=true}, Stage{Number=2, Active=false})}, ReferenceType=someType}}"
	if doc.String() != expected {
		t.Errorf("Expected String output '%s', got '%s'", expected, doc.String())
	}
}
