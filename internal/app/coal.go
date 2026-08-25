package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/anodbake/internal/clock"
	"github.com/lacsar712/anodbake/internal/model"
)

func (a *App) advanceClock(d time.Duration) {
	if mc, ok := a.clk.(*clock.ManualClock); ok {
		mc.Advance(d)
		time.Sleep(time.Millisecond)
	} else {
		time.Sleep(d)
	}
}

func (a *App) bindPitchLoop(holder string, ctx context.Context) context.Context {
	a.mu.Lock()
	if cancel, ok := a.pitchLoopCancels[holder]; ok {
		cancel()
	}
	child, cancel := context.WithCancel(ctx)
	a.pitchLoopCancels[holder] = cancel
	a.mu.Unlock()
	return child
}

func (a *App) cancelPitchLoop(holder string) {
	a.mu.Lock()
	if cancel, ok := a.pitchLoopCancels[holder]; ok {
		cancel()
		delete(a.pitchLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) cancelAllPitchLoops() {
	a.mu.Lock()
	for holder, cancel := range a.pitchLoopCancels {
		cancel()
		delete(a.pitchLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) CoalFeedTPH() float64 {
	return a.Snapshot().Pitfire.PitchFlowTPH
}

func (a *App) RunPitchRamp(ctx context.Context, holder string, targetTPH float64) error {
	loopCtx := a.bindPitchLoop(holder, ctx)
	defer a.cancelPitchLoop(holder)
	for {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		current := snap.Pitfire.PitchFlowTPH
		if current >= targetTPH {
			return nil
		}
		comb := snap.Pitfire
		comb.PitchFlowTPH = current + 1.0
		_ = a.store.UpdatePitfire(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.PitchFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
}

func (a *App) RunCoalFeed(ctx context.Context, holder string, steps int) error {
	loopCtx := a.bindPitchLoop(holder, ctx)
	defer a.cancelPitchLoop(holder)
	for i := 0; steps <= 0 || i < steps; i++ {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		comb := snap.Pitfire
		comb.PitchFlowTPH += 0.5
		_ = a.store.UpdatePitfire(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.PitchFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
	return nil
}
