package store_test

import (
	"path/filepath"
	"testing"
	"time"

	"task250-biosonar/internal/classify"
	"task250-biosonar/internal/model"
	"task250-biosonar/internal/service"
	"task250-biosonar/internal/store"
)

func TestBug06_PingGapBreaksContinuousSegment(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := service.New(st)
	bid, err := st.CreateBatch(&model.SurveyBatch{Name: "gap"})
	if err != nil {
		t.Fatal(err)
	}
	for _, seq := range []int{0, 2} {
		e := &model.EchoWindow{BatchID: bid, PingSeq: seq, Timestamp: time.Unix(int64(seq+1), 0), SoundVelocity: 1500, SlantRange: 20, Channels: []model.EchoChannel{{FrequencyHz: 200000, Depths: []float64{0, 1}, Amplitudes: []float64{-10, -11}}}}
		id, err := st.InsertEcho(e)
		if err != nil {
			t.Fatal(err)
		}
		if err := st.SaveClassification(&classify.Classification{EchoID: id, PredictedID: 1, PredictedCode: "SAND", Uncertainty: 0.1, Results: []classify.Result{{SubstrateID: 1, Probability: 0.9}}}); err != nil {
			t.Fatal(err)
		}
	}
	segs, err := svc.MergeBatch(bid)
	if err != nil {
		t.Fatal(err)
	}
	if len(segs) != 2 {
		t.Fatalf("ping gap was merged into %d segments", len(segs))
	}
}
