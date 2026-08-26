// Package model defines the domain entities, status state machines and
// invariant errors for the biosonar seafloor-classification service.
package model

import "errors"

// Domain errors returned by the store and service layers.
var (
	ErrNotFound           = errors.New("resource not found")
	ErrBatchSealed        = errors.New("batch is sealed and immutable")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrEchoExcluded       = errors.New("echo window is excluded")
	ErrDuplicatePing      = errors.New("duplicate ping sequence in batch")
	ErrInvalidAttitude    = errors.New("invalid attitude values")
	ErrInvalidPosition    = errors.New("invalid position coordinates")
	ErrTimeRegress        = errors.New("echo timestamp regresses batch clock")
	ErrSegmentFinalized   = errors.New("segment already finalized")
	ErrSnapshotSuperseded = errors.New("snapshot already superseded")
	ErrEmptyChannels      = errors.New("echo window has no frequency channels")
	ErrUnknownSubstrate   = errors.New("unknown substrate type")
	ErrNotClassified      = errors.New("echo window not yet classified")
	ErrInvalidTimestamp   = errors.New("invalid echo timestamp")
	ErrSnapshotFrozen     = errors.New("published snapshot freezes segment evidence")
)
