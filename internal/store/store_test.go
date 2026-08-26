package store

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"task250-biosonar/internal/model"
)

func TestEchoTimestampKeepsNanosecondOrdering(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "biosonar.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	bid, err := st.CreateBatch(&model.SurveyBatch{Name: "time-order"})
	if err != nil {
		t.Fatal(err)
	}
	base := time.Unix(100, 900000000).UTC()
	makeEcho := func(seq int, ts time.Time) *model.EchoWindow {
		return &model.EchoWindow{
			BatchID: bid, PingSeq: seq, Timestamp: ts, SoundVelocity: 1500, SlantRange: 20,
			Channels: []model.EchoChannel{{FrequencyHz: 200000, Depths: []float64{0, 1}, Amplitudes: []float64{-10, -11}}},
		}
	}
	if _, err := st.InsertEcho(makeEcho(1, base)); err != nil {
		t.Fatal(err)
	}
	if _, err := st.InsertEcho(makeEcho(2, base.Add(-time.Nanosecond))); !errors.Is(err, model.ErrTimeRegress) {
		t.Fatalf("expected nanosecond time regression, got %v", err)
	}
}
