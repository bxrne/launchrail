package windeffect

import (
	"github.com/bxrne/launchrail/pkg/systems"
	"github.com/zerodha/logf"
	"math"
)

type WindEffectPlugin struct {
	log       logf.Logger
	windSpeed float64
}

var Plugin WindEffectPlugin

func (p *WindEffectPlugin) Initialize(log logf.Logger) error {
	p.log = log
	p.windSpeed = 5.0 // m/s base wind speed
	return nil
}

func (p *WindEffectPlugin) Name() string {
	return "WindEffect"
}

func (p *WindEffectPlugin) Version() string {
	return "1.0.0"
}

func (p *WindEffectPlugin) BeforeSimStep(state *systems.RocketState) error {
	// Apply wind effect based on altitude
	windEffect := p.windSpeed * math.Sin(state.Time)
	state.Velocity += windEffect * 0.1 // Apply 10% of wind effect
	return nil
}

func (p *WindEffectPlugin) AfterSimStep(state *systems.RocketState) error {
	return nil
}

func (p *WindEffectPlugin) Cleanup() error {
	return nil
}
