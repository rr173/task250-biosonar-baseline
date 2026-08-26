package model

import "fmt"

// TransitionBatch validates and returns an error for an illegal batch move.
func TransitionBatch(from, to BatchStatus) error {
	if !CanTransitionBatch(from, to) {
		return fmt.Errorf("%w: batch %s -> %s", ErrInvalidTransition, from, to)
	}
	return nil
}

// TransitionEcho validates and returns an error for an illegal echo move.
func TransitionEcho(from, to EchoStatus) error {
	if !CanTransitionEcho(from, to) {
		return fmt.Errorf("%w: echo %s -> %s", ErrInvalidTransition, from, to)
	}
	return nil
}

// TransitionSegment validates and returns an error for an illegal segment move.
func TransitionSegment(from, to SegmentStatus) error {
	if !CanTransitionSegment(from, to) {
		return fmt.Errorf("%w: segment %s -> %s", ErrInvalidTransition, from, to)
	}
	return nil
}

// TransitionSnapshot validates and returns an error for an illegal snapshot move.
func TransitionSnapshot(from, to SnapshotStatus) error {
	if !CanTransitionSnapshot(from, to) {
		return fmt.Errorf("%w: snapshot %s -> %s", ErrInvalidTransition, from, to)
	}
	return nil
}
