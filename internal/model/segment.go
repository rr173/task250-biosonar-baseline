package model

// SegmentStatus drives the lifecycle of a merged seafloor segment.
//
// candidate -> continuous | uncertain -> confirmed | rejected
//
// "confirmed" and "rejected" are terminal.
type SegmentStatus string

const (
	SegCandidate   SegmentStatus = "candidate"
	SegContinuous  SegmentStatus = "continuous"
	SegUncertain   SegmentStatus = "uncertain"
	SegConfirmed   SegmentStatus = "confirmed"
	SegRejected    SegmentStatus = "rejected"
)

var segTransitions = map[SegmentStatus][]SegmentStatus{
	SegCandidate:  {SegContinuous, SegUncertain, SegRejected},
	SegContinuous: {SegConfirmed, SegUncertain, SegRejected},
	SegUncertain:  {SegConfirmed, SegRejected},
	SegConfirmed:  {},
	SegRejected:   {},
}

// CanTransitionSegment reports whether a segment may move from->to.
func CanTransitionSegment(from, to SegmentStatus) bool {
	for _, n := range segTransitions[from] {
		if n == to {
			return true
		}
	}
	return false
}

// SubstrateSegment is a contiguous run of pings assigned to one substrate.
type SubstrateSegment struct {
	ID          int64         `json:"id"`
	BatchID     int64         `json:"batch_id"`
	SubstrateID int64         `json:"substrate_id"`
	StartSeq    int           `json:"start_seq"`
	EndSeq      int           `json:"end_seq"`
	Status      SegmentStatus `json:"status"`
	Confidence  float64       `json:"confidence"`
}
