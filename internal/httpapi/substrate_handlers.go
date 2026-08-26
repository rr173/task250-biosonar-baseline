package httpapi

import (
	"errors"
	"net/http"

	"task250-biosonar/internal/model"
)

type createSubstrateReq struct {
	Code     string    `json:"code"`
	Name     string    `json:"name"`
	Centroid []float64 `json:"centroid"`
	CovDiag  []float64 `json:"cov_diag"`
}

func (s *Server) createSubstrate(w http.ResponseWriter, r *http.Request) {
	var req createSubstrateReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Code == "" || len(req.Centroid) != model.FeatureDim || len(req.CovDiag) != model.FeatureDim {
		writeError(w, http.StatusBadRequest, errors.New("code and a 4-dimensional centroid/cov_diag are required"))
		return
	}
	sub := &model.SubstrateType{Code: req.Code, Name: req.Name, Centroid: req.Centroid, CovDiag: req.CovDiag}
	id, err := s.svc.Store().CreateSubstrate(sub)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id, "code": sub.Code})
}

func (s *Server) listSubstrates(w http.ResponseWriter, r *http.Request) {
	subs, err := s.svc.Store().ListSubstrates()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, subs)
}
