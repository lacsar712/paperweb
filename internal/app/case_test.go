package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lacsar712/paperweb/internal/app"
	"github.com/lacsar712/paperweb/internal/clock"
	"github.com/lacsar712/paperweb/internal/config"
	"github.com/lacsar712/paperweb/internal/model"
)

func TestCase(t *testing.T) {
	clk := clock.NewManual(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := config.Default("FLAME-1")
	a, err := app.BootstrapWithClock(cfg, clk)
	if err != nil {
		t.Fatal(err)
	}
	comb := a.Snapshot().Headbox
	comb.BurnerPhase = model.BurnerStable
	comb.WireframeTempF = 400
	if err := a.Store().UpdateHeadbox(cfg.UnitID, comb); err != nil {
		t.Fatal(err)
	}
	err = a.OnSheetLoss(context.Background(), "maint-op")
	if err == nil {
		t.Fatal("expected sheet loss error")
	}
	if !errors.Is(err, model.ErrSheetLoss) {
		t.Fatalf("expected ErrSheetLoss, got %v", err)
	}
}
