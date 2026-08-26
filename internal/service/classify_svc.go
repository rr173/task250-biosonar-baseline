package service

import (
	"task250-biosonar/internal/classify"
	"task250-biosonar/internal/echo"
	"task250-biosonar/internal/model"
)

// ClassifyEcho extracts features from a corrected ping, classifies it against
// the substrate catalogue and persists the posterior. A raw ping is corrected
// on the fly first.
func (svc *Service) ClassifyEcho(echoID int64) (*classify.Classification, error) {
	e, err := svc.store.GetEcho(echoID)
	if err != nil {
		return nil, err
	}
	if e.Status == model.EchoExcluded || e.Status == model.EchoAttitudeAnom {
		return nil, model.ErrEchoExcluded
	}
	if e.Status == model.EchoRaw {
		if err := svc.CorrectEcho(echoID); err != nil {
			return nil, err
		}
		e, _ = svc.store.GetEcho(echoID)
	}
	fv, err := echo.Extract(e)
	if err != nil {
		return nil, err
	}
	if err := svc.store.SaveFeature(echoID, fv); err != nil {
		return nil, err
	}
	subs, err := svc.store.ListSubstrates()
	if err != nil {
		return nil, err
	}
	cls, err := classify.Classify(echoID, fv, subs)
	if err != nil {
		return nil, err
	}
	if err := svc.store.SaveClassification(cls); err != nil {
		return nil, err
	}
	return cls, nil
}
