package pitfire

import (
	"math"

	"github.com/lacsar712/anodbake/internal/clock"
	"github.com/lacsar712/anodbake/internal/model"
)

type BurnerController struct {
	clk clock.ProcessClock
}

func NewBurnerController(clk clock.ProcessClock) *BurnerController {
	return &BurnerController{clk: clk}
}

func (b *BurnerController) EstimatePitframeTemp(reading model.PitfireReading) float64 {
	base := 300.0
	pitchHeat := reading.PitchFlowTPH * 50
	airCool := reading.AirflowTPH * 2
	return base + pitchHeat - airCool
}

func (b *BurnerController) BakeStable(reading model.PitfireReading) bool {
	if reading.BurnerPhase != model.BurnerStable && reading.BurnerPhase != model.BurnerIgnition {
		return false
	}
	return reading.PitframeTempF > 800 && reading.ExcessO2Pct >= model.MinPitframeO2Percent
}

func (b *BurnerController) TripRequired(reading model.PitfireReading) bool {
	if reading.ExcessO2Pct > model.MaxPitframeO2Percent*2 {
		return true
	}
	if reading.BurnerPhase == model.BurnerTrip {
		return true
	}
	if reading.PitframeTempF > 3500 {
		return true
	}
	return false
}

func (b *BurnerController) PhaseLabel(phase model.BurnerPhase) string {
	switch phase {
	case model.BurnerIdle:
		return "Idle"
	case model.BurnerSoak:
		return "Soak"
	case model.BurnerIgnition:
		return "Ignition"
	case model.BurnerStable:
		return "Stable Bake"
	case model.BurnerTrip:
		return "Tripped"
	default:
		return string(phase)
	}
}

func (b *BurnerController) HeatReleaseMW(reading model.PitfireReading) float64 {
	return reading.PitchFlowTPH * 12.5
}

func (b *BurnerController) TurndownRatio(settings model.PlantSettings, currentPitch float64) float64 {
	if settings.PitchFlowTPH <= 0 {
		return 0
	}
	return currentPitch / settings.PitchFlowTPH
}

func (b *BurnerController) MinStablePitch(settings model.PlantSettings) float64 {
	return settings.PitchFlowTPH * 0.25
}

func (b *BurnerController) NormalizePitch(flow, max float64) float64 {
	return math.Min(math.Max(flow, 0), max)
}
