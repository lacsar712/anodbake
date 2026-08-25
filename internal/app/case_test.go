package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lacsar712/anodbake/internal/app"
	"github.com/lacsar712/anodbake/internal/clock"
	"github.com/lacsar712/anodbake/internal/config"
	"github.com/lacsar712/anodbake/internal/model"
)

func TestCase(t *testing.T) {
	clk := clock.NewManual(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := config.Default("FLAME-1")
	a, err := app.BootstrapWithClock(cfg, clk)
	if err != nil {
		t.Fatal(err)
	}
	comb := a.Snapshot().Pitfire
	comb.BurnerPhase = model.BurnerStable
	comb.PitframeTempF = 400
	if err := a.Store().UpdatePitfire(cfg.UnitID, comb); err != nil {
		t.Fatal(err)
	}
	err = a.OnBakeLoss(context.Background(), "maint-op")
	if err == nil {
		t.Fatal("expected bake loss error")
	}
	if !errors.Is(err, model.ErrBakeLoss) {
		t.Fatalf("expected ErrBakeLoss, got %v", err)
	}
}
