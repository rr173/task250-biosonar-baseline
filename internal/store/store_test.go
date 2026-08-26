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

// TestEchoRejectsSameSecondRegression locks the invariant that when two pings
// arrive within the same calendar second with the later one received first, the
// earlier one must still be rejected as a time regression. A naive storage that
// truncates timestamps to whole seconds would collapse both into the same
// value and miss the regression. Equal timestamps remain permitted.
func TestEchoRejectsSameSecondRegression(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "biosonar.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	bid, err := st.CreateBatch(&model.SurveyBatch{Name: "same-second"})
	if err != nil {
		t.Fatal(err)
	}
	// Both pings share the same calendar second; the later one is ingested first.
	sec := time.Unix(1_700_000_000, 0).UTC()
	later := sec.Add(900_000_001)  // .900000001s
	earlier := sec.Add(900_000_000) // .900000000s (earlier by 1ns, same second)

	makeEcho := func(seq int, ts time.Time) *model.EchoWindow {
		return &model.EchoWindow{
			BatchID: bid, PingSeq: seq, Timestamp: ts, SoundVelocity: 1500, SlantRange: 20,
			Channels: []model.EchoChannel{{FrequencyHz: 200000, Depths: []float64{0, 1}, Amplitudes: []float64{-10, -11}}},
		}
	}
	if _, err := st.InsertEcho(makeEcho(1, later)); err != nil {
		t.Fatalf("ingest later: %v", err)
	}
	// The earlier ping, received second, must be rejected as a real regression.
	if _, err := st.InsertEcho(makeEcho(2, earlier)); !errors.Is(err, model.ErrTimeRegress) {
		t.Fatalf("expected same-second time regression to be rejected, got %v", err)
	}
	// An equal timestamp (same instant, distinct ping sequence) is permitted.
	if _, err := st.InsertEcho(makeEcho(3, later)); err != nil {
		t.Fatalf("expected equal timestamp to be permitted, got %v", err)
	}
}
