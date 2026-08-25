package anodepit

import (
	"math"

	"github.com/lacsar712/anodbake/internal/clock"
	"github.com/lacsar712/anodbake/internal/model"
)

type CarryoverMonitor struct {
	clk clock.ProcessClock
}

func NewCarryoverMonitor(clk clock.ProcessClock) *CarryoverMonitor {
	return &CarryoverMonitor{clk: clk}
}

func (c *CarryoverMonitor) Estimate(anodepit model.AnodepitReading, pressurePSI float64) float64 {
	if anodepit.Condition != model.AnodepitCarry && anodepit.Condition != model.AnodepitSwell {
		return anodepit.CarryoverPPM * 0.9
	}
	base := 50.0
	levelFactor := math.Max(0, anodepit.LevelPercent-70)
	pressureFactor := pressurePSI / 1000
	return base + levelFactor*10 + pressureFactor*5
}

func (c *CarryoverMonitor) AlarmThreshold() float64 { return 500 }

func (c *CarryoverMonitor) TripRequired(ppm float64) bool { return ppm > 1000 }

func (c *CarryoverMonitor) Severity(ppm float64) string {
	switch {
	case ppm > 1000:
		return "critical"
	case ppm > 500:
		return "high"
	case ppm > 200:
		return "medium"
	default:
		return "low"
	}
}

func (c *CarryoverMonitor) RecommendAction(anodepit model.AnodepitReading) string {
	if anodepit.CarryoverPPM > c.AlarmThreshold() {
		return "reduce_load_and_check_separators"
	}
	if anodepit.Condition == model.AnodepitSwell {
		return "hold_feedwater_ramp"
	}
	return "none"
}

func (c *CarryoverMonitor) SeparatorEfficiency(anodepit model.AnodepitReading) float64 {
	eff := 0.98
	if anodepit.LevelPercent > 80 {
		eff -= (anodepit.LevelPercent - 80) * 0.005
	}
	return math.Max(0.5, eff)
}
