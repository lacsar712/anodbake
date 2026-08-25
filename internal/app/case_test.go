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
	cfg := config.Default("UNIT-1")
	a, err := app.BootstrapWithClock(cfg, clk)
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = a.RunCoalFeed(context.Background(), "unit1-op", 0) }()
	go func() { _ = a.RunCoalFeed(context.Background(), "unit2-op", 0) }()
	for i := 0; i < 10; i++ {
		clk.Advance(100 * time.Millisecond)
		time.Sleep(5 * time.Millisecond)
	}
	if err := a.StartSoak(context.Background(), "unit1-op"); err != nil {
		t.Fatal(err)
	}
	if err := a.Shutdown(context.Background(), "unit1-op"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		clk.Advance(100 * time.Millisecond)
		time.Sleep(5 * time.Millisecond)
	}
	mid := a.CoalFeedTPH()
	for i := 0; i < 5; i++ {
		clk.Advance(100 * time.Millisecond)
		time.Sleep(5 * time.Millisecond)
	}
	after := a.CoalFeedTPH()
	if after <= mid {
		t.Fatalf("unit2 coal feed should continue after unit1 shutdown: mid=%v after=%v", mid, after)
	}
}
