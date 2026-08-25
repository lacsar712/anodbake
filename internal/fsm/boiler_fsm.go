package fsm

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/anodbake/internal/model"
)

type PitlineFSM struct {
	mu            sync.RWMutex
	state         model.PlantState
	pitchPermissive bool
	soakComplete  bool
	hooks          *HookChain
}

func NewPitlineFSM(unitID string) *PitlineFSM {
	_ = unitID
	return &PitlineFSM{state: model.StateColdStandby, hooks: NewHookChain()}
}

func (f *PitlineFSM) Hooks() *HookChain { return f.hooks }

func (f *PitlineFSM) State() model.PlantState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

func (f *PitlineFSM) SetPitchPermissive(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pitchPermissive = ok
}

func (f *PitlineFSM) SetSoakComplete(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.soakComplete = ok
}

func (f *PitlineFSM) PitchPermissive() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.pitchPermissive
}

func (f *PitlineFSM) Dispatch(ctx context.Context, event PlantEvent) (model.PlantState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	select {
	case <-ctx.Done():
		return f.state, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if event == EvTrip {
		from := f.state
		if f.hooks != nil {
			if err := f.hooks.RunBefore(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		f.state = model.StateTrip
		if f.hooks != nil {
			if err := f.hooks.RunAfter(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		return f.state, nil
	}
	next, ok := NextState(f.state, event)
	if !ok {
		// Rejected transitions must not trigger after-hook side effects
		// (e.g. the pitframe drive pulse); the state never changed, so no
		// accepted transition occurred. Returning here leaves the plant in
		// its current state without poking the execution chain.
		return f.state, fmt.Errorf("%s from %s: %w", event, f.state, ErrIllegalTransition)
	}
	if event == EvIgnite && !f.pitchPermissive {
		return f.state, fmt.Errorf("%w", model.ErrPitchPermissive)
	}
	if event == EvSoakComplete && !f.soakComplete {
		return f.state, fmt.Errorf("%w", model.ErrSoakIncomplete)
	}
	from := f.state
	if f.hooks != nil {
		if err := f.hooks.RunBefore(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	f.state = next
	if f.hooks != nil {
		if err := f.hooks.RunAfter(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	return f.state, nil
}

func (f *PitlineFSM) ForceState(state model.PlantState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}
