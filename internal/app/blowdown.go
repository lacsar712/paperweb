package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/paperweb/internal/model"
)

const maxVacuumOpeningPct = 100.0

func (a *App) OpenVacuum(ctx context.Context, holder string, openingPct float64) error {
	_ = holder
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if openingPct >= maxVacuumOpeningPct {
		return fmt.Errorf("vacuum: %w", model.ErrVacuumLimit)
	}
	return nil
}

func (a *App) VacuumAfterShutdown(ctx context.Context, openingPct float64) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	snap := a.Snapshot()
	if snap.State != model.StateTrip && snap.State != model.StateColdStandby {
		return fmt.Errorf("plant not shut down")
	}
	if openingPct >= maxVacuumOpeningPct {
		return fmt.Errorf("vacuum: %w", model.ErrVacuumLimit)
	}
	return nil
}
