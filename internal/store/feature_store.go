package store

import (
	"database/sql"

	"task250-biosonar/internal/model"
)

// SaveFeature stores the extracted feature vector for an echo window,
// replacing any prior extraction. A sealed batch rejects the write.
func (s *Store) SaveFeature(echoID int64, fv model.FeatureVector) error {
	if err := s.assertEchoBatchOpen(echoID); err != nil {
		return err
	}
	_, err := s.db.Exec(
		`INSERT INTO echo_features(echo_id, feature) VALUES(?,?)
		 ON CONFLICT(echo_id) DO UPDATE SET feature=excluded.feature`,
		echoID, jsonBytes(fv))
	return err
}

// GetFeature loads the extracted feature vector for an echo window.
func (s *Store) GetFeature(echoID int64) (model.FeatureVector, error) {
	var raw string
	err := s.db.QueryRow(`SELECT feature FROM echo_features WHERE echo_id=?`, echoID).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotClassified
	}
	if err != nil {
		return nil, err
	}
	var fv model.FeatureVector
	if err := jsonUnmarshal(raw, &fv); err != nil {
		return nil, err
	}
	return fv, nil
}
