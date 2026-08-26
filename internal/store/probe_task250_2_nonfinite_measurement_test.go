package store_test

import (
	"math"
	"path/filepath"
	"testing"
	"time"

	"task250-biosonar/internal/model"
	"task250-biosonar/internal/store"
)

func TestBug02_NonFiniteMeasurementRejected(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	bid, err := st.CreateBatch(&model.SurveyBatch{Name: "nan"})
	if err != nil {
		t.Fatal(err)
	}
	e := &model.EchoWindow{BatchID: bid, PingSeq: 1, Timestamp: time.Unix(1, 0), SoundVelocity: 1500, SlantRange: 20, Attitude: model.Attitude{Pitch: math.NaN()}, Channels: []model.EchoChannel{{FrequencyHz: 200000, Depths: []float64{0, 1}, Amplitudes: []float64{-10, -11}}}}
	if _, err := st.InsertEcho(e); err == nil {
		t.Fatal("non-finite attitude was accepted")
	}
	got, err := st.ListEchoes(bid)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("rejected echo was persisted: %d", len(got))
	}
}
