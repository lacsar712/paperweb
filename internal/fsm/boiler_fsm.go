package fsm

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/paperweb/internal/model"
)

type WirelineFSM struct {
	mu            sync.RWMutex
	state         model.PlantState
	fiberPermissive bool
	dewaterComplete  bool
	hooks          *HookChain
}

func NewWirelineFSM(unitID string) *WirelineFSM {
	_ = unitID
	return &WirelineFSM{state: model.StateColdStandby, hooks: NewHookChain()}
}

func (f *WirelineFSM) Hooks() *HookChain { return f.hooks }

func (f *WirelineFSM) State() model.PlantState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

func (f *WirelineFSM) SetFiberPermissive(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fiberPermissive = ok
}

func (f *WirelineFSM) SetDewaterComplete(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dewaterComplete = ok
}

func (f *WirelineFSM) FiberPermissive() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.fiberPermissive
}

func (f *WirelineFSM) Dispatch(ctx context.Context, event PlantEvent) (model.PlantState, error) {
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
		// Rejected transition: state is unchanged, so after-hooks (which drive
		// downstream side effects such as the stock-pump drive pulse) must not
		// fire. Running them here would actuate the pump from standby.
		return f.state, fmt.Errorf("%s from %s: %w", event, f.state, ErrIllegalTransition)
	}
	if event == EvIgnite && !f.fiberPermissive {
		return f.state, fmt.Errorf("%w", model.ErrFiberPermissive)
	}
	if event == EvDewaterComplete && !f.dewaterComplete {
		return f.state, fmt.Errorf("%w", model.ErrDewaterIncomplete)
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

func (f *WirelineFSM) ForceState(state model.PlantState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}
