package store_test

import (
	"path/filepath"
	"testing"
	"time"

	"task250-biosonar/internal/model"
	"task250-biosonar/internal/store"
)

func TestBug03_NonMonotonicDepthProfileRejected(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	bid, err := st.CreateBatch(&model.SurveyBatch{Name: "depth-order"})
	if err != nil {
		t.Fatal(err)
	}
	e := &model.EchoWindow{BatchID: bid, PingSeq: 1, Timestamp: time.Unix(1, 0), SoundVelocity: 1500, SlantRange: 20, Channels: []model.EchoChannel{{FrequencyHz: 200000, Depths: []float64{0, 2, 1}, Amplitudes: []float64{-10, -11, -12}}}}
	if _, err := st.InsertEcho(e); err == nil {
		t.Fatal("non-monotonic depth profile was accepted")
	}
}
