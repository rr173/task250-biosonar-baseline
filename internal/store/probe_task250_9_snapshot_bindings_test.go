package store_test

import (
	"path/filepath"
	"testing"

	"task250-biosonar/internal/model"
	"task250-biosonar/internal/service"
	"task250-biosonar/internal/store"
)

func TestBug09_SnapshotListRetainsSegmentBindings(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := service.New(st)
	bid, err := st.CreateBatch(&model.SurveyBatch{Name: "snapshot-bindings"})
	if err != nil {
		t.Fatal(err)
	}
	ids, err := st.ReplaceSegments(bid, []model.SubstrateSegment{{BatchID: bid, SubstrateID: 1, StartSeq: 0, EndSeq: 2, Status: model.SegContinuous, Confidence: 0.9}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 {
		t.Fatalf("created %d segments", len(ids))
	}
	if _, err := svc.PublishBatchSnapshot(bid, "v1"); err != nil {
		t.Fatal(err)
	}
	snaps, err := st.ListSnapshots(bid)
	if err != nil {
		t.Fatal(err)
	}
	if len(snaps) != 1 || len(snaps[0].SegmentIDs) != 1 || snaps[0].SegmentIDs[0] != ids[0] {
		t.Fatalf("snapshot bindings lost: %+v", snaps)
	}
}
