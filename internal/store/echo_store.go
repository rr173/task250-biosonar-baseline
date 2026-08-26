package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"task250-biosonar/internal/model"
)

// InsertEcho persists a single ping. It rejects invalid local state, duplicate
// ping sequences within a batch and ingests that regress the batch clock.
func (s *Store) InsertEcho(e *model.EchoWindow) (int64, error) {
	if err := e.Validate(); err != nil {
		return 0, err
	}
	if err := s.assertBatchOpen(e.BatchID); err != nil {
		return 0, err
	}
	last, err := s.LastEchoTimestamp(e.BatchID)
	if err != nil {
		return 0, err
	}
	if !last.IsZero() && e.Timestamp.Before(last) {
		return 0, fmt.Errorf("%w: ping %d predates %s", model.ErrTimeRegress, e.PingSeq, last.Format(time.RFC3339))
	}
	chJSON := jsonBytes(e.Channels)
	attJSON := jsonBytes(e.Attitude)
	if e.Status == "" {
		e.Status = model.EchoRaw
	}
	res, err := s.db.Exec(
		`INSERT INTO echo_windows(batch_id, ping_seq, pos_x, pos_y, ts, attitude, sound_velocity, slant_range, status, channels)
		 VALUES(?,?,?,?,?,?,?,?,?,?)`,
		e.BatchID, e.PingSeq, e.PosX, e.PosY, e.Timestamp.UTC().Format(time.RFC3339Nano),
		string(attJSON), e.SoundVelocity, e.SlantRange, string(e.Status), string(chJSON))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, fmt.Errorf("%w: ping %d", model.ErrDuplicatePing, e.PingSeq)
		}
		return 0, err
	}
	id, _ := res.LastInsertId()
	e.ID = id
	return id, nil
}

// GetEcho loads a ping by ID, reconstructing attitude and channels from JSON.
func (s *Store) GetEcho(id int64) (*model.EchoWindow, error) {
	e := &model.EchoWindow{ID: id}
	var ts, att, ch string
	var cx, cy, cd sql.NullFloat64
	var cat sql.NullString
	err := s.db.QueryRow(
		`SELECT batch_id, ping_seq, pos_x, pos_y, ts, attitude, sound_velocity, slant_range, status, channels,
		        corrected_x, corrected_y, corrected_depth, corrected_at
		 FROM echo_windows WHERE id=?`, id).
		Scan(&e.BatchID, &e.PingSeq, &e.PosX, &e.PosY, &ts, &att, &e.SoundVelocity, &e.SlantRange, &e.Status, &ch,
			&cx, &cy, &cd, &cat)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	e.Timestamp, _ = time.Parse(time.RFC3339Nano, ts)
	if err := jsonUnmarshal(att, &e.Attitude); err != nil {
		return nil, err
	}
	if err := jsonUnmarshal(ch, &e.Channels); err != nil {
		return nil, err
	}
	if cx.Valid {
		e.CorrectedX = cx.Float64
	}
	if cy.Valid {
		e.CorrectedY = cy.Float64
	}
	if cd.Valid {
		e.CorrectedDepth = cd.Float64
	}
	if cat.Valid {
		t, _ := time.Parse(time.RFC3339Nano, cat.String)
		e.CorrectedAt = &t
	}
	return e, nil
}

// ListEchoes returns pings for a batch ordered by sequence.
func (s *Store) ListEchoes(batchID int64) ([]model.EchoWindow, error) {
	rows, err := s.db.Query(
		`SELECT id, batch_id, ping_seq, pos_x, pos_y, ts, attitude, sound_velocity, slant_range, status, channels,
		        corrected_x, corrected_y, corrected_depth, corrected_at
		 FROM echo_windows WHERE batch_id=? ORDER BY ping_seq`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.EchoWindow
	for rows.Next() {
		var e model.EchoWindow
		var ts, att, ch string
		var cx, cy, cd sql.NullFloat64
		var cat sql.NullString
		if err := rows.Scan(&e.ID, &e.BatchID, &e.PingSeq, &e.PosX, &e.PosY, &ts, &att, &e.SoundVelocity, &e.SlantRange, &e.Status, &ch,
			&cx, &cy, &cd, &cat); err != nil {
			return nil, err
		}
		e.Timestamp, _ = time.Parse(time.RFC3339Nano, ts)
		_ = jsonUnmarshal(att, &e.Attitude)
		_ = jsonUnmarshal(ch, &e.Channels)
		if cx.Valid {
			e.CorrectedX = cx.Float64
		}
		if cy.Valid {
			e.CorrectedY = cy.Float64
		}
		if cd.Valid {
			e.CorrectedDepth = cd.Float64
		}
		if cat.Valid {
			t, _ := time.Parse(time.RFC3339Nano, cat.String)
			e.CorrectedAt = &t
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// UpdateEchoStatus validates and applies an echo status transition. A sealed
// batch rejects every echo mutation; the status transition is checked only
// after the batch is confirmed open.
func (s *Store) UpdateEchoStatus(id int64, to model.EchoStatus) error {
	if err := s.assertEchoBatchOpen(id); err != nil {
		return err
	}
	e, err := s.GetEcho(id)
	if err != nil {
		return err
	}
	if err := model.TransitionEcho(e.Status, to); err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE echo_windows SET status=? WHERE id=?`, string(to), id)
	return err
}

// SaveCorrectedGeometry persists the geometry correction output and moves the
// echo to "corrected". A sealed batch rejects the write (guarded by the
// status transition inside UpdateEchoStatus).
func (s *Store) SaveCorrectedGeometry(e *model.EchoWindow) error {
	if e.CorrectedAt == nil {
		return fmt.Errorf("%w: no correction applied", model.ErrInvalidAttitude)
	}
	if err := s.UpdateEchoStatus(e.ID, model.EchoCorrected); err != nil {
		return err
	}
	_, err := s.db.Exec(
		`UPDATE echo_windows SET corrected_x=?, corrected_y=?, corrected_depth=?, corrected_at=? WHERE id=?`,
		e.CorrectedX, e.CorrectedY, e.CorrectedDepth, e.CorrectedAt.UTC().Format(time.RFC3339Nano), e.ID)
	return err
}
