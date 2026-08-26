package segment

import (
	"sort"

	"task250-biosonar/internal/model"
)

// ClassifiedPing is the minimal per-ping classification used for merging.
type ClassifiedPing struct {
	Seq         int
	SubstrateID int64
	Probability float64
	Excluded    bool
}

// MergeConfig tunes segment formation.
type MergeConfig struct {
	// MinSegmentLen is the smallest run (in pings) that qualifies as
	// "continuous"; shorter runs stay "candidate".
	MinSegmentLen int
	// UncertaintyThreshold: a ping with probability below this breaks a run
	// and the resulting segment is flagged "uncertain".
	UncertaintyThreshold float64
}

// DefaultMergeConfig returns the standard merging rules.
func DefaultMergeConfig() MergeConfig {
	return MergeConfig{MinSegmentLen: 3, UncertaintyThreshold: 0.6}
}

type builder struct {
	substrateID int64
	startSeq    int
	endSeq      int
	sumProb     float64
	count       int
}

// Merge turns an ordered set of classified pings into contiguous seafloor
// segments. Excluded pings break continuity; substrate changes or low
// probability also end the current run.
func Merge(batchID int64, pings []ClassifiedPing, cfg MergeConfig) []model.SubstrateSegment {
	ordered := make([]ClassifiedPing, len(pings))
	copy(ordered, pings)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Seq < ordered[j].Seq })

	var segs []model.SubstrateSegment
	var cur *builder
	flush := func() {
		if cur == nil {
			return
		}
		conf := 0.0
		if cur.count > 0 {
			conf = cur.sumProb / float64(cur.count)
		}
		status := model.SegContinuous
		switch {
		case conf < cfg.UncertaintyThreshold:
			status = model.SegUncertain
		case cur.endSeq-cur.startSeq+1 < cfg.MinSegmentLen:
			status = model.SegCandidate
		}
		segs = append(segs, model.SubstrateSegment{
			BatchID:     batchID,
			SubstrateID: cur.substrateID,
			StartSeq:    cur.startSeq,
			EndSeq:      cur.endSeq,
			Status:      status,
			Confidence:  conf,
		})
		cur = nil
	}

	for _, p := range ordered {
		if p.Excluded {
			flush()
			continue
		}
		if cur == nil {
			cur = &builder{substrateID: p.SubstrateID, startSeq: p.Seq, endSeq: p.Seq, sumProb: p.Probability, count: 1}
			continue
		}
		if cur.substrateID != p.SubstrateID || p.Probability < cfg.UncertaintyThreshold {
			flush()
			cur = &builder{substrateID: p.SubstrateID, startSeq: p.Seq, endSeq: p.Seq, sumProb: p.Probability, count: 1}
			continue
		}
		cur.endSeq = p.Seq
		cur.sumProb += p.Probability
		cur.count++
	}
	flush()
	return segs
}
