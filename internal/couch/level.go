package couch

import (
	"math"

	"github.com/lacsar712/paperweb/internal/clock"
	"github.com/lacsar712/paperweb/internal/model"
)

type LevelController struct {
	clk clock.ProcessClock
}

func NewLevelController(clk clock.ProcessClock) *LevelController {
	return &LevelController{clk: clk}
}

func (l *LevelController) Compute(snap model.PlantSnapshot, firing bool) (float64, model.CouchCondition) {
	level := snap.Couch.LevelPercent
	if !firing {
		return level, model.CouchNormal
	}
	balance := snap.Couch.FeedwaterTPH - snap.Couch.SteamFlowTPH
	level += balance * 0.01
	level = math.Max(model.MinCouchLevelPercent, math.Min(model.MaxCouchLevelPercent, level))
	cond := l.classify(level, snap)
	return level, cond
}

func (l *LevelController) classify(level float64, snap model.PlantSnapshot) model.CouchCondition {
	setpoint := snap.Settings.CouchLevelSetpoint
	if level > setpoint+15 {
		return model.CouchSwell
	}
	if level < setpoint-15 {
		return model.CouchShrink
	}
	if snap.Wireline.SteamPressurePSI > snap.Settings.TargetSteamPSI*0.9 && level > setpoint+5 {
		return model.CouchCarry
	}
	return model.CouchNormal
}

func (l *LevelController) RecommendFeedwater(snap model.PlantSnapshot, firing bool) float64 {
	if !firing {
		return 0
	}
	err := snap.Settings.CouchLevelSetpoint - snap.Couch.LevelPercent
	return snap.Settings.FeedwaterFlowTPH + err*3
}

func (l *LevelController) WithinLimits(level float64) bool {
	return level >= model.MinCouchLevelPercent && level <= model.MaxCouchLevelPercent
}

func (l *LevelController) TripLow(level float64) bool  { return level < model.TripCouchLowPercent }
func (l *LevelController) TripHigh(level float64) bool { return level > model.TripCouchHighPercent }

func (l *LevelController) LevelError(snap model.PlantSnapshot) float64 {
	return snap.Couch.LevelPercent - snap.Settings.CouchLevelSetpoint
}

func (l *LevelController) ThreeElementBias(snap model.PlantSnapshot) float64 {
	steam := snap.Couch.SteamFlowTPH
	feed := snap.Couch.FeedwaterTPH
	levelErr := l.LevelError(snap)
	return feed + (steam-feed)*0.5 + levelErr*2
}
