package main

import (
	"math"

	"github.com/bxrne/launchrail/pkg/states"
	"github.com/zerodha/logf"
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

func (p *WindEffectPlugin) BeforeSimStep(state *states.PhysicsState) error {

	f := p.windSpeed * math.Sin(state.Time)
	state.Velocity.Vec.X += f
	return nil
}

func (p *WindEffectPlugin) AfterSimStep(state *states.PhysicsState) error {
	return nil
}

func (p *WindEffectPlugin) Cleanup() error {
	return nil
}
