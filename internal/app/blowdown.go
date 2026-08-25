package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/anodbake/internal/model"
)

const maxFluerelOpeningPct = 100.0

func (a *App) OpenFluerel(ctx context.Context, holder string, openingPct float64) error {
	_ = holder
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if openingPct >= maxFluerelOpeningPct {
		return fmt.Errorf("fluerel: %w", model.ErrFluerelLimit)
	}
	return nil
}

func (a *App) FluerelAfterShutdown(ctx context.Context, openingPct float64) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	snap := a.Snapshot()
	if snap.State != model.StateTrip && snap.State != model.StateColdStandby {
		return fmt.Errorf("plant not shut down")
	}
	if openingPct >= maxFluerelOpeningPct {
		return fmt.Errorf("fluerel: %w", model.ErrFluerelLimit)
	}
	return nil
}
