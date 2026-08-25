package fsm

import (
	"context"
	"testing"
)

func TestCase(t *testing.T) {
	PitframeDrivePulse = nil
	var pulses int
	PitframeDrivePulse = func() { pulses++ }
	b := NewPitlineFSM("u1")
	RegisterPitframeDriveHook(b.Hooks())
	_, err := b.Dispatch(context.Background(), EvReachFiring)
	if err == nil {
		t.Fatal("expected illegal transition error")
	}
	if pulses != 0 {
		t.Fatalf("illegal transition should not pulse furnace drive, got %d", pulses)
	}
}
