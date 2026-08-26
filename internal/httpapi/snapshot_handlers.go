package httpapi

import (
	"errors"
	"net/http"

	"task250-biosonar/internal/model"
)

type publishSnapshotReq struct {
	BatchID int64  `json:"batch_id"`
	Note    string `json:"note"`
}

func (s *Server) publishSnapshot(w http.ResponseWriter, r *http.Request) {
	var req publishSnapshotReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.BatchID == 0 {
		writeError(w, http.StatusBadRequest, errors.New("batch_id is required"))
		return
	}
	snap, err := s.svc.PublishBatchSnapshot(req.BatchID, req.Note)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, snap)
}

func (s *Server) listSnapshots(w http.ResponseWriter, r *http.Request) {
	bid, err := queryBatchID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	snaps, err := s.svc.Store().ListSnapshots(bid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, snaps)
}

func (s *Server) getSnapshot(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	snap, err := s.svc.Store().GetSnapshot(id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, model.ErrNotFound) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, snap)
}
