package components

import (
	"fmt"
	"strconv"

	"github.com/bxrne/launchrail/internal/openrocket"
)

type Fins struct {
	Count int
}

func NewFins(orkData *openrocket.Openrocket) (*Fins, error) {
	count, err := strconv.Atoi(orkData.Rocket.Subcomponents.Stage.Subcomponents.Bodytube.Subcomponents.Trapezoidfinset.Fincount)
	if err != nil {
		return nil, err
	}

	
	return &Fins{
		Count: count,
	}, nil

}

func (f *Fins) String() string {
	return fmt.Sprintf("Fins=%d RootChord=%f TipChord=%f SweepLength=%f SweepAngle=%f Thickness=%f",
		f.Count, f.RootChord, f.TipChord, f.SweepLength, f.SweepAngle, f.Thickness)
}

func (f *Fins) Update() error {
	return nil
}
