package model

import "time"

// SnapshotStatus drives the lifecycle of a published interpretation.
//
// draft -> published -> superseded
//
// "superseded" is terminal; publishing a new snapshot of the same batch
// supersedes the previous one.
type SnapshotStatus string

const (
	SnapDraft      SnapshotStatus = "draft"
	SnapPublished  SnapshotStatus = "published"
	SnapSuperseded SnapshotStatus = "superseded"
)

var snapTransitions = map[SnapshotStatus][]SnapshotStatus{
	SnapDraft:     {SnapPublished},
	SnapPublished:  {SnapSuperseded},
	SnapSuperseded: {},
}

// CanTransitionSnapshot reports whether a snapshot may move from->to.
func CanTransitionSnapshot(from, to SnapshotStatus) bool {
	for _, n := range snapTransitions[from] {
		if n == to {
			return true
		}
	}
	return false
}

// InterpretationSnapshot freezes a set of seafloor segments as a citable
// classification result for one batch.
type InterpretationSnapshot struct {
	ID         int64          `json:"id"`
	BatchID    int64          `json:"batch_id"`
	Status     SnapshotStatus `json:"status"`
	Note       string         `json:"note"`
	CreatedAt  time.Time      `json:"created_at"`
	SegmentIDs []int64        `json:"segment_ids"`
}
