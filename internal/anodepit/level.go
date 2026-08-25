package anodepit

import (
	"math"

	"github.com/lacsar712/anodbake/internal/clock"
	"github.com/lacsar712/anodbake/internal/model"
)

type LevelController struct {
	clk clock.ProcessClock
}

func NewLevelController(clk clock.ProcessClock) *LevelController {
	return &LevelController{clk: clk}
}

func (l *LevelController) Compute(snap model.PlantSnapshot, firing bool) (float64, model.AnodepitCondition) {
	level := snap.Anodepit.LevelPercent
	if !firing {
		return level, model.AnodepitNormal
	}
	balance := snap.Anodepit.FeedwaterTPH - snap.Anodepit.SteamFlowTPH
	level += balance * 0.01
	level = math.Max(model.MinAnodepitLevelPercent, math.Min(model.MaxAnodepitLevelPercent, level))
	cond := l.classify(level, snap)
	return level, cond
}

func (l *LevelController) classify(level float64, snap model.PlantSnapshot) model.AnodepitCondition {
	setpoint := snap.Settings.AnodepitLevelSetpoint
	if level > setpoint+15 {
		return model.AnodepitSwell
	}
	if level < setpoint-15 {
		return model.AnodepitShrink
	}
	if snap.Pitline.SteamPressurePSI > snap.Settings.TargetSteamPSI*0.9 && level > setpoint+5 {
		return model.AnodepitCarry
	}
	return model.AnodepitNormal
}

func (l *LevelController) RecommendFeedwater(snap model.PlantSnapshot, firing bool) float64 {
	if !firing {
		return 0
	}
	err := snap.Settings.AnodepitLevelSetpoint - snap.Anodepit.LevelPercent
	return snap.Settings.FeedwaterFlowTPH + err*3
}

func (l *LevelController) WithinLimits(level float64) bool {
	return level >= model.MinAnodepitLevelPercent && level <= model.MaxAnodepitLevelPercent
}

func (l *LevelController) TripLow(level float64) bool  { return level < model.TripAnodepitLowPercent }
func (l *LevelController) TripHigh(level float64) bool { return level > model.TripAnodepitHighPercent }

func (l *LevelController) LevelError(snap model.PlantSnapshot) float64 {
	return snap.Anodepit.LevelPercent - snap.Settings.AnodepitLevelSetpoint
}

func (l *LevelController) ThreeElementBias(snap model.PlantSnapshot) float64 {
	steam := snap.Anodepit.SteamFlowTPH
	feed := snap.Anodepit.FeedwaterTPH
	levelErr := l.LevelError(snap)
	return feed + (steam-feed)*0.5 + levelErr*2
}
