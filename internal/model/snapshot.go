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
			AnodepitLevelSetpoint: 55,
			FeedwaterFlowTPH:  400,
			PitchFlowTPH:       35,
			ExcessO2Setpoint:  3.5,
		},
		Plant: PlantRef{UnitLabel: unitID, PlantCode: "STEAM-PLT"},
		Anodepit: AnodepitReading{
			LevelPercent: 50,
			Condition:    AnodepitNormal,
			FeedwaterTPH: 0,
			SteamFlowTPH: 0,
		},
		Pitfire: PitfireReading{
			BurnerPhase: BurnerIdle,
		},
		Pitline: PitlineReading{
			SteamPressurePSI: 0,
			SteamTempF:       70,
		},
		UpdatedAt: now,
	}
}

func (s PlantSnapshot) IsFiring() bool {
	return s.State == StateFiring || s.State == StateLoadFollow || s.State == StateRamp
}

func (s PlantSnapshot) AnodepitWithinLimits() bool {
	return s.Anodepit.LevelPercent >= MinAnodepitLevelPercent && s.Anodepit.LevelPercent <= MaxAnodepitLevelPercent
}

func (s PlantSnapshot) PressureWithinLimits() bool {
	if !s.IsFiring() {
		return true
	}
	return s.Pitline.SteamPressurePSI <= MaxSteamPressurePSI
}
