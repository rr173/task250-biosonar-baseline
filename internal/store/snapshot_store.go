package store

import (
	"database/sql"
	"time"

	"task250-biosonar/internal/model"
	"task250-biosonar/internal/versioning"
)

// CreateSnapshot records a draft interpretation snapshot binding a set of
// segments.
func (s *Store) CreateSnapshot(batchID int64, note string, segmentIDs []int64) (int64, error) {
	if err := s.assertBatchOpen(batchID); err != nil {
		return 0, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(
		`INSERT INTO snapshots(batch_id, status, note, created_at) VALUES(?,?,?,?)`,
		batchID, string(model.SnapDraft), note, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	for _, sid := range segmentIDs {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO snapshot_segments(snapshot_id, segment_id) VALUES(?,?)`, id, sid); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// PublishSnapshot moves a draft to "published" and, under the default policy,
// supersedes any previously published snapshot of the same batch.
func (s *Store) PublishSnapshot(id int64) (*model.InterpretationSnapshot, error) {
	snap, err := s.GetSnapshot(id)
	if err != nil {
		return nil, err
	}
	if err := model.TransitionSnapshot(snap.Status, model.SnapPublished); err != nil {
		return nil, err
	}
	policy := versioning.DefaultPublishPolicy()
	if policy.SupersedePrior {
		existing, err := s.ListSnapshots(snap.BatchID)
		if err != nil {
			return nil, err
		}
		for _, priorID := range versioning.SupersedePriorPublished(snap.BatchID, existing) {
			if priorID == id {
				continue
			}
			if _, err := s.db.Exec(`UPDATE snapshots SET status=? WHERE id=?`, string(model.SnapSuperseded), priorID); err != nil {
				return nil, err
			}
		}
	}
	if _, err := s.db.Exec(`UPDATE snapshots SET status=? WHERE id=?`, string(model.SnapPublished), id); err != nil {
		return nil, err
	}
	return s.GetSnapshot(id)
}

// GetSnapshot loads a snapshot with its bound segment IDs.
func (s *Store) GetSnapshot(id int64) (*model.InterpretationSnapshot, error) {
	snap := &model.InterpretationSnapshot{ID: id}
	var ca string
	err := s.db.QueryRow(
		`SELECT batch_id, status, note, created_at FROM snapshots WHERE id=?`, id).
		Scan(&snap.BatchID, &snap.Status, &snap.Note, &ca)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	snap.CreatedAt, _ = time.Parse(time.RFC3339Nano, ca)
	if err := s.loadSnapshotSegments(snap); err != nil {
		return nil, err
	}
	return snap, nil
}

func (s *Store) loadSnapshotSegments(snap *model.InterpretationSnapshot) error {
	rows, err := s.db.Query(`SELECT segment_id FROM snapshot_segments WHERE snapshot_id=? ORDER BY segment_id`, snap.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var sid int64
		if err := rows.Scan(&sid); err != nil {
			return err
		}
		snap.SegmentIDs = append(snap.SegmentIDs, sid)
	}
	return rows.Err()
}

// ListSnapshots returns all snapshots for a batch ordered by ID.
func (s *Store) ListSnapshots(batchID int64) ([]model.InterpretationSnapshot, error) {
	rows, err := s.db.Query(
		`SELECT id, status, note, created_at FROM snapshots WHERE batch_id=? ORDER BY id`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.InterpretationSnapshot
	for rows.Next() {
		var snap model.InterpretationSnapshot
		var ca string
		if err := rows.Scan(&snap.ID, &snap.Status, &snap.Note, &ca); err != nil {
			return nil, err
		}
		snap.BatchID = batchID
		snap.CreatedAt, _ = time.Parse(time.RFC3339Nano, ca)
		out = append(out, snap)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListAllSnapshots returns every snapshot in the database (used for stats).
func (s *Store) ListAllSnapshots() ([]model.InterpretationSnapshot, error) {
	rows, err := s.db.Query(
		`SELECT id, batch_id, status, note, created_at FROM snapshots ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.InterpretationSnapshot
	for rows.Next() {
		var snap model.InterpretationSnapshot
		var ca string
		if err := rows.Scan(&snap.ID, &snap.BatchID, &snap.Status, &snap.Note, &ca); err != nil {
			return nil, err
		}
		snap.CreatedAt, _ = time.Parse(time.RFC3339Nano, ca)
		out = append(out, snap)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	for i := range out {
		if err := s.loadSnapshotSegments(&out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}
