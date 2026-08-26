package segment

import "task250-biosonar/internal/model"

// Boundary marks a change in seafloor classification along the survey line.
type Boundary struct {
	Seq           int    `json:"seq"`
	Reason        string `json:"reason"`
	FromSubstrate int64  `json:"from_substrate"`
	ToSubstrate   int64  `json:"to_substrate"`
}

// Boundaries returns the transitions between consecutive segments, with a
// human-readable reason for each change. Excluded pings do not appear as
// segments, so boundaries only connect real substrate runs.
func Boundaries(segs []model.SubstrateSegment) []Boundary {
	var out []Boundary
	for i := 1; i < len(segs); i++ {
		prev, cur := segs[i-1], segs[i]
		reason := "substrate change"
		switch {
		case cur.Status == model.SegUncertain:
			reason = "entering uncertain segment"
		case prev.Status == model.SegUncertain:
			reason = "emerging from uncertain segment"
		}
		out = append(out, Boundary{
			Seq:           cur.StartSeq,
			Reason:        reason,
			FromSubstrate: prev.SubstrateID,
			ToSubstrate:   cur.SubstrateID,
		})
	}
	return out
}
