package store_test

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"task250-biosonar/internal/model"
	"task250-biosonar/internal/service"
	"task250-biosonar/internal/store"
)

func TestBug07_SealedBatchRejectsAllMutations(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := service.New(st)
	bid, err := st.CreateBatch(&model.SurveyBatch{Name: "sealed"})
	if err != nil {
		t.Fatal(err)
	}
	e := &model.EchoWindow{BatchID: bid, PingSeq: 1, Timestamp: time.Unix(1, 0), SoundVelocity: 1500, SlantRange: 20, Channels: []model.EchoChannel{{FrequencyHz: 200000, Depths: []float64{0, 1}, Amplitudes: []float64{-10, -11}}}}
	id, err := st.InsertEcho(e)
	if err != nil {
		t.Fatal(err)
	}
	for _, status := range []model.BatchStatus{model.BatchPendingCorr, model.BatchPendingClass, model.BatchPublished, model.BatchSealed} {
		if _, err := st.TransitionBatch(bid, status); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.UpdateEchoStatus(id, model.EchoExcluded); !errors.Is(err, model.ErrBatchSealed) {
		t.Fatalf("exclude after seal returned %v", err)
	}
	if _, err := svc.ClassifyEcho(id); !errors.Is(err, model.ErrBatchSealed) {
		t.Fatalf("classify after seal returned %v", err)
	}
	got, err := st.GetEcho(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.EchoRaw {
		t.Fatalf("sealed echo changed status to %s", got.Status)
	}
}
