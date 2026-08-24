package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/paperweb/internal/model"
)

func (a *App) WarmupStatus() (ready bool, detail string) {
	snap := a.Snapshot()
	if snap.Headbox.DewaterStartedAt.IsZero() {
		return false, "dewater not started"
	}
	if !a.dewaterWindow.Ready(snap.Headbox.DewaterStartedAt) {
		return false, "dewater window open"
	}
	if !snap.Headbox.IgnitionAt.IsZero() && !a.warmupWindow.Ready(snap.Headbox.IgnitionAt) {
		return false, "headbox warmup window open"
	}
	if !snap.Couch.LastSwellAt.IsZero() {
		if err := a.couch.RequireSettled(snap.Couch); err != nil {
			return false, "couch swell settling"
		}
	}
	return true, "ready"
}

func (a *App) WaitWarmup(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w", model.ErrContextDone)
		default:
		}
		ready, _ := a.WarmupStatus()
		if ready {
			return nil
		}
	}
}

func (a *App) DewaterRemaining() string {
	snap := a.Snapshot()
	if snap.Headbox.DewaterStartedAt.IsZero() {
		return "not started"
	}
	if a.dewaterWindow.Ready(snap.Headbox.DewaterStartedAt) {
		return "complete"
	}
	return "in progress"
}

func (a *App) HeadboxWarmupRemaining() string {
	snap := a.Snapshot()
	if snap.Headbox.IgnitionAt.IsZero() {
		return "not ignited"
	}
	if a.warmupWindow.Ready(snap.Headbox.IgnitionAt) {
		return "complete"
	}
	return "in progress"
}
