package httpapi

import (
	"errors"
	"net/http"

	"task250-biosonar/internal/model"
)

type mergeReq struct {
	BatchID int64 `json:"batch_id"`
}

func (s *Server) mergeBatch(w http.ResponseWriter, r *http.Request) {
	var req mergeReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.BatchID == 0 {
		writeError(w, http.StatusBadRequest, errors.New("batch_id is required"))
		return
	}
	segs, err := s.svc.MergeBatch(req.BatchID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, segs)
}

func (s *Server) listSegments(w http.ResponseWriter, r *http.Request) {
	bid, err := queryBatchID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	segs, err := s.svc.Store().ListSegments(bid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, segs)
}

func (s *Server) confirmSegment(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.svc.Store().UpdateSegmentStatus(id, model.SegConfirmed); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(model.SegConfirmed)})
}

func (s *Server) rejectSegment(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.svc.Store().UpdateSegmentStatus(id, model.SegRejected); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(model.SegRejected)})
}
