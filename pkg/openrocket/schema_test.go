package openrocket_test

import (
	"testing"

	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/pkg/openrocket"
)

// TEST: GIVEN a OpenrocketDocument struct WHEN calling the String method THEN return a string representation of the OpenrocketDocument struct
func TestSchemaOpenrocketDocumentString(t *testing.T) {
	ord := &openrocket.OpenrocketDocument{
		Version: "version",
		Creator: "creator",
		Rocket:  openrocket.RocketDocument{},
	}

	expected := "OpenrocketDocument{Version=version, Creator=creator, Rocket=RocketDocument{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Designer=, Revision=, MotorConfiguration=MotorConfiguration{ConfigID=, Default=false, Stages=()}, ReferenceType=, Subcomponents={Subcomponents{Stages=()}}}}"
	if ord.String() != expected {
		t.Errorf("Expected %s, got %s", expected, ord.String())
	}
}

// TEST: GIVEN a OpenrocketDocument struct WHEN calling the Bytes method THEN return a byte representation of the OpenrocketDocument
func TestSchemaOpenrocketDocumentByte(t *testing.T) {
	ord := &openrocket.OpenrocketDocument{
		Version: "version",
		Creator: "creator",
		Rocket:  openrocket.RocketDocument{},
	}

	expected := "OpenrocketDocument{Version=version, Creator=creator, Rocket=RocketDocument{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Designer=, Revision=, MotorConfiguration=MotorConfiguration{ConfigID=, Default=false, Stages=()}, ReferenceType=, Subcomponents={Subcomponents{Stages=()}}}}"
	if string(ord.Bytes()) != expected {
		t.Errorf("Expected %s, got %s", expected, string(ord.Bytes()))
	}
}

// TEST: GIVEN a OpenrocketDocument struct WHEN calling the Describe method then a shortform version of String() is returned
func TestSchemaOpenrocketDocumentDescribe(t *testing.T) {
	ord := &openrocket.OpenrocketDocument{
		Version: "version",
		Creator: "creator",
		Rocket:  openrocket.RocketDocument{},
	}

	expected := "Version=version, Creator=creator, Rocket="
	if ord.Describe() != expected {
		t.Errorf("Expected %s, got %s", expected, ord.Describe())
	}
}

// TEST: GIVEN a RocketDocument struct WHEN calling the String method THEN return a string representation of the RocketDocument struct
func TestSchemaRocketDocumentString(t *testing.T) {
	rd := &openrocket.RocketDocument{
		Name:               "name",
		ID:                 "id",
		AxialOffset:        openrocket.AxialOffset{},
		Position:           openrocket.Position{},
		Designer:           "designer",
		Revision:           "revision",
		MotorConfiguration: openrocket.MotorConfiguration{},
		ReferenceType:      "reference",
		Subcomponents:      openrocket.Subcomponents{},
	}

	expected := "RocketDocument{Name=name, ID=id, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Designer=designer, Revision=revision, MotorConfiguration=MotorConfiguration{ConfigID=, Default=false, Stages=()}, ReferenceType=reference, Subcomponents={Subcomponents{Stages=()}}}"
	if rd.String() != expected {
		t.Errorf("Expected %s, got %s", expected, rd.String())
	}
}

// TEST: GIVEN a Stage struct WHEN calling the String method THEN return a string representation of the Stage struct
func TestSchemaStageString(t *testing.T) {
	s := &openrocket.Stage{
		Number: 0,
		Active: true,
	}

	expected := "Stage{Number=0, Active=true}"
	if s.String() != expected {
		t.Errorf("Expected %s, got %s", expected, s.String())
	}
}

// TEST: GIVEN a MotorConfiguration struct WHEN calling the String method THEN return a string representation of the MotorConfiguration struct
func TestSchemaMotorConfigurationString(t *testing.T) {
	mc := &openrocket.MotorConfiguration{
		ConfigID: "config",
		Default:  false,
		Stages:   []openrocket.Stage{},
	}

	expected := "MotorConfiguration{ConfigID=config, Default=false, Stages=()}"
	if mc.String() != expected {
		t.Errorf("Expected %s, got %s", expected, mc.String())
	}
}

// TEST: GIVEN a Subcomponents struct WHEN calling the String method THEN return a string representation of the Subcomponents struct
func TestSchemaSubcomponentsString(t *testing.T) {
	sc := &openrocket.Subcomponents{
		Stages: []openrocket.RocketStage{},
	}

	expected := "Subcomponents{Stages=()}"
	if sc.String() != expected {
		t.Errorf("Expected %s, got %s", expected, sc.String())
	}
}

// TEST: GIVEN a RocketStage struct WHEN calling the String method THEN return a string representation of the RocketStage struct
func TestSchemaRocketStageString(t *testing.T) {
	rs := &openrocket.RocketStage{
		Name:                   "name",
		ID:                     "id",
		SustainerSubcomponents: openrocket.SustainerSubcomponents{},
	}

	expected := "RocketStage{Name=name, ID=id, SustainerSubcomponents=SustainerSubcomponents{Nosecone=Nosecone{Name=, ID=, Finish=, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Shape=, ShapeClipped=false, ShapeParameter=0.00, AftRadius=0.00, AftShoulderRadius=0.00, AftShoulderLength=0.00, AftShoulderThickness=0.00, AftShoulderCapped=false, IsFlipped=false, Subcomponents=NoseSubcomponents{MassComponent=MassComponent{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=}}}, BodyTube=BodyTube{Name=, ID=, Finish=, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Radius=, Subcomponents=BodyTubeSubcomponents{InnerTube=InnerTube{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, RadialDirection=0.00, OuterRadius=0.00, Thickness=0.00, ClusterConfiguration=, ClusterScale=0.00, ClusterRotation=0.00, MotorMount=MotorMount{IgnitionEvent=, IgnitionDelay=0.00, Overhang=0.00, Motor=Motor{ConfigID=, Type=, Manufacturer=, Digest=, Designation=, Diameter=0.00, Length=0.00, Delay=}, IgnitionConfig=IgnitionConfig{ConfigID=, IgnitionEvent=, IgnitionDelay=0.00}}, Subcomponents=InnerTubeSubcomponents{MotorMount=MotorMount{IgnitionEvent=, IgnitionDelay=0.00, Overhang=0.00, Motor=Motor{ConfigID=, Type=, Manufacturer=, Digest=, Designation=, Diameter=0.00, Length=0.00, Delay=}, IgnitionConfig=IgnitionConfig{ConfigID=, IgnitionEvent=, IgnitionDelay=0.00}}}}, TrapezoidFinset=(), Parachute=Parachute{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, CD=, Material=Material{Type=, Density=0.00, Name=}, DeployEvent=, DeployAltitude=0.00, DeployDelay=0.00, DeploymentConfig=DeploymentConfig{ConfigID=, DeployEvent=, DeployAltitude=0.00, DeployDelay=0.00}, Diameter=0.00, LineCount=0, LineLength=0.00, LineMaterial=LineMaterial{Type=, Density=0.00, Name=}}, CenteringRings=(), Shockcord=Shockcord{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, CordLength=0.00, Material=Material{Type=, Density=0.00, Name=}}}}}}"
	actual := rs.String()
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

// TEST: GIVEN a Subcomponents struct WHEN calling the List method THEN return a list of all Subcomponents
func TestSchemaSubcomponentsList(t *testing.T) {
	sc := &openrocket.Subcomponents{
		Stages: []openrocket.RocketStage{},
	}

	if len(sc.List()) != 0 {
		t.Errorf("Expected 0, got %d", len(sc.List()))
	}
}

// TEST: Given a valid OpenrocketDocument struct WHEN calling Validate THEN return nil
func TestSchemaOpenrocketDocumentValidate(t *testing.T) {
	ord := &openrocket.OpenrocketDocument{
		Version: "OpenRocket 15.03",
		Creator: "OpenRocket 15.03",
		Rocket: openrocket.RocketDocument{
			Subcomponents: openrocket.Subcomponents{
				Stages: []openrocket.RocketStage{
					{
						Name: "Sustainer",
					},
				},
			},
		},
	}

	cfg := &config.Engine{
		External: config.External{
			OpenRocketVersion: "15.03",
		},
	}

	if err := ord.Validate(cfg); err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

// TEST: Given an invalid OpenrocketDocument struct WHEN calling Validate THEN return an Error
func TestSchemaOpenrocketDocumentValidateError(t *testing.T) {
	ord := &openrocket.OpenrocketDocument{
		Version: "OpenRocket 15.03",
		Creator: "OpenRocket 15.04",
		Rocket:  openrocket.RocketDocument{},
	}

	cfg := &config.Engine{
		External: config.External{
			OpenRocketVersion: "15.03",
		},
	}

	if err := ord.Validate(cfg); err == nil {
		t.Errorf("Expected error, got nil")
	}
}
