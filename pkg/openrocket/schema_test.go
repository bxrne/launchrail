package openrocket_test

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/bxrne/launchrail/pkg/openrocket"
)

func TestOpenrocketDocumentParsing(t *testing.T) {
	ork_xml := "../../testdata/openrocket/l1.xml"

	data, err := os.ReadFile(ork_xml)
	if err != nil {
		t.Errorf("Failed to read file: %s", err)
	}

	var doc openrocket.OpenrocketDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Errorf("Failed to unmarshal XML data: %s", err)
	}

	if doc.Version != "1.9" {
		t.Errorf("Expected version '1.9', got '%s'", doc.Version)
	}

	if doc.Creator != "OpenRocket 23.09" {
		t.Errorf("Expected creator 'OpenRocket 23.09', got '%s'", doc.Creator)
	}

	if doc.Rocket.Name != "L1 Attempt" {
		t.Errorf("Expected rocket name 'L1 Attempt', got '%s'", doc.Rocket.Name)
	}

	if doc.Rocket.ID != "0833142b-6d19-40d4-9443-954cbb7ef97b" {
		t.Errorf("Expected rocket ID '0833142b-6d19-40d4-9443-954cbb7ef97b', got '%s'", doc.Rocket.ID)
	}

	if doc.Rocket.AxialOffset.Method != "absolute" {
		t.Errorf("Expected axial offset method 'absolute', got '%s'", doc.Rocket.AxialOffset.Method)
	}

	if doc.Rocket.AxialOffset.Value != 0.0 {
		t.Errorf("Expected axial offset value '0.0', got '%f'", doc.Rocket.AxialOffset.Value)
	}

	if doc.Rocket.Position.Value != 0.0 {
		t.Errorf("Expected position value '0.0', got '%f'", doc.Rocket.Position.Value)
	}

	if doc.Rocket.Position.Type != "absolute" {
		t.Errorf("Expected position type 'absolute', got '%s'", doc.Rocket.Position.Type)
	}

	if doc.Rocket.Designer != "Adam Byrne" {
		t.Errorf("Expected designer 'Adam Byrne', got '%s'", doc.Rocket.Designer)
	}

	if doc.Rocket.Revision != "Workshop" {
		t.Errorf("Expected revision 'Workshop', got '%s'", doc.Rocket.Revision)
	}

	if doc.Rocket.MotorConfiguration.ConfigID != "dd819b45-fa7f-47b8-9d31-cbe18f77381a" {
		t.Errorf("Expected motor configuration ID 'dd819b45-fa7f-47b8-9d31-cbe18f77381a', got '%s'", doc.Rocket.MotorConfiguration.ConfigID)
	}

	if doc.Rocket.MotorConfiguration.Default != true {
		t.Errorf("Expected motor configuration default 'true', got '%t'", doc.Rocket.MotorConfiguration.Default)
	}

	if doc.Rocket.MotorConfiguration.Stages[0].Number != 0 {
		t.Errorf("Expected motor configuration stage 0 number '0', got '%d'", doc.Rocket.MotorConfiguration.Stages[0].Number)
	}

	if doc.Rocket.MotorConfiguration.Stages[0].Active != true {
		t.Errorf("Expected motor configuration stage 0 active 'true', got '%t'", doc.Rocket.MotorConfiguration.Stages[0].Active)
	}

	if doc.Rocket.ReferenceType != "maximum" {
		t.Errorf("Expected reference type 'maximum', got '%s'", doc.Rocket.ReferenceType)
	}

	if doc.Rocket.Subcomponents.Stages[0].Name != "Sustainer" {
		t.Errorf("Expected subcomponent stage 0 name 'Sustainer', got '%s'", doc.Rocket.Subcomponents.Stages[0].Name)
	}

	if doc.Rocket.Subcomponents.Stages[0].ID != "a353045a-b4cf-4a3f-bb7f-0aa6d1adfb64" {
		t.Errorf("Expected subcomponent stage 0 ID 'a353045a-b4cf-4a3f-bb7f-0aa6d1adfb64', got '%s'", doc.Rocket.Subcomponents.Stages[0].ID)
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
//
// func TestStringMethod(t *testing.T) {
// 	doc := &openrocket.OpenrocketDocument{
// 		Version: "1.0",
// 		Creator: "TestCreator",
// 		Rocket: openrocket.RocketDocument{
// 			Name: "TestRocket",
// 			ID:   "12345",
// 			AxialOffset: openrocket.AxialOffset{
// 				Method: "static",
// 				Value:  0.5,
// 			},
// 			Position: openrocket.Position{
// 				Value: 1.5,
// 				Type:  "absolute",
// 			},
// 			Designer: "John Doe",
// 			Revision: "1",
// 			MotorConfiguration: openrocket.MotorConfiguration{
// 				ConfigID: "config1",
// 				Default:  true,
// 				Stages: []openrocket.Stage{
// 					{Number: 1, Active: true},
// 					{Number: 2, Active: false},
// 				},
// 			},
// 			ReferenceType: "someType",
// 			Subcomponents: openrocket.Subcomponents{
// 				Stages: []openrocket.RocketStage{
// 					{Name: "Sustainer", ID: "a353045a-b4cf-4a3f-bb7f-0aa6d1adfb64"},
// 					{Name: "Booster", ID: "b353045a-b4cf-4a3f-bb7f-0aa6d1adfb64"},
// 					{Name: "Payload", ID: "c353045a-b4cf-4a3f-bb7f-0aa6d1adfb64"},
// 				},
// 			},
// 		},
// 	}
//
// 	expected := "OpenrocketDocument{Version=1.0, Creator=TestCreator, Rocket=RocketDocument{Name=TestRocket, ID=12345, AxialOffset=AxialOffset{Method=static, Value=0.50}, Position=Position{Value=1.50, Type=absolute}, Designer=John Doe, Revision=1, MotorConfiguration=MotorConfiguration{ConfigID=config1, Default=true, Stages=(Stage{Number=1, Active=true}, Stage{Number=2, Active=false})}, ReferenceType=someType, Subcomponents={Subcomponents{Stages=(RocketStage{Name=Sustainer, ID=a353045a-b4cf-4a3f-bb7f-0aa6d1adfb64, SustainerSubcomponents=SustainerSubcomponents{Nosecone=Nosecone{Name=, ID=, Finish=, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Shape=, ShapeClipped=false, ShapeParameter=0.00, AftRadius=0.00, AftShoulderRadius=0.00, AftShoulderLength=0.00, AftShoulderThickness=0.00, AftShoulderCapped=false, IsFlipped=false}}}, RocketStage{Name=Booster, ID=b353045a-b4cf-4a3f-bb7f-0aa6d1adfb64, SustainerSubcomponents=SustainerSubcomponents{Nosecone=Nosecone{Name=, ID=, Finish=, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Shape=, ShapeClipped=false, ShapeParameter=0.00, AftRadius=0.00, AftShoulderRadius=0.00, AftShoulderLength=0.00, AftShoulderThickness=0.00, AftShoulderCapped=false, IsFlipped=false}}}, RocketStage{Name=Payload, ID=c353045a-b4cf-4a3f-bb7f-0aa6d1adfb64, SustainerSubcomponents=SustainerSubcomponents{Nosecone=Nosecone{Name=, ID=, Finish=, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Shape=, ShapeClipped=false, ShapeParameter=0.00, AftRadius=0.00, AftShoulderRadius=0.00, AftShoulderLength=0.00, AftShoulderThickness=0.00, AftShoulderCapped=false, IsFlipped=false}}})}}}}"
// 	if doc.String() != expected {
// 		t.Errorf("Expected String output '%s', got '%s'", expected, doc.String())
// 	}
// }
