package openrocket_test

import (
	"github.com/bxrne/launchrail/pkg/openrocket"
	"testing"
)

// TEST: GIVEN a SustainerSubcomponents struct WHEN calling the String method THEN return a string representation of the SustainerSubcomponents SustainerSubcomponents
func TestSchemaSustainerSubcomponentsString(t *testing.T) {
	ss := &openrocket.SustainerSubcomponents{
		Nosecone: openrocket.Nosecone{},
		BodyTube: openrocket.BodyTube{},
	}

	expected := "SustainerSubcomponents{Nosecone=Nosecone{Name=, ID=, Finish=, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Shape=, ShapeClipped=false, ShapeParameter=0.00, AftRadius=0.00, AftShoulderRadius=0.00, AftShoulderLength=0.00, AftShoulderThickness=0.00, AftShoulderCapped=false, IsFlipped=false, Subcomponents=NestedSubcomponents{CenteringRing=CenteringRing{Name=, ID=, InstanceCount=0, InstanceSeparation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, OuterRadius=, InnerRadius=}, MassComponent=MassComponent{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=}}}, BodyTube=BodyTube{Name=, ID=, Finish=, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Radius=, Subcomponents=BodyTubeSubcomponents{InnerTube=InnerTube{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, RadialDirection=0.00, OuterRadius=0.00, Thickness=0.00, ClusterConfiguration=, ClusterScale=0.00, ClusterRotation=0.00, MotorMount=MotorMount{IgnitionEvent=, IgnitionDelay=0.00, Overhang=0.00, Motor=Motor{ConfigID=, Type=, Manufacturer=, Digest=, Designation=, Diameter=0.00, Length=0.00, Delay=}, IgnitionConfig=IgnitionConfig{ConfigID=, IgnitionEvent=, IgnitionDelay=0.00}}, Subcomponents=NestedSubcomponents{CenteringRing=CenteringRing{Name=, ID=, InstanceCount=0, InstanceSeparation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, OuterRadius=, InnerRadius=}, MassComponent=MassComponent{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=}}}, TrapezoidFinset=TrapezoidFinset{Name=, ID=, InstanceCount=0, FinCount=0, RadiusOffset=RadiusOffset{Method=, Value=0.00}, AngleOffset=AngleOffset{Method=, Value=0.00}, Rotation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Finish=, Material=Material{Type=, Density=0.00, Name=}, Thickness=0.00, CrossSection=, Cant=0.00, TabHeight=0.00, TabLength=0.00, TabPositions=(), FilletRadius=0.00, RootChord=0.00, TipChord=0.00, SweepLength=0.00, Height=0.00}, Parachute=Parachute{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, CD=, Material=Material{Type=, Density=0.00, Name=}, DeployEvent=, DeployAltitude=0.00, DeployDelay=0.00, DeploymentConfig=DeploymentConfig{ConfigID=, DeployEvent=, DeployAltitude=0.00, DeployDelay=0.00}, Diameter=0.00, LineCount=0, LineLength=0.00, LineMaterial=LineMaterial{Type=, Density=0.00, Name=}}, CenteringRings=(), Shockcord=Shockcord{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, CordLength=0.00, Material=Material{Type=, Density=0.00, Name=}}}}}"
	if ss.String() != expected {
		t.Errorf("Expected %s, got %s", expected, ss.String())
	}
}

// TEST: GIVEN a BodyTube struct WHEN calling the String method THEN return a string representation of the BodyTube SustainerSubcomponents
func TestSchemaBodyTubeString(t *testing.T) {
	bt := &openrocket.BodyTube{
		Name:          "name",
		ID:            "id",
		Finish:        "finish",
		Material:      openrocket.Material{},
		Length:        0.0,
		Thickness:     0.0,
		Radius:        "0.0",
		Subcomponents: openrocket.BodyTubeSubcomponents{},
	}

	expected := "BodyTube{Name=name, ID=id, Finish=finish, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, Thickness=0.00, Radius=0.0, Subcomponents=BodyTubeSubcomponents{InnerTube=InnerTube{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, RadialDirection=0.00, OuterRadius=0.00, Thickness=0.00, ClusterConfiguration=, ClusterScale=0.00, ClusterRotation=0.00, MotorMount=MotorMount{IgnitionEvent=, IgnitionDelay=0.00, Overhang=0.00, Motor=Motor{ConfigID=, Type=, Manufacturer=, Digest=, Designation=, Diameter=0.00, Length=0.00, Delay=}, IgnitionConfig=IgnitionConfig{ConfigID=, IgnitionEvent=, IgnitionDelay=0.00}}, Subcomponents=NestedSubcomponents{CenteringRing=CenteringRing{Name=, ID=, InstanceCount=0, InstanceSeparation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, OuterRadius=, InnerRadius=}, MassComponent=MassComponent{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=}}}, TrapezoidFinset=TrapezoidFinset{Name=, ID=, InstanceCount=0, FinCount=0, RadiusOffset=RadiusOffset{Method=, Value=0.00}, AngleOffset=AngleOffset{Method=, Value=0.00}, Rotation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Finish=, Material=Material{Type=, Density=0.00, Name=}, Thickness=0.00, CrossSection=, Cant=0.00, TabHeight=0.00, TabLength=0.00, TabPositions=(), FilletRadius=0.00, RootChord=0.00, TipChord=0.00, SweepLength=0.00, Height=0.00}, Parachute=Parachute{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, CD=, Material=Material{Type=, Density=0.00, Name=}, DeployEvent=, DeployAltitude=0.00, DeployDelay=0.00, DeploymentConfig=DeploymentConfig{ConfigID=, DeployEvent=, DeployAltitude=0.00, DeployDelay=0.00}, Diameter=0.00, LineCount=0, LineLength=0.00, LineMaterial=LineMaterial{Type=, Density=0.00, Name=}}, CenteringRings=(), Shockcord=Shockcord{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, CordLength=0.00, Material=Material{Type=, Density=0.00, Name=}}}}"

	if bt.String() != expected {
		t.Errorf("Expected %s, got %s", expected, bt.String())
	}
}

// TEST: GIVEN a BodyTubeSubcomponents struct WHEN calling the String method THEN return a string representation of the BodyTubeSubcomponents SustainerSubcomponents
func TestSchemaBodyTubeSubcomponentsString(t *testing.T) {
	bts := &openrocket.BodyTubeSubcomponents{
		InnerTube:       openrocket.InnerTube{},
		TrapezoidFinset: openrocket.TrapezoidFinset{},
		Parachute:       openrocket.Parachute{},
	}

	expected := "BodyTubeSubcomponents{InnerTube=InnerTube{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, RadialDirection=0.00, OuterRadius=0.00, Thickness=0.00, ClusterConfiguration=, ClusterScale=0.00, ClusterRotation=0.00, MotorMount=MotorMount{IgnitionEvent=, IgnitionDelay=0.00, Overhang=0.00, Motor=Motor{ConfigID=, Type=, Manufacturer=, Digest=, Designation=, Diameter=0.00, Length=0.00, Delay=}, IgnitionConfig=IgnitionConfig{ConfigID=, IgnitionEvent=, IgnitionDelay=0.00}}, Subcomponents=NestedSubcomponents{CenteringRing=CenteringRing{Name=, ID=, InstanceCount=0, InstanceSeparation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, OuterRadius=, InnerRadius=}, MassComponent=MassComponent{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=}}}, TrapezoidFinset=TrapezoidFinset{Name=, ID=, InstanceCount=0, FinCount=0, RadiusOffset=RadiusOffset{Method=, Value=0.00}, AngleOffset=AngleOffset{Method=, Value=0.00}, Rotation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Finish=, Material=Material{Type=, Density=0.00, Name=}, Thickness=0.00, CrossSection=, Cant=0.00, TabHeight=0.00, TabLength=0.00, TabPositions=(), FilletRadius=0.00, RootChord=0.00, TipChord=0.00, SweepLength=0.00, Height=0.00}, Parachute=Parachute{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, CD=, Material=Material{Type=, Density=0.00, Name=}, DeployEvent=, DeployAltitude=0.00, DeployDelay=0.00, DeploymentConfig=DeploymentConfig{ConfigID=, DeployEvent=, DeployAltitude=0.00, DeployDelay=0.00}, Diameter=0.00, LineCount=0, LineLength=0.00, LineMaterial=LineMaterial{Type=, Density=0.00, Name=}}, CenteringRings=(), Shockcord=Shockcord{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, CordLength=0.00, Material=Material{Type=, Density=0.00, Name=}}}"

	if bts.String() != expected {
		t.Errorf("Expected %s, got %s", expected, bts.String())
	}
}

// TEST: GIVEN a InnerTube struct WHEN calling the String method THEN return a string representation of the InnerTube SustainerSubcomponents
func TestSchemaInnerTubeString(t *testing.T) {
	it := &openrocket.InnerTube{
		Name:                 "name",
		ID:                   "id",
		AxialOffset:          openrocket.AxialOffset{},
		Position:             openrocket.Position{},
		Material:             openrocket.Material{},
		Length:               0.0,
		RadialPosition:       0.0,
		RadialDirection:      0.0,
		OuterRadius:          0.0,
		Thickness:            0.0,
		ClusterConfiguration: "",
		ClusterScale:         0.0,
		ClusterRotation:      0.0,
		MotorMount:           openrocket.MotorMount{},
		Subcomponents:        openrocket.NoseSubcomponents{},
	}

	expected := "InnerTube{Name=name, ID=id, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, RadialDirection=0.00, OuterRadius=0.00, Thickness=0.00, ClusterConfiguration=, ClusterScale=0.00, ClusterRotation=0.00, MotorMount=MotorMount{IgnitionEvent=, IgnitionDelay=0.00, Overhang=0.00, Motor=Motor{ConfigID=, Type=, Manufacturer=, Digest=, Designation=, Diameter=0.00, Length=0.00, Delay=}, IgnitionConfig=IgnitionConfig{ConfigID=, IgnitionEvent=, IgnitionDelay=0.00}}, Subcomponents=NestedSubcomponents{CenteringRing=CenteringRing{Name=, ID=, InstanceCount=0, InstanceSeparation=0.00, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, Material=Material{Type=, Density=0.00, Name=}, Length=0.00, RadialPosition=0.00, OuterRadius=, InnerRadius=}, MassComponent=MassComponent{Name=, ID=, AxialOffset=AxialOffset{Method=, Value=0.00}, Position=Position{Value=0.00, Type=}, PackedLength=0.00, PackedRadius=0.00, RadialPosition=0.00, RadialDirection=0.00, Mass=0.00, Type=}}}"

	if it.String() != expected {
		t.Errorf("Expected %s, got %s", expected, it.String())
	}
}
