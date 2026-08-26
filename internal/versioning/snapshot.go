package versioning

import "task250-biosonar/internal/model"

// PublishPolicy controls how publishing a new snapshot interacts with prior
// snapshots of the same batch.
type PublishPolicy struct {
	// SupersedePrior moves any previously published snapshot of the batch to
	// "superseded" so only one published interpretation is live at a time.
	SupersedePrior bool
}

// DefaultPublishPolicy returns the standard policy: a new publish supersedes
// the previous one.
func DefaultPublishPolicy() PublishPolicy {
	return PublishPolicy{SupersedePrior: true}
}

// SupersedePriorPublished returns the IDs of snapshots that should be moved to
// "superseded" when a new snapshot is published for batchID.
func SupersedePriorPublished(batchID int64, existing []model.InterpretationSnapshot) []int64 {
	var ids []int64
	for _, s := range existing {
		if s.BatchID == batchID && s.Status == model.SnapPublished {
			ids = append(ids, s.ID)
		}
	}
	return ids
}

// Coverage counts, per substrate, how many non-rejected segments a snapshot
// would contain.
func Coverage(segs []model.SubstrateSegment) map[int64]int {
	m := map[int64]int{}
	for _, s := range segs {
		if s.Status == model.SegRejected {
			continue
		}
		m[s.SubstrateID]++
	}
	return m
}
