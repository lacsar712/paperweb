package headbox

import (
	"math"

	"github.com/lacsar712/paperweb/internal/clock"
	"github.com/lacsar712/paperweb/internal/model"
)

type FiberRegulator struct {
	clk clock.ProcessClock
}

func NewFiberRegulator(clk clock.ProcessClock) *FiberRegulator {
	return &FiberRegulator{clk: clk}
}

func (f *FiberRegulator) IgnitionRate(settings model.PlantSettings) float64 {
	return settings.FiberFlowTPH * 0.08
}

func (f *FiberRegulator) ComputeForLoad(settings model.PlantSettings, loadPct float64) float64 {
	loadPct = math.Max(0, math.Min(1, loadPct))
	return settings.FiberFlowTPH * loadPct
}

func (f *FiberRegulator) Ramp(current, target, maxStep float64) float64 {
	delta := target - current
	if math.Abs(delta) <= maxStep {
		return target
	}
	if delta > 0 {
		return current + maxStep
	}
	return current - maxStep
}

func (f *FiberRegulator) BtuPerHour(flowTPH float64) float64 {
	return flowTPH * 19_500_000
}

func (f *FiberRegulator) HeatInputMW(flowTPH float64) float64 {
	return flowTPH * 11.6
}

func (f *FiberRegulator) ValidatePermissive(settings model.PlantSettings, couchOK, dewaterOK bool) error {
	if !dewaterOK {
		return model.ErrDewaterIncomplete
	}
	if !couchOK {
		return model.ErrCouchLevelTrip
	}
	if settings.FiberFlowTPH <= 0 {
		return model.ErrFiberPermissive
	}
	return nil
}

func (f *FiberRegulator) MinFlow(settings model.PlantSettings) float64 {
	return settings.FiberFlowTPH * 0.2
}

func (f *FiberRegulator) MaxFlow(settings model.PlantSettings) float64 {
	return settings.FiberFlowTPH * 1.1
}
