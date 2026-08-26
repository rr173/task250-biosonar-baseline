package store_test

import (
	"errors"
	"path/filepath"
	"testing"

	"task250-biosonar/internal/model"
	"task250-biosonar/internal/service"
	"task250-biosonar/internal/store"
)

func TestBug08_EmptyBatchCannotPublishSnapshot(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := service.New(st)
	bid, err := st.CreateBatch(&model.SurveyBatch{Name: "empty"})
	if err != nil { t.Fatal(err) }
	if _, err := svc.PublishBatchSnapshot(bid, "empty"); !errors.Is(err, model.ErrNotClassified) {
		t.Fatalf("expected empty snapshot rejection, got %v", err)
	}
	snaps, err := st.ListSnapshots(bid)
	if err != nil { t.Fatal(err) }
	if len(snaps) != 0 { t.Fatalf("failed publish created %d snapshots", len(snaps)) }
}
