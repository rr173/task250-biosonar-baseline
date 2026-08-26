// Package store implements SQLite-backed persistence for the biosonar
// seafloor-classification service. It owns the schema, migrations, the
// substrate seed catalogue and all entity CRUD.
package store

import (
	"database/sql"

	_ "modernc.org/sqlite"

	"task250-biosonar/internal/classify"
)

// Store wraps the application database handle.
type Store struct {
	db *sql.DB
}

// Open connects to (or creates) the SQLite database at path, applies the
// schema and seeds the default substrate catalogue when empty.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite is a single-writer engine.
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	if err := s.seedSubstrates(); err != nil {
		return nil, err
	}
	return s, nil
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

// DB exposes the underlying handle for advanced callers.
func (s *Store) DB() *sql.DB { return s.db }

// seedSubstrates inserts the built-in substrate catalogue the first time the
// table is empty, so classification works out of the box.
func (s *Store) seedSubstrates() error {
	var n int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM substrate_types").Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	for _, sub := range classify.DefaultSubstrates() {
		if _, err := s.db.Exec(
			"INSERT INTO substrate_types(code, name, centroid, cov_diag) VALUES(?,?,?,?)",
			sub.Code, sub.Name, jsonBytes(sub.Centroid), jsonBytes(sub.CovDiag)); err != nil {
			return err
		}
	}
	return nil
}
