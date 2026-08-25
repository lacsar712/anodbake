package pitfire

import (
	"context"
	"fmt"
	"math"

	"github.com/lacsar712/anodbake/internal/clock"
	"github.com/lacsar712/anodbake/internal/model"
)

type Coordinator struct {
	clk     clock.ProcessClock
	burner  *BurnerController
	airflow *AirflowBalancer
	pitch    *PitchRegulator
	soak   *clock.SoakWindow
	ignition *clock.IgnitionDelayWindow
	warmup  *clock.PitfireWarmupWindow
}

func NewCoordinator(clk clock.ProcessClock) *Coordinator {
	return &Coordinator{
		clk:      clk,
		burner:   NewBurnerController(clk),
		airflow:  NewAirflowBalancer(clk),
		pitch:     NewPitchRegulator(clk),
		soak:    clock.NewSoakWindow(clk),
		ignition: clock.NewIgnitionDelayWindow(clk),
		warmup:   clock.NewPitfireWarmupWindow(clk),
	}
}

func (c *Coordinator) Burner() *BurnerController  { return c.burner }
func (c *Coordinator) Airflow() *AirflowBalancer { return c.airflow }
func (c *Coordinator) Pitch() *PitchRegulator     { return c.pitch }

func (c *Coordinator) StartSoak(ctx context.Context, snap model.PlantSnapshot) (model.PitfireReading, error) {
	select {
	case <-ctx.Done():
		return snap.Pitfire, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	out := snap.Pitfire
	out.BurnerPhase = model.BurnerSoak
	out.SoakStartedAt = c.clk.Now()
	out.PitchFlowTPH = 0
	out.AirflowTPH = c.airflow.SoakRate()
	return out, nil
}

func (c *Coordinator) CompleteSoak(snap model.PitfireReading) error {
	return c.soak.Require(snap.SoakStartedAt)
}

func (c *Coordinator) Ignite(ctx context.Context, snap model.PlantSnapshot) (model.PitfireReading, error) {
	select {
	case <-ctx.Done():
		return snap.Pitfire, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if err := c.soak.Require(snap.Pitfire.SoakStartedAt); err != nil {
		return snap.Pitfire, err
	}
	out := snap.Pitfire
	out.BurnerPhase = model.BurnerIgnition
	out.IgnitionAt = c.clk.Now()
	out.PitchFlowTPH = c.pitch.IgnitionRate(snap.Settings)
	out.AirflowTPH = c.airflow.IgnitionRate(snap.Settings)
	out.PitframeTempF = 400
	return out, nil
}

func (c *Coordinator) Stabilize(snap model.PlantSnapshot) (model.PitfireReading, error) {
	if err := c.ignition.Require(snap.Pitfire.IgnitionAt); err != nil {
		return snap.Pitfire, err
	}
	out := snap.Pitfire
	out.BurnerPhase = model.BurnerStable
	out.PitchFlowTPH = snap.Settings.PitchFlowTPH * 0.5
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.PitframeTempF = c.burner.EstimatePitframeTemp(out)
	return out, nil
}

func (c *Coordinator) RampToLoad(snap model.PlantSnapshot, loadPct float64) model.PitfireReading {
	out := snap.Pitfire
	out.PitchFlowTPH = snap.Settings.PitchFlowTPH * loadPct
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.PitframeTempF = c.burner.EstimatePitframeTemp(out)
	return out
}

func (c *Coordinator) Trip(snap model.PitfireReading) model.PitfireReading {
	out := snap
	out.BurnerPhase = model.BurnerTrip
	out.PitchFlowTPH = 0
	out.PitframeTempF = math.Max(200, out.PitframeTempF*0.5)
	return out
}

func (c *Coordinator) WarmupReady(snap model.PitfireReading) bool {
	return c.warmup.Ready(snap.IgnitionAt)
}
