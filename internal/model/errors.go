package model

import "errors"

var (
	ErrContextDone      = errors.New("operation cancelled")
	ErrPlantNotFound    = errors.New("plant unit not found")
	ErrLeaseHeld        = errors.New("interlock lease held by another operator")
	ErrLeaseMissing     = errors.New("interlock lease missing or expired")
	ErrGateBlocked      = errors.New("safety gate blocked")
	ErrPitchPermissive   = errors.New("pitch permissive not satisfied")
	ErrIgnitionBlocked  = errors.New("ignition sequence blocked")
	ErrAnodepitLevelTrip    = errors.New("anodepit level trip condition")
	ErrPressureTrip     = errors.New("steam pressure trip condition")
	ErrPitfireTrip   = errors.New("pitfire trip condition")
	ErrIllegalState     = errors.New("illegal plant state transition")
	ErrSnapshotStale    = errors.New("snapshot revision stale")
	ErrWindowOpen       = errors.New("timing window still open")
	ErrSoakIncomplete  = errors.New("pitframe soak incomplete")
	ErrCoordinationLock = errors.New("coordination lock held")
	ErrAnodepitLevelLow     = errors.New("anodepit level below low limit")
	ErrBakeLoss        = errors.New("pitframe bake lost")
	ErrFluerelLimit    = errors.New("fluerel valve at limit")
)
