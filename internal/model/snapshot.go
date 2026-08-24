package model

import "time"

func CloneSnapshot(s PlantSnapshot) PlantSnapshot {
	out := s
	out.Alarms = append([]AlarmEvent(nil), s.Alarms...)
	return out
}

func DefaultSnapshot(unitID string) PlantSnapshot {
	now := time.Now()
	return PlantSnapshot{
		UnitID: unitID,
		State:  StateColdStandby,
		Settings: PlantSettings{
			Mode:              ModeBaseLoad,
			TargetMW:          150,
			TargetSteamPSI:    NormalSteamPressurePSI,
			CouchLevelSetpoint: 55,
			FeedwaterFlowTPH:  400,
			FiberFlowTPH:       35,
			ExcessO2Setpoint:  3.5,
		},
		Plant: PlantRef{UnitLabel: unitID, PlantCode: "STEAM-PLT"},
		Couch: CouchReading{
			LevelPercent: 50,
			Condition:    CouchNormal,
			FeedwaterTPH: 0,
			SteamFlowTPH: 0,
		},
		Headbox: HeadboxReading{
			BurnerPhase: BurnerIdle,
		},
		Wireline: WirelineReading{
			SteamPressurePSI: 0,
			SteamTempF:       70,
		},
		UpdatedAt: now,
	}
}

func (s PlantSnapshot) IsFiring() bool {
	return s.State == StateFiring || s.State == StateLoadFollow || s.State == StateRamp
}

func (s PlantSnapshot) CouchWithinLimits() bool {
	return s.Couch.LevelPercent >= MinCouchLevelPercent && s.Couch.LevelPercent <= MaxCouchLevelPercent
}

func (s PlantSnapshot) PressureWithinLimits() bool {
	if !s.IsFiring() {
		return true
	}
	return s.Wireline.SteamPressurePSI <= MaxSteamPressurePSI
}
