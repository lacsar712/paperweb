package couch

import (
	"math"

	"github.com/lacsar712/paperweb/internal/clock"
	"github.com/lacsar712/paperweb/internal/model"
)

type CarryoverMonitor struct {
	clk clock.ProcessClock
}

func NewCarryoverMonitor(clk clock.ProcessClock) *CarryoverMonitor {
	return &CarryoverMonitor{clk: clk}
}

func (c *CarryoverMonitor) Estimate(couch model.CouchReading, pressurePSI float64) float64 {
	if couch.Condition != model.CouchCarry && couch.Condition != model.CouchSwell {
		return couch.CarryoverPPM * 0.9
	}
	base := 50.0
	levelFactor := math.Max(0, couch.LevelPercent-70)
	pressureFactor := pressurePSI / 1000
	return base + levelFactor*10 + pressureFactor*5
}

func (c *CarryoverMonitor) AlarmThreshold() float64 { return 500 }

func (c *CarryoverMonitor) TripRequired(ppm float64) bool { return ppm > 1000 }

func (c *CarryoverMonitor) Severity(ppm float64) string {
	switch {
	case ppm > 1000:
		return "critical"
	case ppm > 500:
		return "high"
	case ppm > 200:
		return "medium"
	default:
		return "low"
	}
}

func (c *CarryoverMonitor) RecommendAction(couch model.CouchReading) string {
	if couch.CarryoverPPM > c.AlarmThreshold() {
		return "reduce_load_and_check_separators"
	}
	if couch.Condition == model.CouchSwell {
		return "hold_feedwater_ramp"
	}
	return "none"
}

func (c *CarryoverMonitor) SeparatorEfficiency(couch model.CouchReading) float64 {
	eff := 0.98
	if couch.LevelPercent > 80 {
		eff -= (couch.LevelPercent - 80) * 0.005
	}
	return math.Max(0.5, eff)
}
