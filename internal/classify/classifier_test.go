package classify

import (
	"math"
	"testing"

	"task250-biosonar/internal/model"
)

func TestClassifyMatchesCentroid(t *testing.T) {
	subs := DefaultSubstrates()
	for i := range subs {
		subs[i].ID = int64(i + 1)
	}
	fv := model.FeatureVector{-12, 2.0, 2.5, 0.05} // SAND centroid
	cls, err := Classify(1, fv, subs)
	if err != nil {
		t.Fatal(err)
	}
	if cls.PredictedCode != "SAND" {
		t.Fatalf("expected SAND, got %s", cls.PredictedCode)
	}
	if cls.Uncertainty > 0.2 {
		t.Fatalf("expected low uncertainty, got %f", cls.Uncertainty)
	}
	if cls.BestProbability() <= 0.5 {
		t.Fatalf("expected dominant probability > 0.5, got %f", cls.BestProbability())
	}
}

func TestClassifyMud(t *testing.T) {
	subs := DefaultSubstrates()
	for i := range subs {
		subs[i].ID = int64(i + 1)
	}
	fv := model.FeatureVector{-22, 1.5, 8.0, 0.02} // MUD centroid
	cls, err := Classify(2, fv, subs)
	if err != nil {
		t.Fatal(err)
	}
	if cls.PredictedCode != "MUD" {
		t.Fatalf("expected MUD, got %s", cls.PredictedCode)
	}
}

func TestClassifyEmpty(t *testing.T) {
	if _, err := Classify(1, model.FeatureVector{0, 0, 0, 0}, nil); err == nil {
		t.Fatal("expected error for empty substrates")
	}
}

func TestClassifyRejectsIllegalSubstrate(t *testing.T) {
	subs := DefaultSubstrates()
	for i := range subs {
		subs[i].ID = int64(i + 1)
	}
	fv := model.FeatureVector{-12, 2.0, 2.5, 0.05}

	// Each of these illegal catalogue entries must cause Classify to refuse
	// outright instead of emitting a plausible-looking but unusable posterior.
	ill := []struct {
		name string
		mut  func(s []model.SubstrateType) []model.SubstrateType
	}{
		{"zero covariance", func(s []model.SubstrateType) []model.SubstrateType { s[0].CovDiag[1] = 0; return s }},
		{"NaN covariance", func(s []model.SubstrateType) []model.SubstrateType { s[1].CovDiag[0] = math.NaN(); return s }},
		{"+Inf covariance", func(s []model.SubstrateType) []model.SubstrateType { s[2].CovDiag[3] = math.Inf(1); return s }},
		{"NaN centroid", func(s []model.SubstrateType) []model.SubstrateType { s[3].Centroid[0] = math.NaN(); return s }},
	}
	for _, c := range ill {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Classify(1, fv, c.mut(cloneSubs(subs))); err == nil {
				t.Fatalf("expected error for %s, got nil (posterior would be unusable)", c.name)
			}
		})
	}
}

func cloneSubs(s []model.SubstrateType) []model.SubstrateType {
	out := make([]model.SubstrateType, len(s))
	copy(out, s)
	for i := range out {
		out[i].Centroid = append([]float64(nil), s[i].Centroid...)
		out[i].CovDiag = append([]float64(nil), s[i].CovDiag...)
	}
	return out
}
