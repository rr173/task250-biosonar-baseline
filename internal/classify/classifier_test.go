package classify

import (
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
