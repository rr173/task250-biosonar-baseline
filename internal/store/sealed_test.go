package store

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"task250-biosonar/internal/classify"
	"task250-biosonar/internal/echo"
	"task250-biosonar/internal/geometry"
	"task250-biosonar/internal/model"
	"task250-biosonar/internal/segment"
)

// seededStore opens a fresh store, advances one batch all the way through to
// "sealed" and returns it along with a single corrected+classified ping whose
// interpretation data is already published.
func seededStore(t *testing.T) (*Store, int64, int64) {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "biosonar.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	bid, err := st.CreateBatch(&model.SurveyBatch{Name: "sealed-line"})
	if err != nil {
		t.Fatal(err)
	}
	base := time.Unix(1000, 0).UTC()
	e := &model.EchoWindow{
		BatchID:       bid,
		PingSeq:       1,
		PosX:          100, PosY: 200,
		Timestamp:     base,
		Attitude:      model.Attitude{Heading: 90},
		SoundVelocity: 1500, SlantRange: 20,
		Channels: []model.EchoChannel{{
			FrequencyHz:  200000,
			IncidenceDeg: 30,
			Depths:       []float64{0, 1, 2},
			Amplitudes:   []float64{-10, -11, -12},
		}},
	}
	eid, err := st.InsertEcho(e)
	if err != nil {
		t.Fatal(err)
	}
	if err := geometry.Correct(e); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveCorrectedGeometry(e); err != nil {
		t.Fatal(err)
	}
	subs, err := st.ListSubstrates()
	if err != nil {
		t.Fatal(err)
	}
	fv, err := echo.Extract(e)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SaveFeature(eid, fv); err != nil {
		t.Fatal(err)
	}
	cls, err := classify.Classify(eid, fv, subs)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SaveClassification(cls); err != nil {
		t.Fatal(err)
	}
	// Merge into a segment and publish a snapshot so interpretation data exists.
	segs := segment.Merge(bid, []segment.ClassifiedPing{
		{Seq: 1, SubstrateID: cls.PredictedID, Probability: cls.BestProbability()},
	}, segment.DefaultMergeConfig())
	if _, err := st.ReplaceSegments(bid, segs, segment.Boundaries(segs)); err != nil {
		t.Fatal(err)
	}
	// Publish the interpretation snapshot, then drive the batch through to the
	// terminal "sealed" status.
	if _, err := st.CreateSnapshot(bid, "interpretation", []int64{segs[0].ID}); err != nil {
		t.Fatal(err)
	}
	snap, _ := st.ListSnapshots(bid)
	snapID := snap[0].ID
	if _, err := st.PublishSnapshot(snapID); err != nil {
		t.Fatal(err)
	}
	for _, to := range []model.BatchStatus{
		model.BatchPendingCorr, model.BatchPendingClass, model.BatchPublished, model.BatchSealed,
	} {
		if _, err := st.TransitionBatch(bid, to); err != nil {
			t.Fatalf("transition to %s: %v", to, err)
		}
	}
	return st, bid, eid
}

func TestSealedBatchRejectsEchoExclude(t *testing.T) {
	st, _, eid := seededStore(t)
	err := st.UpdateEchoStatus(eid, model.EchoExcluded)
	if !errors.Is(err, model.ErrBatchSealed) {
		t.Fatalf("expected ErrBatchSealed, got %v", err)
	}
	got, err := st.GetEcho(eid)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.EchoCorrected {
		t.Fatalf("sealed echo status rewritten: %s", got.Status)
	}
}

func TestSealedBatchRejectsEchoCorrection(t *testing.T) {
	st, _, eid := seededStore(t)
	e, err := st.GetEcho(eid)
	if err != nil {
		t.Fatal(err)
	}
	// Force a different correction value to detect a silent overwrite.
	e.CorrectedDepth = 999
	before := e.CorrectedDepth
	err = st.SaveCorrectedGeometry(e)
	if !errors.Is(err, model.ErrBatchSealed) {
		t.Fatalf("expected ErrBatchSealed, got %v", err)
	}
	got, err := st.GetEcho(eid)
	if err != nil {
		t.Fatal(err)
	}
	if got.CorrectedDepth == before {
		t.Fatalf("corrected geometry overwritten on sealed batch")
	}
}

func TestSealedBatchRejectsClassificationRewrite(t *testing.T) {
	st, _, eid := seededStore(t)
	before, err := st.GetClassification(eid)
	if err != nil {
		t.Fatal(err)
	}
	// A brand-new posterior pointing at a different substrate.
	cls := &classify.Classification{
		EchoID:      eid,
		PredictedID: before.PredictedID + 1,
		Results:     []classify.Result{{SubstrateID: before.PredictedID + 1, Probability: 1}},
	}
	if err := st.SaveClassification(cls); !errors.Is(err, model.ErrBatchSealed) {
		t.Fatalf("expected ErrBatchSealed, got %v", err)
	}
	got, err := st.GetClassification(eid)
	if err != nil {
		t.Fatal(err)
	}
	if got.PredictedID != before.PredictedID {
		t.Fatalf("classification rewritten on sealed batch: %d != %d", got.PredictedID, before.PredictedID)
	}
}

func TestSealedBatchRejectsFeatureRewrite(t *testing.T) {
	st, _, eid := seededStore(t)
	err := st.SaveFeature(eid, model.FeatureVector{1, 2, 3, 4})
	if !errors.Is(err, model.ErrBatchSealed) {
		t.Fatalf("expected ErrBatchSealed, got %v", err)
	}
}

func TestSealedBatchRejectsSegmentReplace(t *testing.T) {
	st, bid, _ := seededStore(t)
	segs := segment.Merge(bid, nil, segment.DefaultMergeConfig())
	_, err := st.ReplaceSegments(bid, segs, nil)
	if !errors.Is(err, model.ErrBatchSealed) {
		t.Fatalf("expected ErrBatchSealed, got %v", err)
	}
	got, err := st.ListSegments(bid)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatalf("segment data dropped on sealed batch")
	}
}

func TestSealedBatchRejectsSegmentStatusChange(t *testing.T) {
	st, bid, _ := seededStore(t)
	segs, err := st.ListSegments(bid)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateSegmentStatus(segs[0].ID, model.SegConfirmed); !errors.Is(err, model.ErrBatchSealed) {
		t.Fatalf("expected ErrBatchSealed, got %v", err)
	}
	got, err := st.GetSegment(segs[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != segs[0].Status {
		t.Fatalf("segment status rewritten on sealed batch: %s", got.Status)
	}
}
