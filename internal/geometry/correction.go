// Package geometry reconciles raw slant-range pings against vessel attitude,
// sound velocity and heave to produce the true seafloor sample point.
package geometry

import (
	"math"
	"time"

	"task250-biosonar/internal/model"
)

func deg2rad(d float64) float64 { return d * math.Pi / 180 }

// Default attitude limits beyond which a ping is flagged as an attitude
// anomaly rather than corrected normally.
const (
	DefaultPitchLimitDeg = 12.0
	DefaultRollLimitDeg  = 12.0
)

// IsAttitudeAnomaly reports whether the vessel attitude at ping time is too
// unstable to trust the geometric correction.
func IsAttitudeAnomaly(a model.Attitude, pitchLimit, rollLimit float64) bool {
	return math.Abs(a.Pitch) > pitchLimit || math.Abs(a.Roll) > rollLimit
}

// Correct computes the seafloor sample point implied by an echo window's
// slant range, attitude, sound velocity and heave. It mutates the window's
// corrected fields and stamps CorrectedAt. The caller is responsible for the
// raw -> corrected status transition.
func Correct(e *model.EchoWindow) error {
	if e == nil || math.IsNaN(e.SlantRange) || math.IsInf(e.SlantRange, 0) ||
		e.SlantRange <= 0 || math.IsNaN(e.SoundVelocity) || math.IsInf(e.SoundVelocity, 0) ||
		e.SoundVelocity <= 0 || math.IsNaN(e.Attitude.Pitch) || math.IsInf(e.Attitude.Pitch, 0) ||
		math.IsNaN(e.Attitude.Roll) || math.IsInf(e.Attitude.Roll, 0) ||
		math.IsNaN(e.Attitude.Heading) || math.IsInf(e.Attitude.Heading, 0) ||
		math.IsNaN(e.Attitude.Heave) || math.IsInf(e.Attitude.Heave, 0) {
		return model.ErrInvalidAttitude
	}
	pr := deg2rad(e.Attitude.Pitch)
	rr := deg2rad(e.Attitude.Roll)

	// Depth below the transducer after tilt correction; heave shifts the
	// transducer vertically relative to the mean sea surface.
	depth := e.SlantRange*math.Cos(pr)*math.Cos(rr) + e.Attitude.Heave

	// Horizontal offsets induced by vessel tilt.
	across := e.SlantRange * math.Sin(rr)
	along := e.SlantRange * math.Sin(pr)

	// Rotate the horizontal offset into the survey-frame x/y using heading.
	hr := deg2rad(e.Attitude.Heading)
	dx := across*math.Cos(hr) - along*math.Sin(hr)
	dy := across*math.Sin(hr) + along*math.Cos(hr)

	e.CorrectedX = e.PosX + dx
	e.CorrectedY = e.PosY + dy
	e.CorrectedDepth = depth

	now := time.Now().UTC()
	e.CorrectedAt = &now
	return nil
}

// OffsetMagnitude returns the horizontal displacement introduced by attitude
// correction, useful for diagnostics and anomaly evidence.
func OffsetMagnitude(e *model.EchoWindow) float64 {
	dx := e.CorrectedX - e.PosX
	dy := e.CorrectedY - e.PosY
	return math.Sqrt(dx*dx + dy*dy)
}
