package store_test

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"task250-biosonar/internal/model"
	"task250-biosonar/internal/store"
)

func TestBug05_SubSecondTimestampRegression(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	bid, err := st.CreateBatch(&model.SurveyBatch{Name: "subsecond"})
	if err != nil {
		t.Fatal(err)
	}
	base := time.Unix(100, 900000000).UTC()
	makeEcho := func(seq int, ts time.Time) *model.EchoWindow {
		return &model.EchoWindow{BatchID: bid, PingSeq: seq, Timestamp: ts, SoundVelocity: 1500, SlantRange: 20, Channels: []model.EchoChannel{{FrequencyHz: 200000, Depths: []float64{0, 1}, Amplitudes: []float64{-10, -11}}}}
	}
	if _, err := st.InsertEcho(makeEcho(1, base)); err != nil {
		t.Fatal(err)
	}
	if _, err := st.InsertEcho(makeEcho(2, base.Add(-time.Nanosecond))); !errors.Is(err, model.ErrTimeRegress) {
		t.Fatalf("expected time regression, got %v", err)
	}
}
