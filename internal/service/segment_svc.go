package service

import (
	"task250-biosonar/internal/model"
	"task250-biosonar/internal/segment"
)

// MergeBatch gathers the per-ping classifications of a batch, merges spatially
// contiguous runs into seafloor segments and persists them together with their
// boundaries. It replaces any prior merge for the same batch.
func (svc *Service) MergeBatch(batchID int64) ([]model.SubstrateSegment, error) {
	rows, err := svc.store.ClassifiedPings(batchID)
	if err != nil {
		return nil, err
	}
	echoes, err := svc.store.ListEchoes(batchID)
	if err != nil {
		return nil, err
	}
	excluded := map[int]bool{}
	for _, e := range echoes {
		if e.Status == model.EchoExcluded || e.Status == model.EchoAttitudeAnom {
			excluded[e.PingSeq] = true
		}
	}
	pings := make([]segment.ClassifiedPing, 0, len(rows))
	for _, r := range rows {
		pings = append(pings, segment.ClassifiedPing{
			Seq:         r.PingSeq,
			SubstrateID: r.SubstrateID,
			Probability: r.Probability,
			Excluded:    excluded[r.PingSeq],
		})
	}
	segs := segment.Merge(batchID, pings, svc.mergeCfg)
	bounds := segment.Boundaries(segs)
	if _, err := svc.store.ReplaceSegments(batchID, segs, bounds); err != nil {
		return nil, err
	}
	return svc.store.ListSegments(batchID)
}
