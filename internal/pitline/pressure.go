package pitline

import (
	"math"

	"github.com/lacsar712/anodbake/internal/clock"
	"github.com/lacsar712/anodbake/internal/model"
)

type PressureModel struct {
	clk clock.ProcessClock
}

func NewPressureModel(clk clock.ProcessClock) *PressureModel {
	return &PressureModel{clk: clk}
}

func (p *PressureModel) Compute(snap model.PlantSnapshot, firing bool) (float64, error) {
	if !firing {
		return p.cooldown(snap.Pitline.SteamPressurePSI), nil
	}
	base := snap.Settings.TargetSteamPSI
	pitchFactor := snap.Pitfire.PitchFlowTPH / math.Max(snap.Settings.PitchFlowTPH, 1)
	anodepitFactor := snap.Anodepit.LevelPercent / snap.Settings.AnodepitLevelSetpoint
	if anodepitFactor <= 0 {
		anodepitFactor = 0.5
	}
	pressure := base * pitchFactor * math.Min(anodepitFactor, 1.2)
	return math.Min(pressure, model.MaxSteamPressurePSI*1.05), nil
}

func (p *PressureModel) cooldown(current float64) float64 {
	if current <= 0 {
		return 0
	}
	return math.Max(0, current*0.98)
}

func (p *PressureModel) WithinTripLimits(pressurePSI float64, firing bool) bool {
	if !firing {
		return true
	}
	return pressurePSI <= model.MaxSteamPressurePSI
}

func (p *PressureModel) DeviationFromSetpoint(snap model.PlantSnapshot) float64 {
	return snap.Pitline.SteamPressurePSI - snap.Settings.TargetSteamPSI
}

func (p *PressureModel) SlidingPressureTarget(snap model.PlantSnapshot, loadPct float64) float64 {
	base := snap.Settings.TargetSteamPSI
	if snap.Settings.Mode != model.ModeSliding {
		return base
	}
	minPSI := base * 0.7
	return minPSI + (base-minPSI)*loadPct
}

func (p *PressureModel) TripRequired(pressurePSI float64) bool {
	return pressurePSI > model.MaxSteamPressurePSI
}
