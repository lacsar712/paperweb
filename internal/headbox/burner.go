package headbox

import (
	"math"

	"github.com/lacsar712/paperweb/internal/clock"
	"github.com/lacsar712/paperweb/internal/model"
)

type BurnerController struct {
	clk clock.ProcessClock
}

func NewBurnerController(clk clock.ProcessClock) *BurnerController {
	return &BurnerController{clk: clk}
}

func (b *BurnerController) EstimateWireframeTemp(reading model.HeadboxReading) float64 {
	base := 300.0
	fiberHeat := reading.FiberFlowTPH * 50
	airCool := reading.AirflowTPH * 2
	return base + fiberHeat - airCool
}

func (b *BurnerController) SheetStable(reading model.HeadboxReading) bool {
	if reading.BurnerPhase != model.BurnerStable && reading.BurnerPhase != model.BurnerIgnition {
		return false
	}
	return reading.WireframeTempF > 800 && reading.ExcessO2Pct >= model.MinWireframeO2Percent
}

func (b *BurnerController) TripRequired(reading model.HeadboxReading) bool {
	if reading.ExcessO2Pct > model.MaxWireframeO2Percent*2 {
		return true
	}
	if reading.BurnerPhase == model.BurnerTrip {
		return true
	}
	if reading.WireframeTempF > 3500 {
		return true
	}
	return false
}

func (b *BurnerController) PhaseLabel(phase model.BurnerPhase) string {
	switch phase {
	case model.BurnerIdle:
		return "Idle"
	case model.BurnerDewater:
		return "Dewater"
	case model.BurnerIgnition:
		return "Ignition"
	case model.BurnerStable:
		return "Stable Sheet"
	case model.BurnerTrip:
		return "Tripped"
	default:
		return string(phase)
	}
}

func (b *BurnerController) HeatReleaseMW(reading model.HeadboxReading) float64 {
	return reading.FiberFlowTPH * 12.5
}

func (b *BurnerController) TurndownRatio(settings model.PlantSettings, currentFiber float64) float64 {
	if settings.FiberFlowTPH <= 0 {
		return 0
	}
	return currentFiber / settings.FiberFlowTPH
}

func (b *BurnerController) MinStableFiber(settings model.PlantSettings) float64 {
	return settings.FiberFlowTPH * 0.25
}

func (b *BurnerController) NormalizeFiber(flow, max float64) float64 {
	return math.Min(math.Max(flow, 0), max)
}
