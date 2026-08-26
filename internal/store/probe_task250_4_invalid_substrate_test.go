package store_test

import (
	"math"
	"path/filepath"
	"testing"

	"task250-biosonar/internal/model"
	"task250-biosonar/internal/store"
)

func TestBug04_InvalidSubstrateModelRejected(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	sub := &model.SubstrateType{Code: "BAD", Name: "Bad", Centroid: []float64{0, 0, 0, 0}, CovDiag: []float64{1, math.NaN(), 1, 1}}
	if _, err := st.CreateSubstrate(sub); err == nil {
		t.Fatal("non-finite covariance was accepted")
	}
}
