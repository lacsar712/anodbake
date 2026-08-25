package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/anodbake/internal/app"
	"github.com/lacsar712/anodbake/internal/clock"
	"github.com/lacsar712/anodbake/internal/config"
)

func TestCase(t *testing.T) {
	clk := clock.NewManual(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := config.Default("COAL-1")
	a, err := app.BootstrapWithClock(cfg, clk)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.StartSoak(ctx, "shift-op"); err != nil {
		t.Fatal(err)
	}
	go func() { _ = a.RunCoalFeed(ctx, "shift-op", 0) }()
	var before float64
	for i := 0; i < 30; i++ {
		clk.Advance(100 * time.Millisecond)
		time.Sleep(5 * time.Millisecond)
		before = a.CoalFeedTPH()
		if before > 0 {
			break
		}
	}
	if before == 0 {
		t.Fatal("coal feed never started")
	}
	if err := a.Shutdown(ctx, "shift-op"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		clk.Advance(100 * time.Millisecond)
		time.Sleep(5 * time.Millisecond)
	}
	after := a.CoalFeedTPH()
	if after > before {
		t.Fatalf("coal feed continued after shutdown: before=%v after=%v", before, after)
	}
}
