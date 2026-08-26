package service

import (
	"task250-biosonar/internal/geometry"
	"task250-biosonar/internal/model"
)

// IngestEcho stores a ping and, when vessel attitude is too unstable to trust
// the geometric correction, immediately flags it as an attitude anomaly.
func (svc *Service) IngestEcho(e *model.EchoWindow) (int64, error) {
	id, err := svc.store.InsertEcho(e)
	if err != nil {
		return 0, err
	}
	if geometry.IsAttitudeAnomaly(e.Attitude, svc.geoPitchLimit, svc.geoRollLimit) {
		if err := svc.store.UpdateEchoStatus(id, model.EchoAttitudeAnom); err != nil {
			return id, err
		}
	}
	return id, nil
}

// CorrectEcho reconciles a ping's slant range against attitude, sound velocity
// and heave, persisting the corrected seafloor sample point.
func (svc *Service) CorrectEcho(echoID int64) error {
	e, err := svc.store.GetEcho(echoID)
	if err != nil {
		return err
	}
	if e.Status == model.EchoExcluded || e.Status == model.EchoAttitudeAnom {
		return model.ErrEchoExcluded
	}
	if err := geometry.Correct(e); err != nil {
		return err
	}
	return svc.store.SaveCorrectedGeometry(e)
}
