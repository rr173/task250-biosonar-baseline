package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"task250-biosonar/internal/httpapi"
	"task250-biosonar/internal/service"
	"task250-biosonar/internal/store"
)

func TestBug01_InvalidTimestampIsRejected(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	h := httptest.NewServer(httpapi.NewServer(service.New(st)).Handler())
	defer h.Close()
	var batch struct {
		ID int64 `json:"id"`
	}
	resp, err := http.Post(h.URL+"/api/batches", "application/json", bytes.NewBufferString(`{"name":"bad-time"}`))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create batch status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&batch); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	var body map[string]any
	if err := json.Unmarshal([]byte(`{"batch_id":0,"ping_seq":1,"pos_x":1,"pos_y":2,"timestamp":"not-a-time","attitude":{"pitch_deg":0,"roll_deg":0,"heading_deg":0,"heave_m":0},"sound_velocity":1500,"slant_range":20,"channels":[{"frequency_hz":200000,"incidence_deg":20,"depths_m":[0,1],"amplitudes_db":[-10,-11]}]}`), &body); err != nil {
		t.Fatal(err)
	}
	body["batch_id"] = batch.ID
	encoded, _ := json.Marshal(body)
	resp, err = h.Client().Post(h.URL+"/api/echoes", "application/json", bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	resp.Body.Close()
	resp, err = h.Client().Get(h.URL + "/api/echoes?batch_id=" + strconv.FormatInt(batch.ID, 10))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var echoes []any
	if err := json.NewDecoder(resp.Body).Decode(&echoes); err != nil {
		t.Fatal(err)
	}
	if len(echoes) != 0 {
		t.Fatalf("invalid timestamp created %d echoes", len(echoes))
	}
}
