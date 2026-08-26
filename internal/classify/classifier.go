package classify

import (
	"math"

	"task250-biosonar/internal/model"
)

// Result is one substrate's posterior for a single echo window.
type Result struct {
	SubstrateID   int64   `json:"substrate_id"`
	Probability   float64 `json:"probability"`
	Likelihood    float64 `json:"likelihood"`
	LogLikelihood float64 `json:"log_likelihood"`
}

// Classification is the full posterior over candidate substrates for one
// echo window.
type Classification struct {
	EchoID        int64    `json:"echo_id"`
	PredictedID   int64    `json:"predicted_id"`
	PredictedCode string   `json:"predicted_code"`
	Uncertainty   float64  `json:"uncertainty"` // 1 - max probability
	Results       []Result `json:"results"`
}

// gaussianLogLik computes the diagonal-covariance Gaussian log-likelihood of
// a feature vector under a substrate model.
func gaussianLogLik(fv model.FeatureVector, s model.SubstrateType) float64 {
	var ll float64
	for i := range fv {
		d := fv[i] - s.Centroid[i]
		v := s.CovDiag[i]
		ll += -0.5*math.Log(2*math.Pi*v) - 0.5*d*d/v
	}
	return ll
}

// Classify returns the posterior distribution of substrates for a feature
// vector. Probabilities are a softmax over the per-substrate log-likelihoods
// for numerical stability.
func Classify(echoID int64, fv model.FeatureVector, subs []model.SubstrateType) (*Classification, error) {
	if len(subs) == 0 {
		return nil, model.ErrUnknownSubstrate
	}
	if len(fv) != model.FeatureDim {
		return nil, model.ErrUnknownSubstrate
	}
	results := make([]Result, 0, len(subs))
	maxLL := math.Inf(-1)
	for _, s := range subs {
		if err := s.Validate(); err != nil {
			return nil, err
		}
		ll := gaussianLogLik(fv, s)
		results = append(results, Result{
			SubstrateID:   s.ID,
			LogLikelihood: ll,
			Likelihood:    math.Exp(ll),
		})
		if ll > maxLL {
			maxLL = ll
		}
	}
	var sum float64
	for i := range results {
		e := math.Exp(results[i].LogLikelihood - maxLL)
		results[i].Probability = e
		sum += e
	}
	for i := range results {
		results[i].Probability /= sum
		if math.IsNaN(results[i].Probability) || math.IsInf(results[i].Probability, 0) {
			return nil, model.ErrUnknownSubstrate
		}
	}
	best := 0
	for i := range results {
		if results[i].Probability > results[best].Probability {
			best = i
		}
	}
	cls := &Classification{
		EchoID:      echoID,
		PredictedID: results[best].SubstrateID,
		Uncertainty: 1 - results[best].Probability,
		Results:     results,
	}
	for _, s := range subs {
		if s.ID == cls.PredictedID {
			cls.PredictedCode = s.Code
		}
	}
	return cls, nil
}

// BestProbability returns the maximum posterior probability across substrates.
func (c *Classification) BestProbability() float64 {
	best := 0.0
	for _, r := range c.Results {
		if r.Probability > best {
			best = r.Probability
		}
	}
	return best
}
