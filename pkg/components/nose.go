package components

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bxrne/launchrail/internal/openrocket"
)

type Nose struct {
	Density              float64
	Length               float64
	Thickness            float64
	Shape                NoseShape
	ShapeFactor          float64
	AftDiameter          float64
	AftShoulderLength    float64
	AftShoulderDiameter  float64
	AftShoulderThickness float64
	AftShoulderCapped    bool
	Flipped              bool
}

func (n *Nose) Update() error {
	return nil
}

func (n *Nose) String() string {
	return fmt.Sprintf("Shape=%s Length=%f ShapeFactor=%f AftDiameter=%f", n.Shape.String(), n.Length, n.ShapeFactor, n.AftDiameter)
}

type NoseShape int

const (
	OGIVE_NOSE NoseShape = iota
	CONIC_NOSE
	ELLIPSE_NOSE
	POWER_NOSE
	HAACK_NOSE
)

func (n *NoseShape) String() string {
	return [...]string{"OGIVE", "CONIC", "ELLIPSE", "POWER", "HAACK"}[*n]
}

func stringToNoseShape(shape string) NoseShape {
	switch shape {
	case "OGIVE":
		return OGIVE_NOSE
	case "CONIC":
		return CONIC_NOSE
	case "ELLIPSE":
		return ELLIPSE_NOSE
	case "POWER":
		return POWER_NOSE
	case "HAACK":
		return HAACK_NOSE
	default:
		return OGIVE_NOSE
	}
}

func NewNose(orkData *openrocket.Openrocket) (*Nose, error) {
	nose := orkData.Rocket.Subcomponents.Stage.Subcomponents.Nosecone

	densityVal, err := strconv.ParseFloat(nose.Material.Density, 64)
	if err != nil {
		return nil, err
	}

	lengthVal, err := strconv.ParseFloat(nose.Length, 64)
	if err != nil {
		return nil, err
	}

	thicknessVal, err := strconv.ParseFloat(nose.Thickness, 64)
	if err != nil {
		return nil, err
	}

	nose.Shape = strings.ToUpper(nose.Shape)
	noseShape := stringToNoseShape(nose.Shape)

	shapeParamater, err := strconv.ParseFloat(nose.Shapeparameter, 64)
	if err != nil {
		return nil, err
	}

	aftRadius, err := strconv.ParseFloat(nose.Aftradius, 64)
	if err != nil {
		return nil, err
	}

	aftShoulderLength, err := strconv.ParseFloat(nose.Aftshoulderlength, 64)
	if err != nil {
		return nil, err
	}

	aftShoulderRadius, err := strconv.ParseFloat(nose.Aftshoulderradius, 64)
	if err != nil {
		return nil, err
	}

	aftShoulderThickness, err := strconv.ParseFloat(nose.Aftshoulderthickness, 64)
	if err != nil {
		return nil, err
	}

	aftShoulderCapped, err := strconv.ParseBool(nose.Aftshouldercapped)
	if err != nil {
		return nil, err
	}

	flipped, err := strconv.ParseBool(nose.Isflipped)
	if err != nil {
		return nil, err
	}

	return &Nose{
		Density:              densityVal,
		Length:               lengthVal,
		Thickness:            thicknessVal,
		Shape:                noseShape,
		ShapeFactor:          shapeParamater,
		AftDiameter:          aftRadius * 2,
		AftShoulderLength:    aftShoulderLength,
		AftShoulderDiameter:  aftShoulderRadius * 2,
		AftShoulderThickness: aftShoulderThickness,
		AftShoulderCapped:    aftShoulderCapped,
		Flipped:              flipped,
	}, nil
}
