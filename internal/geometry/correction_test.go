package geometry

import (
	"testing"

	"task250-biosonar/internal/model"
)

func TestCorrectZeroAttitude(t *testing.T) {
	e := &model.EchoWindow{
		PosX:          100,
		PosY:          50,
		SlantRange:    50,
		SoundVelocity: 1500,
		Attitude:      model.Attitude{Pitch: 0, Roll: 0, Heading: 0, Heave: 0},
	}
	if err := Correct(e); err != nil {
		t.Fatal(err)
	}
	if e.CorrectedDepth != 50 {
		t.Fatalf("depth want 50 got %f", e.CorrectedDepth)
	}
	if e.CorrectedX != 100 || e.CorrectedY != 50 {
		t.Fatalf("pos want 100,50 got %f,%f", e.CorrectedX, e.CorrectedY)
	}
	if e.CorrectedAt == nil {
		t.Fatal("CorrectedAt not stamped")
	}
}

func TestIsAttitudeAnomaly(t *testing.T) {
	ok := IsAttitudeAnomaly(model.Attitude{Pitch: 20, Roll: 1}, DefaultPitchLimitDeg, DefaultRollLimitDeg)
	if !ok {
		t.Fatal("expected anomaly for pitch 20")
	}
	ok = IsAttitudeAnomaly(model.Attitude{Pitch: 1, Roll: 1}, DefaultPitchLimitDeg, DefaultRollLimitDeg)
	if ok {
		t.Fatal("did not expect anomaly for small attitude")
	}
}
