package store_test

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"task250-biosonar/internal/classify"
	"task250-biosonar/internal/model"
	"task250-biosonar/internal/service"
	"task250-biosonar/internal/store"
)

func TestBug10_RecomputePreservesPublishedSnapshotEvidence(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := service.New(st)
	bid, err := st.CreateBatch(&model.SurveyBatch{Name: "snapshot-freeze"})
	if err != nil {
		t.Fatal(err)
	}
	for seq := 0; seq < 3; seq++ {
		e := &model.EchoWindow{BatchID: bid, PingSeq: seq, Timestamp: time.Unix(int64(seq+1), 0), SoundVelocity: 1500, SlantRange: 20, Channels: []model.EchoChannel{{FrequencyHz: 200000, Depths: []float64{0, 1}, Amplitudes: []float64{-10, -11}}}}
		id, err := st.InsertEcho(e)
		if err != nil {
			t.Fatal(err)
		}
		if err := st.SaveClassification(&classify.Classification{EchoID: id, PredictedID: 1, PredictedCode: "SAND", Uncertainty: 0.1, Results: []classify.Result{{SubstrateID: 1, Probability: 0.9}}}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.MergeBatch(bid); err != nil {
		t.Fatal(err)
	}
	snap, err := svc.PublishBatchSnapshot(bid, "v1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.MergeBatch(bid); !errors.Is(err, model.ErrSnapshotFrozen) {
		t.Fatalf("recompute after publish returned %v", err)
	}
	reloaded, err := st.GetSnapshot(snap.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.SegmentIDs) == 0 {
		t.Fatal("published snapshot lost all segment evidence")
	}
	if _, err := st.GetSegment(reloaded.SegmentIDs[0]); err != nil {
		t.Fatalf("snapshot segment evidence disappeared: %v", err)
	}
}
