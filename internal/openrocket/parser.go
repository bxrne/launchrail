package openrocket

import (
	"archive/zip"
	"encoding/xml"

	"github.com/bxrne/launchrail/pkg/components"
	"github.com/bxrne/launchrail/pkg/entities"
	"github.com/pkg/errors"
)

// openOrkFile opens a .ork file and returns a zip.ReadCloser.
func openOrkFile(input string) (*zip.ReadCloser, error) {
	r, err := zip.OpenReader(input)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Decompress decompresses a .ork file and parses the XML content.
func Decompress(filePath string) (*Openrocket, error) {
	orkRC, err := openOrkFile(filePath)
	if err != nil {
		return nil, err
	}
	defer orkRC.Close()

	var rocket Openrocket
	for _, f := range orkRC.File {
		if f.Name == "rocket.ork" {
			rc, err := f.Open()
			if err != nil {
				return nil, errors.Wrap(err, "opening rocket.ork")
			}
			defer rc.Close()

			err = xml.NewDecoder(rc).Decode(&rocket)
			if err != nil {
				return nil, errors.Wrap(err, "decoding rocket.ork")
			}
		}
	}

	return &rocket, nil
}

// ConvertToECS converts the parsed OpenRocket data to ECS components.
func (o *Openrocket) ConvertToECS() *entities.ECS {
	ecs := entities.NewECS()

	for _, component := range o.Rocket.Components {
		entity := ecs.CreateEntity()
		rocketComponent := &components.RocketComponent{
			Name:     component.Name,
			Mass:     component.Mass,
			Position: component.Position,
		}
		ecs.AddComponent(entity, rocketComponent, "RocketComponent")
		ecs.AddComponent(entity, &components.Velocity{}, "Velocity")
		ecs.AddComponent(entity, &components.Position{}, "Position")
	}

	for _, stage := range o.Rocket.Stages {
		entity := ecs.CreateEntity()
		stageComponent := &components.Stage{
			Name: stage.Name,
		}
		ecs.AddComponent(entity, stageComponent, "Stage")
		ecs.AddComponent(entity, &components.Velocity{}, "Velocity")
		ecs.AddComponent(entity, &components.Position{}, "Position")
	}

	for _, finSet := range o.Rocket.FinSets {
		entity := ecs.CreateEntity()
		finSetComponent := &components.FinSet{
			Name:     finSet.Name,
			FinCount: finSet.FinCount,
			FinShape: finSet.FinShape,
		}
		ecs.AddComponent(entity, finSetComponent, "FinSet")
		ecs.AddComponent(entity, &components.Velocity{}, "Velocity")
		ecs.AddComponent(entity, &components.Position{}, "Position")
	}

	for _, bodyTube := range o.Rocket.BodyTubes {
		entity := ecs.CreateEntity()
		bodyTubeComponent := &components.BodyTube{
			Name:   bodyTube.Name,
			Radius: bodyTube.Radius,
			Length: bodyTube.Length,
		}
		ecs.AddComponent(entity, bodyTubeComponent, "BodyTube")
		ecs.AddComponent(entity, &components.Velocity{}, "Velocity")
		ecs.AddComponent(entity, &components.Position{}, "Position")
	}

	for _, noseCone := range o.Rocket.NoseCones {
		entity := ecs.CreateEntity()
		noseConeComponent := &components.NoseCone{
			Name: noseCone.Name,
		}
		ecs.AddComponent(entity, noseConeComponent, "NoseCone")
		ecs.AddComponent(entity, &components.Velocity{}, "Velocity")
		ecs.AddComponent(entity, &components.Position{}, "Position")
	}

	for _, transition := range o.Rocket.Transitions {
		entity := ecs.CreateEntity()
		transitionComponent := &components.Transition{
			Name: transition.Name,
		}
		ecs.AddComponent(entity, transitionComponent, "Transition")
		ecs.AddComponent(entity, &components.Velocity{}, "Velocity")
		ecs.AddComponent(entity, &components.Position{}, "Position")
	}

	for _, launchLug := range o.Rocket.LaunchLugs {
		entity := ecs.CreateEntity()
		launchLugComponent := &components.LaunchLug{
			Name:    launchLug.Name,
			LugType: launchLug.LugType,
		}
		ecs.AddComponent(entity, launchLugComponent, "LaunchLug")
		ecs.AddComponent(entity, &components.Velocity{}, "Velocity")
		ecs.AddComponent(entity, &components.Position{}, "Position")
	}

	for _, trapezoidal := range o.Rocket.Trapezoidal {
		entity := ecs.CreateEntity()
		trapezoidalComponent := &components.TrapezoidalFinSet{
			Name:             trapezoidal.Name,
			TrapezoidalShape: trapezoidal.TrapezoidalShape,
		}
		ecs.AddComponent(entity, trapezoidalComponent, "TrapezoidalFinSet")
		ecs.AddComponent(entity, &components.Velocity{}, "Velocity")
		ecs.AddComponent(entity, &components.Position{}, "Position")
	}

	for _, elliptical := range o.Rocket.Elliptical {
		entity := ecs.CreateEntity()
		ellipticalComponent := &components.EllipticalFinSet{
			Name:            elliptical.Name,
			EllipticalShape: elliptical.EllipticalShape,
		}
		ecs.AddComponent(entity, ellipticalComponent, "EllipticalFinSet")
		ecs.AddComponent(entity, &components.Velocity{}, "Velocity")
		ecs.AddComponent(entity, &components.Position{}, "Position")
	}

	for _, freeform := range o.Rocket.Freeform {
		entity := ecs.CreateEntity()
		freeformComponent := &components.FreeformFinSet{
			Name:          freeform.Name,
			FreeformShape: freeform.FreeformShape,
		}
		ecs.AddComponent(entity, freeformComponent, "FreeformFinSet")
		ecs.AddComponent(entity, &components.Velocity{}, "Velocity")
		ecs.AddComponent(entity, &components.Position{}, "Position")
	}

	return ecs
}
