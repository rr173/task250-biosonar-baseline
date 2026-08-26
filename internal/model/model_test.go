package model

import (
	"math"
	"testing"
	"time"
)

func TestEchoValidationRejectsNonFiniteAndUnorderedMeasurements(t *testing.T) {
	mk := func(depths, amps []float64) *EchoWindow {
		return &EchoWindow{
			BatchID:       1,
			PingSeq:       1,
			Timestamp:     time.Unix(10, 0).UTC(),
			SoundVelocity: 1500,
			SlantRange:    40,
			Channels: []EchoChannel{{
				FrequencyHz: 200000,
				Depths:      depths,
				Amplitudes:  amps,
			}},
		}
	}

	bad := mk([]float64{0, 1, 2}, []float64{-10, -11, -12})
	bad.Attitude.Pitch = math.NaN()
	if err := bad.Validate(); err != ErrInvalidAttitude {
		t.Fatalf("expected invalid attitude, got %v", err)
	}

	// An inverted depth axis makes penetration ordering-dependent.
	if err := mk([]float64{0, 2, 1}, []float64{-10, -11, -12}).Validate(); err != ErrEmptyChannels {
		t.Fatalf("expected invalid depth profile (inverted), got %v", err)
	}

	// A repeated depth is equally illegal: the axis must be strictly
	// monotonic for penetration to be well-defined.
	if err := mk([]float64{0, 1, 1, 2}, []float64{-10, -11, -12, -13}).Validate(); err != ErrEmptyChannels {
		t.Fatalf("expected invalid depth profile (repeated), got %v", err)
	}

	// The baseline strictly-increasing profile must still validate.
	if err := mk([]float64{0, 1, 2}, []float64{-10, -11, -12}).Validate(); err != nil {
		t.Fatalf("baseline profile should validate, got %v", err)
	}
}
