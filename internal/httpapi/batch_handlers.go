package httpapi

import (
	"errors"
	"net/http"

	"task250-biosonar/internal/model"
)

type createBatchReq struct {
	Name   string `json:"name"`
	Vessel string `json:"vessel"`
}

func (s *Server) createBatch(w http.ResponseWriter, r *http.Request) {
	var req createBatchReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, errors.New("name is required"))
		return
	}
	b := &model.SurveyBatch{Name: req.Name, Vessel: req.Vessel}
	id, err := s.svc.Store().CreateBatch(b)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id, "status": string(b.Status)})
}

func (s *Server) listBatches(w http.ResponseWriter, r *http.Request) {
	bs, err := s.svc.Store().ListBatches()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, bs)
}

func (s *Server) getBatch(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	b, err := s.svc.Store().GetBatch(id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, model.ErrNotFound) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) sealBatch(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	b, err := s.svc.Store().TransitionBatch(id, model.BatchSealed)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrInvalidTransition) {
			status = http.StatusConflict
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) listClassifications(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	rows, err := s.svc.Store().ClassifiedPings(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
