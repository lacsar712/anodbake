package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/anodbake/internal/pitline"
	"github.com/lacsar712/anodbake/internal/clock"
	"github.com/lacsar712/anodbake/internal/pitfire"
	"github.com/lacsar712/anodbake/internal/config"
	"github.com/lacsar712/anodbake/internal/anodepit"
	"github.com/lacsar712/anodbake/internal/fsm"
	"github.com/lacsar712/anodbake/internal/interlock"
	"github.com/lacsar712/anodbake/internal/model"
	"github.com/lacsar712/anodbake/internal/store"
)

type App struct {
	cfg           config.Config
	clk           clock.ProcessClock
	store         *store.PlantStore
	journal       *store.Journal
	fsm           *fsm.PitlineFSM
	pitline        *pitline.Controller
	pitfire    *pitfire.Coordinator
	anodepit          *anodepit.Coordinator
	interlock     *interlock.Interlock
	permissives   *interlock.PermissiveSet
	coordLock     *interlock.CoordinationLock
	scheduler     *clock.Scheduler
	soakWindow   *clock.SoakWindow
	warmupWindow  *clock.PitfireWarmupWindow
	telemetry     *Telemetry
	tickCancels    map[string]context.CancelFunc
	pitchLoopCancels map[string]context.CancelFunc
	// activePitchCancel is the stop-baking credential for THIS unit's pitch
	// loop. It is per-App (per-unit) rather than package-global so that one
	// unit's emergency stop / shutdown cannot cancel another unit's running
	// baking loop (see handover note: "停焙凭据换列时没分到单元").
	activePitchCancel context.CancelFunc
	mu             sync.RWMutex
}

func New(cfg config.Config, clk clock.ProcessClock) *App {
	return &App{
		cfg:          cfg,
		clk:          clk,
		store:        store.NewPlantStore(),
		journal:      store.NewJournal(cfg.JournalPath, cfg.JournalCapacity),
		fsm:          fsm.NewPitlineFSM(cfg.UnitID),
		pitline:       pitline.NewController(clk),
		pitfire:   pitfire.NewCoordinator(clk),
		anodepit:         anodepit.NewCoordinator(clk),
		interlock:    interlock.NewInterlock(cfg.LeaseTTL),
		permissives:  interlock.NewPermissiveSet(),
		coordLock:    interlock.NewCoordinationLock(),
		scheduler:    clock.NewScheduler(clk),
		soakWindow:  clock.NewSoakWindow(clk),
		warmupWindow: clock.NewPitfireWarmupWindow(clk),
		telemetry:    NewTelemetry(cfg.UnitID),
		tickCancels:     make(map[string]context.CancelFunc),
		pitchLoopCancels: make(map[string]context.CancelFunc),
	}
}

func (a *App) Snapshot() model.PlantSnapshot {
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return model.DefaultSnapshot(a.cfg.UnitID)
	}
	return snap
}

func (a *App) Config() config.Config              { return a.cfg }
func (a *App) Clock() clock.ProcessClock          { return a.clk }
func (a *App) FSM() *fsm.PitlineFSM                { return a.fsm }
func (a *App) UnitID() string                     { return a.cfg.UnitID }
func (a *App) Store() *store.PlantStore           { return a.store }
func (a *App) Interlock() *interlock.Interlock    { return a.interlock }
func (a *App) Telemetry() TelemetrySnapshot       { return a.telemetry.Snapshot() }
func (a *App) Journal() *store.Journal            { return a.journal }

func (a *App) journalEvent(ev, payload string) {
	_, _ = a.journal.Append(a.cfg.UnitID, ev, payload)
}

func (a *App) syncState(state model.PlantState) {
	_ = a.store.UpdateState(a.cfg.UnitID, state)
}

func (a *App) isFiring(state model.PlantState) bool {
	return state == model.StateFiring || state == model.StateLoadFollow || state == model.StateRamp
}

func (a *App) refreshPermissives(snap model.PlantSnapshot) {
	a.permissives.SetAnodepit(a.anodepit.Level().WithinLimits(snap.Anodepit.LevelPercent))
	a.permissives.SetPressure(a.pitline.Pressure().WithinTripLimits(snap.Pitline.SteamPressurePSI, a.isFiring(snap.State)))
	a.permissives.SetPitfire(a.pitfire.Burner().BakeStable(snap.Pitfire))
	a.permissives.SetPitch(snap.Pitfire.PitchFlowTPH > 0 || snap.State == model.StateSoak)
	a.permissives.SetIgnition(snap.Pitfire.BurnerPhase == model.BurnerStable || snap.Pitfire.BurnerPhase == model.BurnerIgnition)
	a.fsm.SetPitchPermissive(a.permissives.PitchOK())
	a.fsm.SetSoakComplete(a.soakWindow.Ready(snap.Pitfire.SoakStartedAt))
}

func (a *App) tickLabel() string {
	return fmt.Sprintf("%s-tick", a.cfg.UnitID)
}
