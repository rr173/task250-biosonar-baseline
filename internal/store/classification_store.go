package store

import (
	"database/sql"

	"task250-biosonar/internal/classify"
	"task250-biosonar/internal/model"
)

// SaveClassification upserts the posterior for one echo window.
func (s *Store) SaveClassification(c *classify.Classification) error {
	_, err := s.db.Exec(
		`INSERT INTO classifications(echo_id, substrate_id, code, probability, uncertainty, result_json)
		 VALUES(?,?,?,?,?,?)
		 ON CONFLICT(echo_id) DO UPDATE SET
		   substrate_id=excluded.substrate_id, code=excluded.code,
		   probability=excluded.probability, uncertainty=excluded.uncertainty,
		   result_json=excluded.result_json`,
		c.EchoID, c.PredictedID, c.PredictedCode, c.BestProbability(), c.Uncertainty, jsonBytes(c.Results))
	return err
}

// GetClassification loads the stored posterior for one echo window.
func (s *Store) GetClassification(echoID int64) (*classify.Classification, error) {
	var substrateID int64
	var code string
	var prob, unc float64
	var raw string
	err := s.db.QueryRow(
		`SELECT substrate_id, code, probability, uncertainty, result_json FROM classifications WHERE echo_id=?`, echoID).
		Scan(&substrateID, &code, &prob, &unc, &raw)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotClassified
	}
	if err != nil {
		return nil, err
	}
	var results []classify.Result
	_ = jsonUnmarshal(raw, &results)
	return &classify.Classification{
		EchoID:        echoID,
		PredictedID:   substrateID,
		PredictedCode: code,
		Uncertainty:   unc,
		Results:       results,
	}, nil
}

// ClassifiedPingRow is a per-ping classification enriched with its ping
// sequence, used to order and merge a survey line.
type ClassifiedPingRow struct {
	EchoID        int64
	PingSeq       int
	SubstrateID   int64
	PredictedCode string
	Probability   float64
	Uncertainty   float64
}

// ClassifiedPings returns all posteriors for a batch ordered by ping sequence,
// ready for segment merging.
func (s *Store) ClassifiedPings(batchID int64) ([]ClassifiedPingRow, error) {
	rows, err := s.db.Query(
		`SELECT c.echo_id, e.ping_seq, c.substrate_id, c.code, c.probability, c.uncertainty
		 FROM classifications c JOIN echo_windows e ON e.id=c.echo_id
		 WHERE e.batch_id=? ORDER BY e.ping_seq`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ClassifiedPingRow
	for rows.Next() {
		var r ClassifiedPingRow
		if err := rows.Scan(&r.EchoID, &r.PingSeq, &r.SubstrateID, &r.PredictedCode, &r.Probability, &r.Uncertainty); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
