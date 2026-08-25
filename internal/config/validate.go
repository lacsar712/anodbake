package config

import (
	"fmt"
	"strings"

	"github.com/lacsar712/anodbake/internal/model"
)

func Validate(cfg Config) error {
	if strings.TrimSpace(cfg.UnitID) == "" {
		return fmt.Errorf("unit_id required")
	}
	if cfg.ListenAddr == "" {
		return fmt.Errorf("listen_addr required")
	}
	if err := validateSettings(cfg.Settings); err != nil {
		return err
	}
	return nil
}

func validateSettings(s model.PlantSettings) error {
	if s.TargetMW < 0 {
		return fmt.Errorf("target_mw cannot be negative")
	}
	if s.TargetSteamPSI <= 0 {
		return fmt.Errorf("target_steam_psi must be positive")
	}
	if s.AnodepitLevelSetpoint < model.MinAnodepitLevelPercent || s.AnodepitLevelSetpoint > model.MaxAnodepitLevelPercent {
		return fmt.Errorf("anodepit level setpoint out of range")
	}
	if s.PitchFlowTPH <= 0 {
		return fmt.Errorf("pitch_flow_tph must be positive")
	}
	if s.ExcessO2Setpoint < model.MinPitframeO2Percent || s.ExcessO2Setpoint > model.MaxPitframeO2Percent {
		return fmt.Errorf("excess_o2 setpoint out of range")
	}
	return nil
}
