package model

import (
	"math"
	"time"
)

// EchoStatus tracks a single ping from raw capture to final disposition.
//
// raw -> corrected -> (attitude_anomaly | excluded)
//
// "corrected" means geometry has been reconciled against vessel attitude.
// "attitude_anomaly" and "excluded" are terminal disposals.
type EchoStatus string

const (
	EchoRaw          EchoStatus = "raw"
	EchoCorrected    EchoStatus = "corrected"
	EchoAttitudeAnom EchoStatus = "attitude_anomaly"
	EchoExcluded     EchoStatus = "excluded"
)

var echoTransitions = map[EchoStatus][]EchoStatus{
	EchoRaw:          {EchoCorrected, EchoAttitudeAnom, EchoExcluded},
	EchoCorrected:    {EchoExcluded, EchoAttitudeAnom},
	EchoAttitudeAnom: {EchoExcluded},
	EchoExcluded:     {},
}

// CanTransitionEcho reports whether an echo window may move from->to.
func CanTransitionEcho(from, to EchoStatus) bool {
	for _, n := range echoTransitions[from] {
		if n == to {
			return true
		}
	}
	return false
}

// Attitude is the vessel motion state at the instant a ping was fired.
type Attitude struct {
	Pitch   float64 `json:"pitch_deg"`   // fore-aft tilt, degrees
	Roll    float64 `json:"roll_deg"`    // port-starboard tilt, degrees
	Heading float64 `json:"heading_deg"` // compass heading, degrees
	Heave   float64 `json:"heave_m"`     // vertical displacement, metres
}

// EchoChannel is one frequency band's amplitude-vs-range return.
type EchoChannel struct {
	FrequencyHz  float64   `json:"frequency_hz"`
	IncidenceDeg float64   `json:"incidence_deg"`
	Depths       []float64 `json:"depths_m"`
	Amplitudes   []float64 `json:"amplitudes_db"`
}

// EchoWindow is a single ping's multispectral return, the core measurement
// entity of the service.
type EchoWindow struct {
	ID            int64         `json:"id"`
	BatchID       int64         `json:"batch_id"`
	PingSeq       int           `json:"ping_seq"`
	PosX          float64       `json:"pos_x_m"`
	PosY          float64       `json:"pos_y_m"`
	Timestamp     time.Time     `json:"timestamp"`
	Attitude      Attitude      `json:"attitude"`
	SoundVelocity float64       `json:"sound_velocity_mps"`
	SlantRange    float64       `json:"slant_range_m"`
	Status        EchoStatus    `json:"status"`
	Channels      []EchoChannel `json:"channels"`

	// Corrected geometry, filled after the correction step.
	CorrectedX     float64    `json:"corrected_x_m"`
	CorrectedY     float64    `json:"corrected_y_m"`
	CorrectedDepth float64    `json:"corrected_depth_m"`
	CorrectedAt    *time.Time `json:"corrected_at,omitempty"`
}

// Validate performs the local invariant checks that do not require the store.
func (e *EchoWindow) Validate() error {
	if e.Timestamp.IsZero() {
		return ErrInvalidTimestamp
	}
	// The origin (0,0) is a legitimate local survey coordinate, so only
	// non-finite values (uninitialised / corrupt) are rejected.
	if math.IsNaN(e.PosX) || math.IsNaN(e.PosY) ||
		math.IsInf(e.PosX, 0) || math.IsInf(e.PosY, 0) {
		return ErrInvalidPosition
	}
	if !finite(e.SoundVelocity) || e.SoundVelocity <= 0 {
		return ErrInvalidAttitude
	}
	if !finite(e.SlantRange) || e.SlantRange <= 0 {
		return ErrInvalidAttitude
	}
	if !finite(e.Attitude.Pitch) || !finite(e.Attitude.Roll) ||
		!finite(e.Attitude.Heading) || !finite(e.Attitude.Heave) {
		return ErrInvalidAttitude
	}
	if len(e.Channels) == 0 {
		return ErrEmptyChannels
	}
	for _, c := range e.Channels {
		if len(c.Depths) != len(c.Amplitudes) || len(c.Depths) == 0 {
			return ErrEmptyChannels
		}
		if !finite(c.FrequencyHz) || c.FrequencyHz <= 0 || !finite(c.IncidenceDeg) {
			return ErrInvalidAttitude
		}
		for i, depth := range c.Depths {
			if !finite(depth) || !finite(c.Amplitudes[i]) {
				return ErrEmptyChannels
			}
			// Each channel's samples must be strictly increasing down the
			// water column: an inverted or repeated profile makes the
			// penetration depth (and thus the substrate verdict) depend on
			// input ordering, so an illegal profile is rejected at ingestion
			// rather than silently classified.
			if i > 0 && depth <= c.Depths[i-1] {
				return ErrEmptyChannels
			}
		}
	}
	return nil
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
