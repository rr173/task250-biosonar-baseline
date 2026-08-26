package store

import (
	"fmt"
)

// migrate creates the immutable schema. Every entity carries a status column
// whose transitions are enforced by the model package, not by SQL.
func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS survey_batches (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			vessel TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			sealed_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS substrate_types (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			centroid TEXT NOT NULL,
			cov_diag TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS echo_windows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			batch_id INTEGER NOT NULL,
			ping_seq INTEGER NOT NULL,
			pos_x REAL NOT NULL,
			pos_y REAL NOT NULL,
			ts TEXT NOT NULL,
			attitude TEXT NOT NULL,
			sound_velocity REAL NOT NULL,
			slant_range REAL NOT NULL,
			status TEXT NOT NULL,
			channels TEXT NOT NULL,
			corrected_x REAL,
			corrected_y REAL,
			corrected_depth REAL,
			corrected_at TEXT,
			UNIQUE(batch_id, ping_seq)
		)`,
		`CREATE TABLE IF NOT EXISTS echo_features (
			echo_id INTEGER PRIMARY KEY,
			feature TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS classifications (
			echo_id INTEGER PRIMARY KEY,
			substrate_id INTEGER NOT NULL,
			code TEXT NOT NULL,
			probability REAL NOT NULL,
			uncertainty REAL NOT NULL,
			result_json TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS substrate_segments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			batch_id INTEGER NOT NULL,
			substrate_id INTEGER NOT NULL,
			start_seq INTEGER NOT NULL,
			end_seq INTEGER NOT NULL,
			status TEXT NOT NULL,
			confidence REAL NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS segment_boundaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			segment_id INTEGER NOT NULL,
			seq INTEGER NOT NULL,
			reason TEXT NOT NULL,
			from_substrate INTEGER NOT NULL,
			to_substrate INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			batch_id INTEGER NOT NULL,
			status TEXT NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS snapshot_segments (
			snapshot_id INTEGER NOT NULL,
			segment_id INTEGER NOT NULL,
			PRIMARY KEY (snapshot_id, segment_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_echo_batch ON echo_windows(batch_id)`,
		`CREATE INDEX IF NOT EXISTS idx_segment_batch ON substrate_segments(batch_id)`,
		`CREATE INDEX IF NOT EXISTS idx_snapshot_batch ON snapshots(batch_id)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}
