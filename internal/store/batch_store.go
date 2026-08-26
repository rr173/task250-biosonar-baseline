package store

import (
	"database/sql"
	"fmt"
	"time"

	"task250-biosonar/internal/model"
)

// CreateBatch inserts a new survey batch, defaulting its status to "receiving".
func (s *Store) CreateBatch(b *model.SurveyBatch) (int64, error) {
	if b.Status == "" {
		b.Status = model.BatchReceiving
	}
	b.CreatedAt = time.Now().UTC()
	res, err := s.db.Exec(
		`INSERT INTO survey_batches(name, vessel, status, created_at) VALUES(?,?,?,?)`,
		b.Name, b.Vessel, string(b.Status), b.CreatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	b.ID = id
	return id, nil
}

// GetBatch loads a batch by ID.
func (s *Store) GetBatch(id int64) (*model.SurveyBatch, error) {
	b := &model.SurveyBatch{ID: id}
	var ca, sa sql.NullString
	err := s.db.QueryRow(
		`SELECT name, vessel, status, created_at, sealed_at FROM survey_batches WHERE id=?`, id).
		Scan(&b.Name, &b.Vessel, &b.Status, &ca, &sa)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	b.CreatedAt, _ = time.Parse(time.RFC3339Nano, ca.String)
	if sa.Valid {
		t, _ := time.Parse(time.RFC3339Nano, sa.String)
		b.SealedAt = &t
	}
	return b, nil
}

// ListBatches returns all batches ordered by ID.
func (s *Store) ListBatches() ([]model.SurveyBatch, error) {
	rows, err := s.db.Query(
		`SELECT id, name, vessel, status, created_at, sealed_at FROM survey_batches ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.SurveyBatch
	for rows.Next() {
		var b model.SurveyBatch
		var ca, sa sql.NullString
		if err := rows.Scan(&b.ID, &b.Name, &b.Vessel, &b.Status, &ca, &sa); err != nil {
			return nil, err
		}
		b.CreatedAt, _ = time.Parse(time.RFC3339Nano, ca.String)
		if sa.Valid {
			t, _ := time.Parse(time.RFC3339Nano, sa.String)
			b.SealedAt = &t
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// TransitionBatch validates and applies a batch status change, stamping
// sealed_at when moving to "sealed".
func (s *Store) TransitionBatch(id int64, to model.BatchStatus) (*model.SurveyBatch, error) {
	b, err := s.GetBatch(id)
	if err != nil {
		return nil, err
	}
	if err := model.TransitionBatch(b.Status, to); err != nil {
		return nil, err
	}
	if to == model.BatchSealed {
		now := time.Now().UTC()
		if _, err := s.db.Exec(
			`UPDATE survey_batches SET status=?, sealed_at=? WHERE id=?`,
			string(to), now.Format(time.RFC3339Nano), id); err != nil {
			return nil, err
		}
		b.SealedAt = &now
	} else {
		if _, err := s.db.Exec(
			`UPDATE survey_batches SET status=? WHERE id=?`, string(to), id); err != nil {
			return nil, err
		}
	}
	b.Status = to
	return b, nil
}

// assertBatchOpen returns an error if the batch is sealed.
func (s *Store) assertBatchOpen(id int64) error {
	b, err := s.GetBatch(id)
	if err != nil {
		return err
	}
	if b.Status == model.BatchSealed {
		return fmt.Errorf("%w: batch %d", model.ErrBatchSealed, id)
	}
	return nil
}

// assertEchoBatchOpen resolves the owning batch of an echo window and rejects
// writes when that batch is sealed. A sealed batch freezes every echo,
// classification, feature and segment belonging to it.
func (s *Store) assertEchoBatchOpen(echoID int64) error {
	var batchID int64
	err := s.db.QueryRow(`SELECT batch_id FROM echo_windows WHERE id=?`, echoID).Scan(&batchID)
	if err == sql.ErrNoRows {
		return model.ErrNotFound
	}
	if err != nil {
		return err
	}
	return s.assertBatchOpen(batchID)
}

// LastEchoTimestamp returns the most recent ping timestamp for a batch, used to
// reject time-regressing ingests. It returns the zero time when none exist.
func (s *Store) LastEchoTimestamp(batchID int64) (time.Time, error) {
	var ts sql.NullString
	err := s.db.QueryRow(
		`SELECT MAX(ts) FROM echo_windows WHERE batch_id=?`, batchID).Scan(&ts)
	if err != nil {
		return time.Time{}, err
	}
	if !ts.Valid {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339Nano, ts.String)
}
