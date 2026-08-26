package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"task250-biosonar/internal/model"
	"task250-biosonar/internal/service"
	"task250-biosonar/internal/store"
)

func newTestServer(t *testing.T) (*Server, int64) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "biosonar.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	svc := service.New(st)
	bid, err := st.CreateBatch(&model.SurveyBatch{Name: "ts-line"})
	if err != nil {
		t.Fatal(err)
	}
	return NewServer(svc), bid
}

func postEcho(t *testing.T, h http.Handler, body map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/echoes", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func validEchoBody(bid int64, ts string) map[string]interface{} {
	return map[string]interface{}{
		"batch_id":       bid,
		"ping_seq":       1,
		"pos_x":          0,
		"pos_y":          0,
		"timestamp":      ts,
		"sound_velocity": 1500,
		"slant_range":    50,
		"channels": []map[string]interface{}{
			{"frequency_hz": 200000, "incidence_deg": 30, "depths_m": []float64{0, 1}, "amplitudes_db": []float64{-10, -11}},
		},
	}
}

func TestIngestEchoRejectsMalformedTimestamp(t *testing.T) {
	srv, bid := newTestServer(t)
	h := srv.Handler()

	for _, bad := range []string{"", "not-a-timestamp", "2026-08-26 12:00:00", "2026/08/26T12:00:00Z"} {
		rec := postEcho(t, h, validEchoBody(bid, bad))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("timestamp %q: expected 400, got %d (body=%s)", bad, rec.Code, rec.Body.String())
		}
		var resp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if !errors.Is(model.ErrInvalidTimestamp, model.ErrInvalidTimestamp) || resp["error"] == "" {
			t.Fatalf("expected non-empty error, got %v", resp)
		}
	}

	// No echo must have been persisted for any rejected ingest.
	got, err := srv.svc.Store().ListEchoes(bid)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected zero echoes after rejections, got %d", len(got))
	}
}

func TestIngestEchoAcceptsValidTimestamp(t *testing.T) {
	srv, bid := newTestServer(t)
	h := srv.Handler()

	want := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	rec := postEcho(t, h, validEchoBody(bid, want.Format(time.RFC3339Nano)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for valid timestamp, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	e, err := srv.svc.Store().GetEcho(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !e.Timestamp.Equal(want) {
		t.Fatalf("stored timestamp = %s, want %s (server must not rewrite the client time)", e.Timestamp, want)
	}
}
