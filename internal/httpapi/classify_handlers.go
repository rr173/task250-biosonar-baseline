package httpapi

import (
	"errors"
	"net/http"

	"task250-biosonar/internal/model"
)

func (s *Server) classifyEcho(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	cls, err := s.svc.ClassifyEcho(id)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, model.ErrNotFound) {
			status = http.StatusNotFound
		}
		if errors.Is(err, model.ErrEchoExcluded) {
			status = http.StatusConflict
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, cls)
}

func (s *Server) getClassification(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	cls, err := s.svc.Store().GetClassification(id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrNotClassified) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, cls)
}
