package segment

import (
	"testing"

	"task250-biosonar/internal/model"
)

func TestMergeTwoRuns(t *testing.T) {
	pings := []ClassifiedPing{
		{Seq: 0, SubstrateID: 1, Probability: 0.9},
		{Seq: 1, SubstrateID: 1, Probability: 0.92},
		{Seq: 2, SubstrateID: 1, Probability: 0.88},
		{Seq: 3, SubstrateID: 2, Probability: 0.91},
		{Seq: 4, SubstrateID: 2, Probability: 0.9},
		{Seq: 5, SubstrateID: 2, Probability: 0.93},
	}
	segs := Merge(1, pings, DefaultMergeConfig())
	if len(segs) != 2 {
		t.Fatalf("want 2 segments got %d", len(segs))
	}
	if segs[0].SubstrateID != 1 || segs[1].SubstrateID != 2 {
		t.Fatalf("wrong substrate assignment: %d %d", segs[0].SubstrateID, segs[1].SubstrateID)
	}
	if segs[0].Status != model.SegContinuous {
		t.Fatalf("seg0 want continuous got %s", segs[0].Status)
	}
	if segs[0].EndSeq != 2 || segs[1].StartSeq != 3 {
		t.Fatalf("segment span wrong: %d-%d vs %d-%d", segs[0].StartSeq, segs[0].EndSeq, segs[1].StartSeq, segs[1].EndSeq)
	}
}

func TestMergeExcludedBreaksRun(t *testing.T) {
	pings := []ClassifiedPing{
		{Seq: 0, SubstrateID: 1, Probability: 0.9},
		{Seq: 1, SubstrateID: 1, Probability: 0.9, Excluded: true},
		{Seq: 2, SubstrateID: 1, Probability: 0.9},
	}
	segs := Merge(1, pings, DefaultMergeConfig())
	if len(segs) != 2 {
		t.Fatalf("excluded ping should split run, want 2 got %d", len(segs))
	}
}

func TestBoundaries(t *testing.T) {
	segs := []model.SubstrateSegment{
		{ID: 1, SubstrateID: 1, StartSeq: 0, EndSeq: 2, Status: model.SegContinuous},
		{ID: 2, SubstrateID: 2, StartSeq: 3, EndSeq: 5, Status: model.SegContinuous},
	}
	b := Boundaries(segs)
	if len(b) != 1 {
		t.Fatalf("want 1 boundary got %d", len(b))
	}
	if b[0].Seq != 3 || b[0].FromSubstrate != 1 || b[0].ToSubstrate != 2 {
		t.Fatalf("boundary metadata wrong: %+v", b[0])
	}
}
