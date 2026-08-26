package model

import "time"

// BatchStatus drives the lifecycle of a survey line.
//
// receiving -> pending_correction -> pending_classification -> published -> sealed
//
// "sealed" is terminal and makes the whole batch immutable (echoes,
// classifications, segments and snapshots are frozen).
type BatchStatus string

const (
	BatchReceiving    BatchStatus = "receiving"
	BatchPendingCorr  BatchStatus = "pending_correction"
	BatchPendingClass BatchStatus = "pending_classification"
	BatchPublished    BatchStatus = "published"
	BatchSealed       BatchStatus = "sealed"
)

var batchTransitions = map[BatchStatus][]BatchStatus{
	BatchReceiving:    {BatchPendingCorr},
	BatchPendingCorr:  {BatchPendingClass},
	BatchPendingClass: {BatchPublished},
	BatchPublished:    {BatchSealed},
	BatchSealed:       {},
}

// CanTransitionBatch reports whether a batch may move from->to.
func CanTransitionBatch(from, to BatchStatus) bool {
	for _, n := range batchTransitions[from] {
		if n == to {
			return true
		}
	}
	return false
}

// SurveyBatch groups a contiguous run of pings collected along one survey line.
type SurveyBatch struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Vessel    string     `json:"vessel"`
	Status    BatchStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	SealedAt  *time.Time `json:"sealed_at,omitempty"`
}
