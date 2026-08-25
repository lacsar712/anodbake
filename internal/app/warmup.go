package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/anodbake/internal/model"
)

func (a *App) WarmupStatus() (ready bool, detail string) {
	snap := a.Snapshot()
	if snap.Pitfire.SoakStartedAt.IsZero() {
		return false, "soak not started"
	}
	if !a.soakWindow.Ready(snap.Pitfire.SoakStartedAt) {
		return false, "soak window open"
	}
	if !snap.Pitfire.IgnitionAt.IsZero() && !a.warmupWindow.Ready(snap.Pitfire.IgnitionAt) {
		return false, "pitfire warmup window open"
	}
	if !snap.Anodepit.LastSwellAt.IsZero() {
		if err := a.anodepit.RequireSettled(snap.Anodepit); err != nil {
			return false, "anodepit swell settling"
		}
	}
	return true, "ready"
}

func (a *App) WaitWarmup(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w", model.ErrContextDone)
		default:
		}
		ready, _ := a.WarmupStatus()
		if ready {
			return nil
		}
	}
}

func (a *App) SoakRemaining() string {
	snap := a.Snapshot()
	if snap.Pitfire.SoakStartedAt.IsZero() {
		return "not started"
	}
	if a.soakWindow.Ready(snap.Pitfire.SoakStartedAt) {
		return "complete"
	}
	return "in progress"
}

func (a *App) PitfireWarmupRemaining() string {
	snap := a.Snapshot()
	if snap.Pitfire.IgnitionAt.IsZero() {
		return "not ignited"
	}
	if a.warmupWindow.Ready(snap.Pitfire.IgnitionAt) {
		return "complete"
	}
	return "in progress"
}
