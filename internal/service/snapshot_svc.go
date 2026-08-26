package service

import (
	"task250-biosonar/internal/model"
)

// PublishBatchSnapshot creates a draft snapshot from a batch's current
// segments and immediately publishes it, superseding any prior published
// snapshot of the same batch.
func (svc *Service) PublishBatchSnapshot(batchID int64, note string) (*model.InterpretationSnapshot, error) {
	if _, err := svc.store.GetBatch(batchID); err != nil {
		return nil, err
	}
	segs, err := svc.store.ListSegments(batchID)
	if err != nil {
		return nil, err
	}
	if len(segs) == 0 {
		return nil, model.ErrNotClassified
	}
	ids := make([]int64, 0, len(segs))
	for _, s := range segs {
		ids = append(ids, s.ID)
	}
	draftID, err := svc.store.CreateSnapshot(batchID, note, ids)
	if err != nil {
		return nil, err
	}
	return svc.store.PublishSnapshot(draftID)
}
