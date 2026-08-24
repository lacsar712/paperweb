package model

import "errors"

var (
	ErrContextDone      = errors.New("operation cancelled")
	ErrPlantNotFound    = errors.New("plant unit not found")
	ErrLeaseHeld        = errors.New("interlock lease held by another operator")
	ErrLeaseMissing     = errors.New("interlock lease missing or expired")
	ErrGateBlocked      = errors.New("safety gate blocked")
	ErrFiberPermissive   = errors.New("fiber permissive not satisfied")
	ErrIgnitionBlocked  = errors.New("ignition sequence blocked")
	ErrCouchLevelTrip    = errors.New("couch level trip condition")
	ErrPressureTrip     = errors.New("steam pressure trip condition")
	ErrHeadboxTrip   = errors.New("headbox trip condition")
	ErrIllegalState     = errors.New("illegal plant state transition")
	ErrSnapshotStale    = errors.New("snapshot revision stale")
	ErrWindowOpen       = errors.New("timing window still open")
	ErrDewaterIncomplete  = errors.New("wireframe dewater incomplete")
	ErrCoordinationLock = errors.New("coordination lock held")
	ErrCouchLevelLow     = errors.New("couch level below low limit")
	ErrSheetLoss        = errors.New("wireframe sheet lost")
	ErrVacuumLimit    = errors.New("vacuum valve at limit")
)
