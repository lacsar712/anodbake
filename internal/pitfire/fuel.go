package pitfire

import (
	"math"

	"github.com/lacsar712/anodbake/internal/clock"
	"github.com/lacsar712/anodbake/internal/model"
)

type PitchRegulator struct {
	clk clock.ProcessClock
}

func NewPitchRegulator(clk clock.ProcessClock) *PitchRegulator {
	return &PitchRegulator{clk: clk}
}

func (f *PitchRegulator) IgnitionRate(settings model.PlantSettings) float64 {
	return settings.PitchFlowTPH * 0.08
}

func (f *PitchRegulator) ComputeForLoad(settings model.PlantSettings, loadPct float64) float64 {
	loadPct = math.Max(0, math.Min(1, loadPct))
	return settings.PitchFlowTPH * loadPct
}

func (f *PitchRegulator) Ramp(current, target, maxStep float64) float64 {
	delta := target - current
	if math.Abs(delta) <= maxStep {
		return target
	}
	if delta > 0 {
		return current + maxStep
	}
	return current - maxStep
}

func (f *PitchRegulator) BtuPerHour(flowTPH float64) float64 {
	return flowTPH * 19_500_000
}

func (f *PitchRegulator) HeatInputMW(flowTPH float64) float64 {
	return flowTPH * 11.6
}

func (f *PitchRegulator) ValidatePermissive(settings model.PlantSettings, anodepitOK, soakOK bool) error {
	if !soakOK {
		return model.ErrSoakIncomplete
	}
	if !anodepitOK {
		return model.ErrAnodepitLevelTrip
	}
	if settings.PitchFlowTPH <= 0 {
		return model.ErrPitchPermissive
	}
	return nil
}

func (f *PitchRegulator) MinFlow(settings model.PlantSettings) float64 {
	return settings.PitchFlowTPH * 0.2
}

func (f *PitchRegulator) MaxFlow(settings model.PlantSettings) float64 {
	return settings.PitchFlowTPH * 1.1
}
