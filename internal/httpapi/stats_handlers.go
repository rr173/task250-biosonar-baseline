package httpapi

import (
	"database/sql"
	"net/http"

	"task250-biosonar/internal/model"
)

type statsResp struct {
	Batches       int `json:"batches"`
	Echoes        int `json:"echoes"`
	Substrates    int `json:"substrates"`
	Segments      int `json:"segments"`
	Snapshots     int `json:"snapshots"`
	SealedBatches int `json:"sealed_batches"`
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	resp := statsResp{}
	bs, _ := s.svc.Store().ListBatches()
	resp.Batches = len(bs)
	for _, b := range bs {
		if b.Status == model.BatchSealed {
			resp.SealedBatches++
		}
	}
	db := s.svc.Store().DB()
	resp.Echoes = countRows(db, "echo_windows")
	resp.Substrates = countRows(db, "substrate_types")
	resp.Segments = countRows(db, "substrate_segments")
	resp.Snapshots = countRows(db, "snapshots")
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func countRows(db *sql.DB, table string) int {
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		return 0
	}
	return n
}
