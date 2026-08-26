package model

import (
	"math"
	"testing"
	"time"
)

func TestEchoValidationRejectsNonFiniteAndUnorderedMeasurements(t *testing.T) {
	base := EchoWindow{
		BatchID:       1,
		PingSeq:       1,
		Timestamp:     time.Unix(10, 0).UTC(),
		SoundVelocity: 1500,
		SlantRange:    40,
		Channels: []EchoChannel{{
			FrequencyHz: 200000,
			Depths:      []float64{0, 1, 2},
			Amplitudes:  []float64{-10, -11, -12},
		}},
	}
	bad := base
	bad.Attitude.Pitch = math.NaN()
	if err := bad.Validate(); err != ErrInvalidAttitude {
		t.Fatalf("expected invalid attitude, got %v", err)
	}
	bad = base
	bad.Channels[0].Depths = []float64{0, 2, 1}
	if err := bad.Validate(); err != ErrEmptyChannels {
		t.Fatalf("expected invalid depth profile, got %v", err)
	}
}
