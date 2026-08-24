package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/paperweb/internal/clock"
	"github.com/lacsar712/paperweb/internal/model"
)

func (a *App) advanceClock(d time.Duration) {
	if mc, ok := a.clk.(*clock.ManualClock); ok {
		mc.Advance(d)
		time.Sleep(time.Millisecond)
	} else {
		time.Sleep(d)
	}
}

func (a *App) bindFiberLoop(holder string, ctx context.Context) context.Context {
	a.mu.Lock()
	if cancel, ok := a.fiberLoopCancels[holder]; ok {
		cancel()
	}
	child, cancel := context.WithCancel(ctx)
	a.fiberLoopCancels[holder] = cancel
	a.mu.Unlock()
	return child
}

func (a *App) cancelFiberLoop(holder string) {
	a.mu.Lock()
	if cancel, ok := a.fiberLoopCancels[holder]; ok {
		cancel()
		delete(a.fiberLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) cancelAllFiberLoops() {
	a.mu.Lock()
	for holder, cancel := range a.fiberLoopCancels {
		cancel()
		delete(a.fiberLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) CoalFeedTPH() float64 {
	return a.Snapshot().Headbox.FiberFlowTPH
}

func (a *App) RunFiberRamp(ctx context.Context, holder string, targetTPH float64) error {
	loopCtx := a.bindFiberLoop(holder, ctx)
	defer a.cancelFiberLoop(holder)
	for {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		current := snap.Headbox.FiberFlowTPH
		if current >= targetTPH {
			return nil
		}
		comb := snap.Headbox
		comb.FiberFlowTPH = current + 1.0
		_ = a.store.UpdateHeadbox(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.FiberFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
}

func (a *App) RunCoalFeed(ctx context.Context, holder string, steps int) error {
	loopCtx := a.bindFiberLoop(holder, ctx)
	defer a.cancelFiberLoop(holder)
	for i := 0; steps <= 0 || i < steps; i++ {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		comb := snap.Headbox
		comb.FiberFlowTPH += 0.5
		_ = a.store.UpdateHeadbox(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.FiberFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
	return nil
}
