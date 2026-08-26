package store

import (
	"database/sql"

	"task250-biosonar/internal/model"
	"task250-biosonar/internal/segment"
)

// ReplaceSegments atomically deletes existing segments/boundaries for a batch
// and inserts the freshly merged set, returning the new segment IDs in order.
// Boundaries reference a segment by its StartSeq, which is resolved to the
// generated segment ID within the same transaction.
func (s *Store) ReplaceSegments(batchID int64, segs []model.SubstrateSegment, bounds []segment.Boundary) ([]int64, error) {
	if err := s.assertBatchOpen(batchID); err != nil {
		return nil, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM segment_boundaries WHERE segment_id IN (SELECT id FROM substrate_segments WHERE batch_id=?)`, batchID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`DELETE FROM substrate_segments WHERE batch_id=?`, batchID); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(segs))
	idByStart := map[int]int64{}
	for _, seg := range segs {
		res, err := tx.Exec(
			`INSERT INTO substrate_segments(batch_id, substrate_id, start_seq, end_seq, status, confidence)
			 VALUES(?,?,?,?,?,?)`,
			batchID, seg.SubstrateID, seg.StartSeq, seg.EndSeq, string(seg.Status), seg.Confidence)
		if err != nil {
			return nil, err
		}
		id, _ := res.LastInsertId()
		ids = append(ids, id)
		idByStart[seg.StartSeq] = id
	}
	for _, b := range bounds {
		segID, ok := idByStart[b.Seq]
		if !ok {
			continue
		}
		if _, err := tx.Exec(
			`INSERT INTO segment_boundaries(segment_id, seq, reason, from_substrate, to_substrate)
			 VALUES(?,?,?,?,?)`,
			segID, b.Seq, b.Reason, b.FromSubstrate, b.ToSubstrate); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}

// GetSegment loads a segment by ID.
func (s *Store) GetSegment(id int64) (*model.SubstrateSegment, error) {
	seg := &model.SubstrateSegment{ID: id}
	err := s.db.QueryRow(
		`SELECT batch_id, substrate_id, start_seq, end_seq, status, confidence FROM substrate_segments WHERE id=?`, id).
		Scan(&seg.BatchID, &seg.SubstrateID, &seg.StartSeq, &seg.EndSeq, &seg.Status, &seg.Confidence)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return seg, nil
}

// ListSegments returns segments for a batch ordered by start sequence.
func (s *Store) ListSegments(batchID int64) ([]model.SubstrateSegment, error) {
	rows, err := s.db.Query(
		`SELECT id, substrate_id, start_seq, end_seq, status, confidence FROM substrate_segments WHERE batch_id=? ORDER BY start_seq`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.SubstrateSegment
	for rows.Next() {
		var seg model.SubstrateSegment
		if err := rows.Scan(&seg.ID, &seg.SubstrateID, &seg.StartSeq, &seg.EndSeq, &seg.Status, &seg.Confidence); err != nil {
			return nil, err
		}
		seg.BatchID = batchID
		out = append(out, seg)
	}
	return out, rows.Err()
}

// UpdateSegmentStatus validates and applies a segment status transition.
func (s *Store) UpdateSegmentStatus(id int64, to model.SegmentStatus) error {
	seg, err := s.GetSegment(id)
	if err != nil {
		return err
	}
	if err := model.TransitionSegment(seg.Status, to); err != nil {
		return err
	}
	if err := s.assertBatchOpen(seg.BatchID); err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE substrate_segments SET status=? WHERE id=?`, string(to), id)
	return err
}
