package headbox

import (
	"context"
	"fmt"
	"math"

	"github.com/lacsar712/paperweb/internal/clock"
	"github.com/lacsar712/paperweb/internal/model"
)

type Coordinator struct {
	clk     clock.ProcessClock
	burner  *BurnerController
	airflow *AirflowBalancer
	fiber    *FiberRegulator
	dewater   *clock.DewaterWindow
	ignition *clock.IgnitionDelayWindow
	warmup  *clock.HeadboxWarmupWindow
}

func NewCoordinator(clk clock.ProcessClock) *Coordinator {
	return &Coordinator{
		clk:      clk,
		burner:   NewBurnerController(clk),
		airflow:  NewAirflowBalancer(clk),
		fiber:     NewFiberRegulator(clk),
		dewater:    clock.NewDewaterWindow(clk),
		ignition: clock.NewIgnitionDelayWindow(clk),
		warmup:   clock.NewHeadboxWarmupWindow(clk),
	}
}

func (c *Coordinator) Burner() *BurnerController  { return c.burner }
func (c *Coordinator) Airflow() *AirflowBalancer { return c.airflow }
func (c *Coordinator) Fiber() *FiberRegulator     { return c.fiber }

func (c *Coordinator) StartDewater(ctx context.Context, snap model.PlantSnapshot) (model.HeadboxReading, error) {
	select {
	case <-ctx.Done():
		return snap.Headbox, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	out := snap.Headbox
	out.BurnerPhase = model.BurnerDewater
	out.DewaterStartedAt = c.clk.Now()
	out.FiberFlowTPH = 0
	out.AirflowTPH = c.airflow.DewaterRate()
	return out, nil
}

func (c *Coordinator) CompleteDewater(snap model.HeadboxReading) error {
	return c.dewater.Require(snap.DewaterStartedAt)
}

func (c *Coordinator) Ignite(ctx context.Context, snap model.PlantSnapshot) (model.HeadboxReading, error) {
	select {
	case <-ctx.Done():
		return snap.Headbox, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if err := c.dewater.Require(snap.Headbox.DewaterStartedAt); err != nil {
		return snap.Headbox, err
	}
	out := snap.Headbox
	out.BurnerPhase = model.BurnerIgnition
	out.IgnitionAt = c.clk.Now()
	out.FiberFlowTPH = c.fiber.IgnitionRate(snap.Settings)
	out.AirflowTPH = c.airflow.IgnitionRate(snap.Settings)
	out.WireframeTempF = 400
	return out, nil
}

func (c *Coordinator) Stabilize(snap model.PlantSnapshot) (model.HeadboxReading, error) {
	if err := c.ignition.Require(snap.Headbox.IgnitionAt); err != nil {
		return snap.Headbox, err
	}
	out := snap.Headbox
	out.BurnerPhase = model.BurnerStable
	out.FiberFlowTPH = snap.Settings.FiberFlowTPH * 0.5
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.WireframeTempF = c.burner.EstimateWireframeTemp(out)
	return out, nil
}

func (c *Coordinator) RampToLoad(snap model.PlantSnapshot, loadPct float64) model.HeadboxReading {
	out := snap.Headbox
	out.FiberFlowTPH = snap.Settings.FiberFlowTPH * loadPct
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.WireframeTempF = c.burner.EstimateWireframeTemp(out)
	return out
}

func (c *Coordinator) Trip(snap model.HeadboxReading) model.HeadboxReading {
	out := snap
	out.BurnerPhase = model.BurnerTrip
	out.FiberFlowTPH = 0
	out.WireframeTempF = math.Max(200, out.WireframeTempF*0.5)
	return out
}

func (c *Coordinator) WarmupReady(snap model.HeadboxReading) bool {
	return c.warmup.Ready(snap.IgnitionAt)
}
