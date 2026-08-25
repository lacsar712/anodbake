package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/anodbake/internal/fsm"
	"github.com/lacsar712/anodbake/internal/interlock"
	"github.com/lacsar712/anodbake/internal/model"
)

func (a *App) StartSoak(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		if err := a.interlock.Gate().Allow(a.cfg.UnitID, snap.State); err != nil {
			return err
		}
		comb, err := a.pitfire.StartSoak(ctx, snap)
		if err != nil {
			return err
		}
		_ = a.store.UpdatePitfire(a.cfg.UnitID, comb)
		state, err := a.fsm.Dispatch(ctx, fsm.EvStartSoak)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("start_soak", storePayload("holder", holder))
		return nil
	})
}

func (a *App) CompleteSoak(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		if err := a.pitfire.CompleteSoak(snap.Pitfire); err != nil {
			return err
		}
		state, err := a.fsm.Dispatch(ctx, fsm.EvSoakComplete)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("soak_complete", "")
		return nil
	})
}

func (a *App) Ignite(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		a.refreshPermissives(snap)
		if err := a.permissives.CheckIgnition(); err != nil {
			return err
		}
		comb, err := a.pitfire.Ignite(ctx, snap)
		if err != nil {
			return err
		}
		_ = a.store.UpdatePitfire(a.cfg.UnitID, comb)
		state, err := a.fsm.Dispatch(ctx, fsm.EvIgnite)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("ignite", storePayload("holder", holder))
		return nil
	})
}

func (a *App) StabilizeIgnition(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		comb, err := a.pitfire.Stabilize(snap)
		if err != nil {
			return err
		}
		_ = a.store.UpdatePitfire(a.cfg.UnitID, comb)
		state, err := a.fsm.Dispatch(ctx, fsm.EvIgnitionStable)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("ignition_stable", "")
		return nil
	})
}

func (a *App) RampLoad(ctx context.Context, holder string, loadPct float64) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		if err := a.interlock.AuthorizeFiring(a.cfg.UnitID, holder, snap.State, now); err != nil {
			return err
		}
		comb := a.pitfire.RampToLoad(snap, loadPct)
		_ = a.store.UpdatePitfire(a.cfg.UnitID, comb)
		pitlineReading, err := a.pitline.Tick(ctx, snap, true)
		if err != nil {
			return err
		}
		_ = a.store.UpdatePitline(a.cfg.UnitID, pitlineReading)
		state, err := a.fsm.Dispatch(ctx, fsm.EvReachFiring)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("ramp_load", fmt.Sprintf(`{"load_pct":%f}`, loadPct))
		return nil
	})
}

func (a *App) Trip(ctx context.Context, reason string) error {
	snap := a.Snapshot()
	comb := a.pitfire.Trip(snap.Pitfire)
	_ = a.store.UpdatePitfire(a.cfg.UnitID, comb)
	a.interlock.Gate().Block(a.cfg.UnitID, "trip")
	state, err := a.fsm.Dispatch(ctx, fsm.EvTrip)
	if err != nil {
		return err
	}
	a.syncState(state)
	a.journalEvent("trip", storePayload("reason", reason))
	return nil
}

func (a *App) ResetTrip(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		a.interlock.Gate().Unblock(a.cfg.UnitID)
		state, err := a.fsm.Dispatch(ctx, fsm.EvResetTrip)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("reset_trip", "")
		return nil
	})
}

func (a *App) UpdateSettings(ctx context.Context, settings model.PlantSettings) error {
	select {
	case <-ctx.Done():
		return model.ErrContextDone
	default:
	}
	if err := a.pitline.ValidateSettings(settings); err != nil {
		return err
	}
	return a.store.UpdateSettings(a.cfg.UnitID, settings)
}

func (a *App) EnterService(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		state, err := a.fsm.Dispatch(ctx, fsm.EvEnterService)
		if err != nil {
			return err
		}
		a.syncState(state)
		return nil
	})
}

func (a *App) OnBakeLoss(ctx context.Context, holder string) error {
	_ = holder
	snap := a.Snapshot()
	if err := interlock.CheckBakeLoss(snap.Pitfire); err != nil {
		return fmt.Errorf("interlock denied")
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		a.StopTickLoop()
		a.cancelPitchLoop(holder)
		state, err := a.fsm.Dispatch(ctx, fsm.EvShutdown)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("shutdown", "")
		return nil
	})
}

func (a *App) StartTickLoop(parent context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if cancel, ok := a.tickCancels[a.tickLabel()]; ok {
		cancel()
	}
	ctx, cancel := context.WithCancel(parent)
	a.tickCancels[a.tickLabel()] = cancel
	interval := a.cfg.TickInterval
	if interval <= 0 {
		interval = time.Second
	}
	_ = a.scheduler.Schedule(ctx, a.tickLabel(), interval, a.runTick)
}

func (a *App) StopTickLoop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if cancel, ok := a.tickCancels[a.tickLabel()]; ok {
		cancel()
		delete(a.tickCancels, a.tickLabel())
	}
}

func (a *App) runTick(ctx context.Context) error {
	snap := a.Snapshot()
	firing := a.isFiring(snap.State)
	anodepitReading, err := a.anodepit.Tick(ctx, snap, firing)
	if err != nil {
		return err
	}
	_ = a.store.UpdateAnodepit(a.cfg.UnitID, anodepitReading)
	snap, _ = a.store.Require(a.cfg.UnitID)
	pitlineReading, err := a.pitline.Tick(ctx, snap, firing)
	if err != nil {
		return err
	}
	_ = a.store.UpdatePitline(a.cfg.UnitID, pitlineReading)
	a.refreshPermissives(a.Snapshot())
	a.telemetry.RecordTick(firing)
	if a.anodepit.TripRequired(anodepitReading) {
		_ = a.Trip(ctx, "anodepit_level")
		return nil
	}
	// Promote pitfire interlocks (over-temperature / O2 runaway / burner trip)
	// to a plant trip. Without this, a satisfied over-temperature interlock is
	// silently dropped at the cross-layer handoff, the plant never reaches
	// StateTrip, EvResetTrip stays an illegal transition, and the reset page
	// surfaces a generic 409 conflict instead of the over-temperature reset.
	if a.pitfire.Burner().TripRequired(snap.Pitfire) {
		_ = a.Trip(ctx, "pitfire_overtemp")
		return nil
	}
	return nil
}
