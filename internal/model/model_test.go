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

func TestSubstrateValidateRejectsNonFiniteAndNonPositive(t *testing.T) {
	good := func() SubstrateType {
		return SubstrateType{
			Code:     "X",
			Centroid: []float64{-12, 2, 2.5, 0.05},
			CovDiag:  []float64{1.0, 0.5, 0.4, 0.005},
		}
	}
	cases := []struct {
		name  string
		mut   func(s SubstrateType) SubstrateType
	}{
		{"zero covariance", func(s SubstrateType) SubstrateType { s.CovDiag[1] = 0; return s }},
		{"negative covariance", func(s SubstrateType) SubstrateType { s.CovDiag[2] = -0.4; return s }},
		{"NaN covariance", func(s SubstrateType) SubstrateType { s.CovDiag[0] = math.NaN(); return s }},
		{"+Inf covariance", func(s SubstrateType) SubstrateType { s.CovDiag[3] = math.Inf(1); return s }},
		{"-Inf covariance", func(s SubstrateType) SubstrateType { s.CovDiag[0] = math.Inf(-1); return s }},
		{"NaN centroid", func(s SubstrateType) SubstrateType { s.Centroid[0] = math.NaN(); return s }},
		{"+Inf centroid", func(s SubstrateType) SubstrateType { s.Centroid[2] = math.Inf(1); return s }},
		{"wrong centroid dim", func(s SubstrateType) SubstrateType { s.Centroid = []float64{1, 2}; return s }},
		{"wrong cov dim", func(s SubstrateType) SubstrateType { s.CovDiag = []float64{1, 2, 3}; return s }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := c.mut(good())
			if err := s.Validate(); err != ErrUnknownSubstrate {
				t.Fatalf("expected ErrUnknownSubstrate for %s, got %v", c.name, err)
			}
		})
	}
	// Sanity: the unmutated model is accepted.
	g := good()
	if err := g.Validate(); err != nil {
		t.Fatalf("expected clean model to validate, got %v", err)
	}
}
