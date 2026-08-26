package store

import (
	"database/sql"

	"task250-biosonar/internal/model"
)

// CreateSubstrate registers a new substrate class.
func (s *Store) CreateSubstrate(sub *model.SubstrateType) (int64, error) {
	if err := sub.Validate(); err != nil {
		return 0, err
	}
	res, err := s.db.Exec(
		`INSERT INTO substrate_types(code, name, centroid, cov_diag) VALUES(?,?,?,?)`,
		sub.Code, sub.Name, jsonBytes(sub.Centroid), jsonBytes(sub.CovDiag))
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	sub.ID = id
	return id, nil
}

// GetSubstrate loads a substrate by ID.
func (s *Store) GetSubstrate(id int64) (*model.SubstrateType, error) {
	sub := &model.SubstrateType{ID: id}
	var centroid, cov string
	err := s.db.QueryRow(
		`SELECT code, name, centroid, cov_diag FROM substrate_types WHERE id=?`, id).
		Scan(&sub.Code, &sub.Name, &centroid, &cov)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := jsonUnmarshal(centroid, &sub.Centroid); err != nil {
		return nil, err
	}
	if err := jsonUnmarshal(cov, &sub.CovDiag); err != nil {
		return nil, err
	}
	return sub, nil
}

// ListSubstrates returns all substrate classes ordered by ID.
func (s *Store) ListSubstrates() ([]model.SubstrateType, error) {
	rows, err := s.db.Query(
		`SELECT id, code, name, centroid, cov_diag FROM substrate_types ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.SubstrateType
	for rows.Next() {
		var sub model.SubstrateType
		var centroid, cov string
		if err := rows.Scan(&sub.ID, &sub.Code, &sub.Name, &centroid, &cov); err != nil {
			return nil, err
		}
		_ = jsonUnmarshal(centroid, &sub.Centroid)
		_ = jsonUnmarshal(cov, &sub.CovDiag)
		out = append(out, sub)
	}
	return out, rows.Err()
}
