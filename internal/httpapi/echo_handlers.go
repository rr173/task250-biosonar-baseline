package httpapi

import (
	"errors"
	"net/http"
	"time"

	"task250-biosonar/internal/model"
)

type ingestEchoReq struct {
	BatchID       int64               `json:"batch_id"`
	PingSeq       int                 `json:"ping_seq"`
	PosX          float64             `json:"pos_x"`
	PosY          float64             `json:"pos_y"`
	Timestamp     string              `json:"timestamp"`
	Attitude      model.Attitude      `json:"attitude"`
	SoundVelocity float64             `json:"sound_velocity"`
	SlantRange    float64             `json:"slant_range"`
	Channels      []model.EchoChannel `json:"channels"`
}

func (s *Server) ingestEcho(w http.ResponseWriter, r *http.Request) {
	var req ingestEchoReq
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.BatchID == 0 {
		writeError(w, http.StatusBadRequest, errors.New("batch_id is required"))
		return
	}
	ts := time.Now().UTC()
	if parsed, err := time.Parse(time.RFC3339Nano, req.Timestamp); err == nil {
		ts = parsed
	}
	e := &model.EchoWindow{
		BatchID:       req.BatchID,
		PingSeq:       req.PingSeq,
		PosX:          req.PosX,
		PosY:          req.PosY,
		Timestamp:     ts,
		Attitude:      req.Attitude,
		SoundVelocity: req.SoundVelocity,
		SlantRange:    req.SlantRange,
		Channels:      req.Channels,
		Status:        model.EchoRaw,
	}
	id, err := s.svc.IngestEcho(e)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id, "status": string(e.Status)})
}

func (s *Server) listEchoes(w http.ResponseWriter, r *http.Request) {
	bid, err := queryBatchID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	es, err := s.svc.Store().ListEchoes(bid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, es)
}

func (s *Server) getEcho(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	e, err := s.svc.Store().GetEcho(id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, model.ErrNotFound) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) excludeEcho(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.svc.Store().UpdateEchoStatus(id, model.EchoExcluded); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(model.EchoExcluded)})
}

func (s *Server) correctEcho(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.svc.CorrectEcho(id); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": string(model.EchoCorrected)})
}

func (s *Server) echoFeatures(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	fv, err := s.svc.Store().GetFeature(id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, model.ErrNotClassified) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, fv)
}
