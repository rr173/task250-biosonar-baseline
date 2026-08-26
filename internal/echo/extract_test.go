package echo

import (
	"testing"

	"task250-biosonar/internal/model"
)

func TestExtractShape(t *testing.T) {
	ch := model.EchoChannel{
		FrequencyHz: 200000,
		IncidenceDeg: 30,
		Depths:      []float64{0, 0.5, 1.0, 1.5, 2.0},
		Amplitudes:  []float64{-10, -10.2, -10.5, -13, -20},
	}
	e := &model.EchoWindow{Channels: []model.EchoChannel{ch}}
	fv, err := Extract(e)
	if err != nil {
		t.Fatal(err)
	}
	if len(fv) != model.FeatureDim {
		t.Fatalf("want dim %d got %d", model.FeatureDim, len(fv))
	}
	if fv[0] > -9 || fv[0] < -13 {
		t.Fatalf("bs_mean out of expected range: %f", fv[0])
	}
	if fv[2] <= 0 {
		t.Fatalf("penetration should be positive, got %f", fv[2])
	}
}

func TestExtractEmpty(t *testing.T) {
	if _, err := Extract(&model.EchoWindow{}); err == nil {
		t.Fatal("expected error for empty channels")
	}
}
